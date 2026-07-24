package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v5"
	"github.com/spf13/cobra"

	"github.com/calgorr/sms-gateway/balance"
	"github.com/calgorr/sms-gateway/config"
	"github.com/calgorr/sms-gateway/internal/api"
	"github.com/calgorr/sms-gateway/ledger"
	"github.com/calgorr/sms-gateway/message"
	"github.com/calgorr/sms-gateway/pkg/cassandra"
	kafka "github.com/calgorr/sms-gateway/pkg/kafkaclient"
	"github.com/calgorr/sms-gateway/pkg/postgres"
	redis "github.com/calgorr/sms-gateway/pkg/redisclient"
	"github.com/calgorr/sms-gateway/queue"
)

var serveApiCMD = &cobra.Command{
	Use:   "serve-api",
	Short: "Serve the API service",
	Run:   serveApi,
}

func serveApi(_ *cobra.Command, _ []string) {
	// Cassandra
	session, err := cassandra.NewSession(config.C.Cassandra)
	if err != nil {
		log.Fatalf("create cassandra session: %v", err)
	}
	defer session.Close()

	messageRepository := message.NewRepository(session)

	// Postgres
	postgresClient, err := postgres.NewDB(config.C.Postgres)
	if err != nil {
		log.Fatalf("create postgres client: %v", err)
	}
	defer postgresClient.Close()

	ledgerRepository := ledger.NewRepository(postgresClient)

	// Kafka producer
	kafkaProducer, err := kafka.NewProducer(config.C.Kafka)
	if err != nil {
		log.Fatalf("create kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	producer := queue.NewKafkaProducer(kafkaProducer)

	kafkaReader, err := kafka.NewConsumer(config.C.Kafka, queue.InsertSmsTopic)
	if err != nil {
		log.Fatalf("create kafka consumer: %v", err)
	}
	defer kafkaReader.Close()

	consumer := queue.NewConsumer(
		kafkaReader,
		ledgerRepository,
		messageRepository,
		config.C.WorkerPool.NumWorkers,
	)

	// Redis
	redisClient, err := redis.NewClient(config.C.Redis)
	if err != nil {
		log.Fatalf("create redis client: %v", err)
	}
	defer redisClient.Close()

	checkBalanceService := balance.NewRedisStore(redisClient)

	e := echo.New()

	server := api.NewHttpServer(
		e,
		ledgerRepository,
		producer,
		checkBalanceService,
	)

	server.DefineRoutes()

	addr := fmt.Sprintf("%s:%d", config.C.HttpServer.Host, config.C.HttpServer.Port)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// HTTP server
	go func() {
		log.Printf("API server listening on %s", addr)

		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	// Kafka consumer
	go func() {
		log.Println("Kafka consumer started")

		if err := consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Fatalf("consumer stopped: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("shutting down...")
}

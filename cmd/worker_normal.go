package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/calgorr/sms-gateway/config"
	"github.com/calgorr/sms-gateway/dispatcher"
	"github.com/calgorr/sms-gateway/message"
	"github.com/calgorr/sms-gateway/operator"
	"github.com/calgorr/sms-gateway/pkg/cassandra"
	kafka "github.com/calgorr/sms-gateway/pkg/kafkaclient"
	"github.com/calgorr/sms-gateway/queue"
)

var workerNormalCMD = &cobra.Command{
	Use:   "worker-normal",
	Short: "Run the normal worker",
	Run:   runWorkerNormal,
}

func runWorkerNormal(_ *cobra.Command, _ []string) {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	session, err := cassandra.NewSession(config.C.Cassandra)
	if err != nil {
		log.Fatalf("create cassandra session: %v", err)
	}
	defer session.Close()

	messageRepository := message.NewRepository(session)

	operatorClient := operator.NewOperator(config.C.Operator)

	reader, err := kafka.NewConsumer(config.C.Kafka, queue.SmsNormalTopic)
	if err != nil {
		log.Fatalf("create kafka reader: %v", err)
	}
	defer reader.Close()

	worker := dispatcher.NewWorker(
		"normal-worker",
		reader,
		operatorClient,
		messageRepository,
		config.C.WorkerPool.NumWorkers,
	)

	log.Println("starting normal worker")

	if err := worker.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("normal worker failed: %v", err)
	}

	log.Println("normal worker stopped")
}

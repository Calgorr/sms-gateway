package api

import (
	"github.com/labstack/echo/v5"

	"github.com/calgorr/sms-gateway/balance"
	"github.com/calgorr/sms-gateway/internal/api/middleware"
	"github.com/calgorr/sms-gateway/ledger"
	"github.com/calgorr/sms-gateway/message"
	"github.com/calgorr/sms-gateway/queue"
)

type HttpServer struct {
	e                 *echo.Echo
	ledgerRepository  ledger.LedgerRepository
	messageRepository message.MessageRepository
	producer          queue.KafkaProducer
	balanceService    balance.CheckBalanceService
}

func NewHttpServer(
	e *echo.Echo,
	ledgerRepository ledger.LedgerRepository,
	producer queue.KafkaProducer,
	balanceService balance.CheckBalanceService,
) *HttpServer {
	return &HttpServer{
		e:                e,
		ledgerRepository: ledgerRepository,
		producer:         producer,
		balanceService:   balanceService,
	}
}

func (s *HttpServer) Start() error {
	return s.e.Start(":8080")
}

func (s *HttpServer) DefineRoutes() {
	s.e.Use(middleware.Logging())

	s.e.POST("/sms", s.SendSMS())
	s.e.POST("/balance/topup", s.TopUpBalance())
	s.e.GET("/reports", s.GetReports())
}

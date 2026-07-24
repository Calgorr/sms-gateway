package api

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/labstack/echo/v5"

	"github.com/calgorr/sms-gateway/balance"
	"github.com/calgorr/sms-gateway/config"
	"github.com/calgorr/sms-gateway/internal/api/dto"
	"github.com/calgorr/sms-gateway/queue"
	"github.com/calgorr/sms-gateway/sms"
)

func (s *HttpServer) SendSMS() echo.HandlerFunc {
	return func(c *echo.Context) error {
		var req dto.SendSMSRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(400, map[string]string{"error": "invalid request body"})
		}

		// Validate the request
		if err := req.Validate(); err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}

		// Check balance using the balance service
		_, err := s.balanceService.CheckAndDeduct(c.Request().Context(), strconv.Itoa(int(req.CustomerID)))
		if err != nil {
			// If cache miss, get balance from ledger
			if errors.Is(err, balance.ErrCacheMiss) {
				ledgerBalance, err := s.ledgerRepository.SumByCustomer(c.Request().Context(), req.CustomerID)
				if err != nil {
					log.Printf("send-sms: rebuild balance for customer %d: %v", req.CustomerID, err)
					return c.JSON(500, map[string]string{"error": "failed to get balance from ledger"})
				}
				if ledgerBalance <= config.C.Opts.CostPerSms {
					return c.JSON(402, map[string]string{"error": "insufficient balance"})
				}
				ledgerBalance -= config.C.Opts.CostPerSms
				err = s.balanceService.SetBalance(c.Request().Context(), strconv.Itoa(int(req.CustomerID)), ledgerBalance)
				if err != nil {
					log.Printf("send-sms: set rebuilt balance for customer %d: %v", req.CustomerID, err)
					return c.JSON(500, map[string]string{"error": "internal server error"})
				}
			} else if errors.Is(err, balance.ErrInsufficientBalance) {
				return c.JSON(402, map[string]string{"error": "insufficient balance"})
			} else {
				log.Printf("send-sms: check balance for customer %d: %v", req.CustomerID, err)
				return c.JSON(500, map[string]string{"error": "failed to check balance"})
			}
		}

		// Generate message ID
		messageID := gocql.TimeUUID().String()

		// Determine topic based on priority
		topic := queue.SmsNormalTopic
		if req.Priority == dto.PriorityExpress {
			topic = queue.SmsExpressTopic
		}

		event := sms.SMSEvent{
			MessageID:  messageID,
			CustomerID: req.CustomerID,
			To:         req.To,
			Text:       req.Text,
			Priority:   req.Priority,
			CreatedAt:  time.Now(),
		}

		// Produce the event to Kafka to insert debit in ledger first (to ensure balance is deducted before processing the SMS)
		if err := s.producer.SendSMS(c.Request().Context(), queue.InsertSmsTopic, event); err != nil {
			log.Printf("send-sms: publish to %s (message %s): %v", queue.InsertSmsTopic, messageID, err)
			return c.JSON(500, map[string]string{"error": "internal server error"})
		}

		// Produce the event to Kafka
		if err := s.producer.SendSMS(c.Request().Context(), topic, event); err != nil {
			log.Printf("send-sms: publish to %s (message %s): %v", topic, messageID, err)
			return c.JSON(500, map[string]string{"error": "internal server error"})
		}

		// Return success response
		return c.JSON(202, map[string]string{
			"message_id": messageID,
			"status":     "accepted",
		})
	}
}

package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/calgorr/sms-gateway/internal/api/dto"
)

func (s *HttpServer) GetReports() echo.HandlerFunc {
	return func(c *echo.Context) error {
		customerIDStr := c.QueryParam("customer_id")
		fromStr := c.QueryParam("from")
		toStr := c.QueryParam("to")

		customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid customer_id",
			})
		}

		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid from format, expected RFC3339",
			})
		}

		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid to format, expected RFC3339",
			})
		}

		report, err := s.messageRepository.Report(
			c.Request().Context(),
			customerID,
			from,
			to,
		)
		if err != nil {
			log.Printf("reports: query for customer %d: %v", customerID, err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		entries := make([]dto.ReportEntry, 0, len(report.Messages))

		for _, msg := range report.Messages {
			entries = append(entries, dto.ReportEntry{
				MessageID: msg.ID.String(),
				To:        msg.To,
				Text:      msg.Text,
				Priority:  string(msg.Priority),
				Status:    string(msg.Status),
				CreatedAt: msg.CreatedAt,
				SentAt:    &msg.SentAt,
			})
		}

		response := dto.ReportResponse{
			CustomerID: strconv.FormatInt(report.CustomerID, 10),
			From:       report.From,
			To:         report.To,

			TotalMessages: report.TotalMessages,
			SuccessCount:  report.SuccessCount,
			FailedCount:   report.FailedCount,
			PendingCount:  report.PendingCount,

			Entries: entries,
		}

		return c.JSON(http.StatusOK, response)
	}
}

package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/calgorr/sms-gateway/internal/api/dto"
)

func (s *HttpServer) TopUpBalance() echo.HandlerFunc {
	return func(c *echo.Context) error {
		var req dto.TopUpRequest

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
		}

		if err := req.Validate(); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		ctx := c.Request().Context()

		if err := s.ledgerRepository.InsertTopup(
			ctx,
			req.CustomerID,
			req.Amount,
		); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to insert topup ledger",
			})
		}

		if err := s.balanceService.AddOrSetBalance(
			ctx,
			strconv.FormatInt(req.CustomerID, 10),
			req.Amount,
		); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to update balance",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"customer_id": req.CustomerID,
			"amount":      req.Amount,
		})
	}
}

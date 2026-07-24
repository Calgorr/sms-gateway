package middleware

import (
	"log"
	"time"

	"github.com/labstack/echo/v5"
)

// Logging returns a middleware that logs one line per request with the method,
// path, response status and how long the handler took.
func Logging() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()

			err := next(c)

			status := 0
			if res, ok := c.Response().(*echo.Response); ok {
				status = res.Status
			}

			log.Printf("%s %s -> %d (%s)",
				c.Request().Method,
				c.Request().URL.Path,
				status,
				time.Since(start),
			)

			return err
		}
	}
}

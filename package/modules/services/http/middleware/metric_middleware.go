package middleware

import "github.com/gofiber/fiber/v2"

type MetricOptions struct {
	// ServiceName is the name of the service being traced
	ServiceName string
	// ServiceVersion is the version of the service
	ServiceVersion string
	// Skip defines a function to skip the middleware
	Skip func(c *fiber.Ctx) bool
	// Whitelist path to skip the middleware
	Whitelist map[string]bool
	// Metrics is the Prometheus registry
}

func MetricMiddleware() fiber.Handler {

	return func(c *fiber.Ctx) error {
		return c.Next()

	}
}

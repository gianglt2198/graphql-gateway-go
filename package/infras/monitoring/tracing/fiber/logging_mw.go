package tcfiber

import (
	fiber "github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/utils"
)

const (
	REQUEST_ID = "requestId"
)

func RequestIDMiddleware(c *fiber.Ctx) error {
	ctx, requestID := utils.ApplyRequestIDWithContext(c.UserContext())
	c.Locals(REQUEST_ID, requestID)
	c.SetUserContext(ctx)
	return c.Next()
}

func LoggerMiddleware(logger *logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logFields := []zapcore.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("request_id", c.Locals("requestId").(string)),
		}
		if err := c.Next(); err != nil {
			logger.GetLogger().With(logFields...).Error("Request failed", zap.Error(err))
			return err
		}

		logger.GetLogger().With(logFields...).Info("Request succeeded")
		return nil
	}
}

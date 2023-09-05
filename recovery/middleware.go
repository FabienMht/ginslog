package recovery

import (
	"context"
	"log/slog"
	"net/http/httputil"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// New returns a gin.HandlerFunc (middleware) that recovers from any
// panics and logs the panic using slog. It sets the HTTP status code to
// 500. By default, the log level is ERROR.
func New(logger *slog.Logger, opts ...ConfigOption) gin.HandlerFunc {
	config := newConfig()
	for _, opt := range opts {
		opt(config)
	}
	config.validate()

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var httpRequest []byte

				if config.isDefaultFields() {
					httpRequest, _ = httputil.DumpRequest(c.Request, false) //nolint: errcheck
				}

				attributes := []slog.Attr{}

				// Add the error
				if config.errorField {
					attributes = append(attributes, slog.Any("error", err))
				}

				// Add the request
				if config.requestField {
					attributes = append(attributes, slog.String("request", string(httpRequest)))
				}

				// Add the stack trace
				if config.stackField {
					attributes = append(attributes, slog.String("stack", string(debug.Stack())))
				}

				// Add custom fields
				if config.customFields != nil {
					attributes = append(attributes, config.customFields(c)...)
				}

				// Log the panic
				logger.LogAttrs(context.Background(), config.defaultLevel, "Panic recovered", attributes...)

				// Call the custom recovery
				config.customRecovery(c, err)
			}
		}()
		c.Next()
	}
}

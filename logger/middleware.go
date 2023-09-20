package logger

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// New returns a gin.HandlerFunc (middleware) that logs requests using slog.
//
// By default, the log level depend on the HTTP return code:
//   - 1XX return codes are logged at INFO level.
//   - 2XX return codes are logged at INFO level.
//   - 3XX return codes are logged at INFO level.
//   - 4XX return codes are logged at WARN level.
//   - 5XX return codes are logged at ERROR level.
//
// By default, all paths are logged. Whitelist and Blacklist
// can be used to filter the paths.
//
// By default, the following fields are logged:
//   - IP address
//   - Status code
//   - HTTP method
//   - Path
//   - User agent
//   - Latency
//   - Request ID (X-Request-ID header)
func New(logger *slog.Logger, opts ...ConfigOption) gin.HandlerFunc {
	config := newConfig()
	for _, opt := range opts {
		opt(config)
	}
	config.validate()

	return func(c *gin.Context) {
		start := time.Now()
		requestID := uuid.New().String()

		if config.requestIDField {
			c.Header("X-Request-ID", requestID)
		}

		// Process the request
		c.Next()

		// Check if the path is whitelisted
		if len(config.whitelistPaths) > 0 {
			for _, v := range config.whitelistPaths {
				if !v.MatchString(c.Request.URL.Path) {
					return
				}
			}
		}

		// Check if the path is blacklisted
		if len(config.blacklistPaths) > 0 {
			for _, v := range config.blacklistPaths {
				if v.MatchString(c.Request.URL.Path) {
					return
				}
			}
		}

		attributes := []slog.Attr{}

		// Add the IP address
		if config.ipField {
			attributes = append(attributes, slog.String("ip", c.ClientIP()))
		}

		// Add the status code
		if config.statusField {
			attributes = append(attributes, slog.Int("status", c.Writer.Status()))
		}

		// Add the HTTP method
		if config.methodField {
			attributes = append(attributes, slog.String("method", c.Request.Method))
		}

		// Add the path
		if config.pathField {
			attributes = append(attributes, slog.String("path", c.Request.URL.Path))
		}

		// Add the user agent
		if config.userAgentField {
			attributes = append(attributes, slog.String("user-agent", c.Request.UserAgent()))
		}

		// Add the latency
		if config.latencyField {
			attributes = append(attributes, slog.Duration("latency", time.Since(start)))
		}

		// Add the request ID
		if config.requestIDField {
			attributes = append(attributes, slog.String("request-id", requestID))
		}

		// Add custom fields
		if config.customFields != nil {
			attributes = append(attributes, config.customFields(c)...)
		}

		// Log according to the status code
		for _, httpLevel := range config.httpLevels {
			if httpLevel.match(c.Writer.Status()) {
				logger.LogAttrs(context.Background(), httpLevel.level, "Incoming request", attributes...)

				// Call the custom logger
				config.customLogger(c, logger)

				return
			}
		}

		// If no status code matched, use the default level
		logger.LogAttrs(context.Background(), config.defaultLevel, "Incoming request", attributes...)

		// Call the custom logger
		config.customLogger(c, logger)
	}
}

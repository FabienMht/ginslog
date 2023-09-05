package recovery

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CustomFields allows to add custom fields to the log line.
type CustomFields func(c *gin.Context) []slog.Attr

// Config represents the recovery middleware configuration.
type Config struct {
	// Default log level.
	defaultLevel slog.Level

	// Custom recevery function.
	customRecovery gin.RecoveryFunc

	// Custom function to add custom fields to the log line.
	customFields CustomFields

	// Default fields to log.
	// Error from the panic.
	errorField bool
	// HTTP request.
	requestField bool
	// Stack trace.
	stackField bool
}

// newConfig returns a new Config.
func newConfig() *Config {
	return &Config{
		defaultLevel: slog.LevelError,
		customRecovery: func(c *gin.Context, err interface{}) {
			c.AbortWithStatus(http.StatusInternalServerError)
		},
		customFields: nil,
		errorField:   true,
		requestField: true,
		stackField:   true,
	}
}

// isDefaultFields checks if any default fields are used.
func (c *Config) isDefaultFields() bool {
	return c.errorField ||
		c.requestField ||
		c.stackField
}

// validate validates the Config.
func (c *Config) validate() {
	if !c.isDefaultFields() && c.customFields == nil {
		panic("no fields to log")
	}
}

// ConfigOption allows to customize the middleware config.
type ConfigOption func(*Config)

// WithDefaultLevel allows to set the default log level.
func WithDefaultLevel(level slog.Level) ConfigOption {
	return func(c *Config) {
		c.defaultLevel = level
	}
}

// WithCustomRecovery allows to set a custom recovery function.
func WithCustomRecovery(customRecovery gin.RecoveryFunc) ConfigOption {
	return func(c *Config) {
		c.customRecovery = customRecovery
	}
}

// WithCustomFields allows to set a custom function to add custom fields to the log line.
func WithCustomFields(customFields CustomFields) ConfigOption {
	return func(c *Config) {
		c.customFields = customFields
	}
}

// WithoutDefaultFields to not use the default fields in the log line.
func WithoutDefaultFields() ConfigOption {
	return func(c *Config) {
		c.errorField = false
		c.requestField = false
		c.stackField = false
	}
}

// WithoutError to not log the error field.
func WithoutError() ConfigOption {
	return func(c *Config) {
		c.errorField = false
	}
}

// WithoutRequest to not log the request field.
func WithoutRequest() ConfigOption {
	return func(c *Config) {
		c.requestField = false
	}
}

// WithoutStack to not log the stack field.
func WithoutStack() ConfigOption {
	return func(c *Config) {
		c.stackField = false
	}
}

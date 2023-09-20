package logger

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/gin-gonic/gin"
)

const (
	// HTTP response codes groups regex defined in the RFC 9110.
	HTTPInformationalRegex = "^1[0-9]{2}$"
	HTTPSuccessfulRegex    = "^2[0-9]{2}$"
	HTTPRedirectionRegex   = "^3[0-9]{2}$"
	HTTPClientErrorRegex   = "^4[0-9]{2}$"
	HTTPServerErrorRegex   = "^5[0-9]{2}$"
)

// CustomFields allows to add custom fields to the log line.
type CustomFields func(c *gin.Context) []slog.Attr

// CustomLogger allows to call a custom logger function.
type CustomLogger func(c *gin.Context, logger *slog.Logger)

// httpLevel associates a log level to an HTTP return code regex.
type httpLevel struct {
	// Log level to use.
	level slog.Level
	// Compiled regex.
	regexp *regexp.Regexp
}

// NewhttpLevel returns a new httpLevel.
func newhttpLevel(expr string, level slog.Level) *httpLevel {
	return &httpLevel{
		level:  level,
		regexp: regexp.MustCompile(expr),
	}
}

// Match returns true if the HTTP return code matches the regex.
func (h *httpLevel) match(code int) bool {
	return h.regexp.MatchString(fmt.Sprintf("%d", code))
}

// Config represents the logging middleware configuration.
type Config struct {
	// Default log level.
	defaultLevel slog.Level

	// Log level based on the HTTP return code.
	// Regex can be used to match multiple codes.
	httpLevels []*httpLevel

	// Whitelist or blacklist paths.
	// By default, all paths are logged.
	// If a whitelist is set, only whitelisted paths are logged.
	// If a blacklist is set, all paths except blacklisted are logged.
	whitelistPaths []*regexp.Regexp
	blacklistPaths []*regexp.Regexp

	// Custom logger function.
	customLogger CustomLogger

	// Custom function to add custom fields to the log line.
	customFields CustomFields

	// Default fields to log.
	// Client IP address.
	ipField bool
	// HTTP return code.
	statusField bool
	// HTTP method.
	methodField bool
	// HTTP path.
	pathField bool
	// User agent.
	userAgentField bool
	// Request latency.
	latencyField bool
	// UUID generated X-Request-ID header.
	requestIDField bool
}

// newConfig returns a new Config.
func newConfig() *Config {
	return &Config{
		defaultLevel: slog.LevelInfo,
		httpLevels: []*httpLevel{
			newhttpLevel(HTTPInformationalRegex, slog.LevelInfo),
			newhttpLevel(HTTPSuccessfulRegex, slog.LevelInfo),
			newhttpLevel(HTTPRedirectionRegex, slog.LevelInfo),
			newhttpLevel(HTTPClientErrorRegex, slog.LevelWarn),
			newhttpLevel(HTTPServerErrorRegex, slog.LevelError),
		},
		whitelistPaths: []*regexp.Regexp{},
		blacklistPaths: []*regexp.Regexp{},
		customFields:   nil,
		ipField:        true,
		statusField:    true,
		methodField:    true,
		pathField:      true,
		userAgentField: true,
		latencyField:   true,
		requestIDField: true,
	}
}

// isDefaultFields checks if any default fields are used.
func (c *Config) isDefaultFields() bool {
	return c.ipField ||
		c.statusField ||
		c.methodField ||
		c.pathField ||
		c.userAgentField ||
		c.latencyField ||
		c.requestIDField
}

// validate validates the Config.
func (c *Config) validate() {
	if len(c.whitelistPaths) != 0 && len(c.blacklistPaths) != 0 {
		panic("whitelist and blacklist can't be used together")
	}
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

// WithHTTPLevels allows to set the log level based on the HTTP return code.
// The map key is a regex to match the HTTP return code. It panics if the regex is invalid.
func WithHTTPLevels(httpLevels map[string]slog.Level) ConfigOption {
	return func(c *Config) {
		c.httpLevels = []*httpLevel{}
		for k, v := range httpLevels {
			c.httpLevels = append(c.httpLevels, newhttpLevel(k, v))
		}
	}
}

// WithWhitelistPath allows to whitelist paths. It panics if the regex is invalid.
// If a whitelist is set, only whitelisted paths are logged.
func WithWhitelistPath(whitelistPath []string) ConfigOption {
	return func(c *Config) {
		for _, v := range whitelistPath {
			c.whitelistPaths = append(c.whitelistPaths, regexp.MustCompile(v))
		}
	}
}

// WithBlacklistPath allows to blacklist paths. It panics if the regex is invalid.
// If a blacklist is set, all paths except blacklisted are logged.
func WithBlacklistPath(blacklistPath []string) ConfigOption {
	return func(c *Config) {
		for _, v := range blacklistPath {
			c.blacklistPaths = append(c.blacklistPaths, regexp.MustCompile(v))
		}
	}
}

// WithCustomLogger allows to set a custom logger function.
func WithCustomLogger(customLogger CustomLogger) ConfigOption {
	return func(c *Config) {
		c.customLogger = customLogger
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
		c.ipField = false
		c.statusField = false
		c.methodField = false
		c.pathField = false
		c.userAgentField = false
		c.latencyField = false
		c.requestIDField = false
	}
}

// WithoutIP to not add the IP address to the log line.
func WithoutIP() ConfigOption {
	return func(c *Config) {
		c.ipField = false
	}
}

// WithoutStatus to not add the HTTP return code to the log line.
func WithoutStatus() ConfigOption {
	return func(c *Config) {
		c.statusField = false
	}
}

// WithoutMethod to not add the HTTP method to the log line.
func WithoutMethod() ConfigOption {
	return func(c *Config) {
		c.methodField = false
	}
}

// WithoutPath to not add the HTTP path to the log line.
func WithoutPath() ConfigOption {
	return func(c *Config) {
		c.pathField = false
	}
}

// WithoutUserAgent to not add the user agent to the log line.
func WithoutUserAgent() ConfigOption {
	return func(c *Config) {
		c.userAgentField = false
	}
}

// WithoutLatency to not add the request latency to the log line.
func WithoutLatency() ConfigOption {
	return func(c *Config) {
		c.latencyField = false
	}
}

// WithoutRequestID to not add the X-Request-ID header to the log line.
func WithoutRequestID() ConfigOption {
	return func(c *Config) {
		c.requestIDField = false
	}
}

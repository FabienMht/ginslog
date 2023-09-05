package logger

import (
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FabienMht/ginslog/slogtest"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var skipFields = []string{"ip", "latency"}

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		opts       []ConfigOption
		code       int
		wantFields []slog.Attr
		wantLevel  slog.Level
		wantPanic  bool
	}{
		{
			name: "default options",
			opts: []ConfigOption{},
			code: 200,
			wantFields: []slog.Attr{
				slog.String("ip", ""),
				slog.Int("status", 200),
				slog.String("method", "GET"),
				slog.String("path", "/test"),
				slog.String("user-agent", "test"),
				slog.String("latency", ""),
				slog.String("request-id", "52fdfc07-2182-454f-963f-5f0f9a621d72"),
			},
			wantLevel: slog.LevelInfo,
		},
		{
			name: "set HTTP level",
			opts: []ConfigOption{
				WithHTTPLevels(map[string]slog.Level{HTTPSuccessfulRegex: slog.LevelWarn}),
			},
			code: 200,
			wantFields: []slog.Attr{
				slog.String("ip", ""),
				slog.Int("status", 200),
				slog.String("method", "GET"),
				slog.String("path", "/test"),
				slog.String("user-agent", "test"),
				slog.String("latency", ""),
				slog.String("request-id", "52fdfc07-2182-454f-963f-5f0f9a621d72"),
			},
			wantLevel: slog.LevelWarn,
		},
		{
			name: "without IP",
			opts: []ConfigOption{WithoutIP()},
			code: 200,
			wantFields: []slog.Attr{
				slog.Int("status", 200),
				slog.String("method", "GET"),
				slog.String("path", "/test"),
				slog.String("user-agent", "test"),
				slog.String("latency", ""),
				slog.String("request-id", "52fdfc07-2182-454f-963f-5f0f9a621d72"),
			},
			wantLevel: slog.LevelInfo,
		},
		{
			name: "without method",
			opts: []ConfigOption{WithoutMethod()},
			code: 200,
			wantFields: []slog.Attr{
				slog.String("ip", ""),
				slog.Int("status", 200),
				slog.String("path", "/test"),
				slog.String("user-agent", "test"),
				slog.String("latency", ""),
				slog.String("request-id", "52fdfc07-2182-454f-963f-5f0f9a621d72"),
			},
			wantLevel: slog.LevelInfo,
		},
		{
			name: "without path",
			opts: []ConfigOption{WithoutPath()},
			code: 200,
			wantFields: []slog.Attr{
				slog.String("ip", ""),
				slog.Int("status", 200),
				slog.String("method", "GET"),
				slog.String("user-agent", "test"),
				slog.String("latency", ""),
				slog.String("request-id", "52fdfc07-2182-454f-963f-5f0f9a621d72"),
			},
			wantLevel: slog.LevelInfo,
		},
		{
			name: "without user agent",
			opts: []ConfigOption{WithoutUserAgent()},
			code: 200,
			wantFields: []slog.Attr{
				slog.String("ip", ""),
				slog.Int("status", 200),
				slog.String("method", "GET"),
				slog.String("path", "/test"),
				slog.String("latency", ""),
				slog.String("request-id", "52fdfc07-2182-454f-963f-5f0f9a621d72"),
			},
			wantLevel: slog.LevelInfo,
		},
		{
			name: "without latency",
			opts: []ConfigOption{WithoutLatency()},
			code: 200,
			wantFields: []slog.Attr{
				slog.String("ip", ""),
				slog.Int("status", 200),
				slog.String("method", "GET"),
				slog.String("path", "/test"),
				slog.String("user-agent", "test"),
				slog.String("request-id", "52fdfc07-2182-454f-963f-5f0f9a621d72"),
			},
			wantLevel: slog.LevelInfo,
		},
		{
			name: "without request ID",
			opts: []ConfigOption{WithoutRequestID()},
			code: 200,
			wantFields: []slog.Attr{
				slog.String("ip", ""),
				slog.Int("status", 200),
				slog.String("method", "GET"),
				slog.String("path", "/test"),
				slog.String("user-agent", "test"),
				slog.String("latency", ""),
			},
			wantLevel: slog.LevelInfo,
		},
		{
			name:       "without default fields",
			opts:       []ConfigOption{WithoutDefaultFields()},
			wantFields: []slog.Attr{},
			wantLevel:  slog.LevelInfo,
			wantPanic:  true,
		},
		{
			name: "custom with default fields",
			opts: []ConfigOption{
				WithCustomFields(
					func(c *gin.Context) []slog.Attr {
						return []slog.Attr{slog.String("content-type", c.GetHeader("Content-Type"))}
					},
				),
			},
			code: 200,
			wantFields: []slog.Attr{
				slog.String("ip", ""),
				slog.Int("status", 200),
				slog.String("method", "GET"),
				slog.String("path", "/test"),
				slog.String("user-agent", "test"),
				slog.String("latency", ""),
				slog.String("request-id", "52fdfc07-2182-454f-963f-5f0f9a621d72"),
				slog.String("content-type", "test"),
			},
			wantLevel: slog.LevelInfo,
		},
		{
			name: "custom without default fields",
			opts: []ConfigOption{
				WithoutDefaultFields(),
				WithCustomFields(
					func(c *gin.Context) []slog.Attr {
						return []slog.Attr{slog.String("content-type", c.GetHeader("Content-Type"))}
					},
				),
			},
			code: 200,
			wantFields: []slog.Attr{
				slog.String("content-type", "test"),
			},
			wantLevel: slog.LevelInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set a fixed random seed to get a fixed request ID
			uuid.SetRand(rand.New(rand.NewSource(1)))

			// Create a new logger with a mock handler
			logger := slog.New(slogtest.NewMockHandler(
				slog.NewTextHandler(os.Stderr, nil),
				t,
				tt.wantLevel,
				tt.wantFields,
				skipFields,
			))

			gin.SetMode(gin.TestMode)
			router := gin.New()
			if tt.wantPanic {
				require.Panics(t, func() { router.Use(New(logger, tt.opts...)) })
				return
			}
			router.Use(New(logger, tt.opts...))

			// Define routes
			router.GET("/test", func(c *gin.Context) {
				c.JSON(tt.code, nil)
			})

			// Create a new request
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/test", nil)
			req.Header.Set("User-Agent", "test")
			req.Header.Set("Content-Type", "test")
			require.NoError(t, err)
			router.ServeHTTP(resp, req)
		})
	}
}

func TestNewWhiteListBlackList(t *testing.T) {
	tests := []struct {
		name       string
		opts       []ConfigOption
		code       int
		wantFields []slog.Attr
		wantLevel  slog.Level
		wantPanic  bool
	}{
		{
			name: "whitelist test1",
			opts: []ConfigOption{WithBlacklistPath([]string{"/test1"})},
			code: 200,
			wantFields: []slog.Attr{
				slog.String("ip", ""),
				slog.Int("status", 200),
				slog.String("method", "GET"),
				slog.String("path", "/test1"),
				slog.String("user-agent", "test1"),
				slog.String("latency", ""),
				slog.String("request-id", "9566c74d-1003-4c4d-bbbb-0407d1e2c649"),
			},
			wantLevel: slog.LevelInfo,
		},
		{
			name: "blacklist test1",
			opts: []ConfigOption{WithBlacklistPath([]string{"/test1"})},
			code: 200,
			wantFields: []slog.Attr{
				slog.String("ip", ""),
				slog.Int("status", 200),
				slog.String("method", "GET"),
				slog.String("path", "/test2"),
				slog.String("user-agent", "test2"),
				slog.String("latency", ""),
				slog.String("request-id", "9566c74d-1003-4c4d-bbbb-0407d1e2c649"),
			},
			wantLevel: slog.LevelInfo,
		},
		{
			name: "whitelist and blacklist test1",
			opts: []ConfigOption{
				WithWhitelistPath([]string{"/test1"}),
				WithBlacklistPath([]string{"/test1"}),
			},
			code:       200,
			wantFields: []slog.Attr{},
			wantLevel:  slog.LevelInfo,
			wantPanic:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set a fixed random seed to get a fixed request ID
			uuid.SetRand(rand.New(rand.NewSource(1)))

			// Create a new logger with a mock handler
			logger := slog.New(slogtest.NewMockHandler(
				slog.NewTextHandler(os.Stderr, nil),
				t,
				tt.wantLevel,
				tt.wantFields,
				skipFields,
			))

			router := gin.New()
			if tt.wantPanic {
				require.Panics(t, func() { router.Use(New(logger, tt.opts...)) })
				return
			}
			router.Use(New(logger, tt.opts...))

			// Define routes
			router.GET("/test1", func(c *gin.Context) {
				c.JSON(tt.code, nil)
			})
			router.GET("/test2", func(c *gin.Context) {
				c.JSON(tt.code, nil)
			})

			// Create the two requests
			resp := httptest.NewRecorder()
			req1, err := http.NewRequest("GET", "/test1", nil)
			req1.Header.Set("User-Agent", "test1")
			require.NoError(t, err)
			req2, err := http.NewRequest("GET", "/test2", nil)
			req2.Header.Set("User-Agent", "test2")
			require.NoError(t, err)
			router.ServeHTTP(resp, req1)
			router.ServeHTTP(resp, req2)
		})
	}
}

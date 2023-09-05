package recovery

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FabienMht/ginslog/slogtest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

var skipFields = []string{"error", "request", "stack"}

func TestNewRecovery(t *testing.T) {
	tests := []struct {
		name       string
		opts       []ConfigOption
		wantFields []slog.Attr
		wantLevel  slog.Level
		wantStatus int
		wantPanic  bool
	}{
		{
			name: "default options",
			opts: []ConfigOption{},
			wantFields: []slog.Attr{
				slog.Any("error", nil),
				slog.String("request", ""),
				slog.String("stack", ""),
			},
			wantLevel:  slog.LevelError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "with default level",
			opts: []ConfigOption{WithDefaultLevel(slog.LevelWarn)},
			wantFields: []slog.Attr{
				slog.Any("error", nil),
				slog.String("request", ""),
				slog.String("stack", ""),
			},
			wantLevel:  slog.LevelWarn,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "without error",
			opts: []ConfigOption{WithoutError()},
			wantFields: []slog.Attr{
				slog.String("request", ""),
				slog.String("stack", ""),
			},
			wantLevel:  slog.LevelError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "without request",
			opts: []ConfigOption{WithoutRequest()},
			wantFields: []slog.Attr{
				slog.Any("error", nil),
				slog.String("stack", ""),
			},
			wantLevel:  slog.LevelError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "without stack",
			opts: []ConfigOption{WithoutStack()},
			wantFields: []slog.Attr{
				slog.Any("error", nil),
				slog.String("request", ""),
			},
			wantLevel:  slog.LevelError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "without default fields",
			opts:       []ConfigOption{WithoutDefaultFields()},
			wantFields: []slog.Attr{},
			wantLevel:  slog.LevelError,
			wantStatus: http.StatusInternalServerError,
			wantPanic:  true,
		},
		{
			name: "with custom fields with default fields",
			opts: []ConfigOption{
				WithCustomFields(
					func(c *gin.Context) []slog.Attr {
						return []slog.Attr{slog.String("content-type", c.GetHeader("Content-Type"))}
					},
				),
			},
			wantFields: []slog.Attr{
				slog.Any("error", nil),
				slog.String("request", ""),
				slog.String("stack", ""),
				slog.String("content-type", "test"),
			},
			wantLevel:  slog.LevelError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "with custom fields without default fields",
			opts: []ConfigOption{
				WithoutDefaultFields(),
				WithCustomFields(
					func(c *gin.Context) []slog.Attr {
						return []slog.Attr{slog.String("content-type", c.GetHeader("Content-Type"))}
					},
				),
			},
			wantFields: []slog.Attr{
				slog.String("content-type", "test"),
			},
			wantLevel:  slog.LevelError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "with custom recovery",
			opts: []ConfigOption{
				WithCustomRecovery(func(c *gin.Context, err interface{}) {
					c.AbortWithStatus(http.StatusNotFound)
				}),
			},
			wantFields: []slog.Attr{
				slog.Any("error", nil),
				slog.String("request", ""),
				slog.String("stack", ""),
			},
			wantLevel:  slog.LevelError,
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				panic("test")
			})

			// Create a new request
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Content-Type", "test")
			require.NoError(t, err)
			router.ServeHTTP(resp, req)

			// Check the response
			require.Equal(t, tt.wantStatus, resp.Code)
		})
	}
}

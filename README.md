# ginslog

![Go Version](https://img.shields.io/github/go-mod/go-version/FabienMht/ginslog.svg)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/FabienMht/ginslog)
[![Go Report Card](https://goreportcard.com/badge/github.com/FabienMht/ginslog)](https://goreportcard.com/report/github.com/FabienMht/ginslog)
[![Sourcegraph](https://sourcegraph.com/github.com/FabienMht/ginslog/-/badge.svg)](https://sourcegraph.com/github.com/FabienMht/ginslog)
[![Tag](https://img.shields.io/github/tag/FabienMht/ginslog.svg)](https://github.com/FabienMht/ginslog/tags)
[![Contributors](https://img.shields.io/github/contributors/FabienMht/ginslog)](https://github.com/FabienMht/ginslog/graphs/contributors)
[![License](https://img.shields.io/github/license/FabienMht/ginslog)](./LICENSE)

A fully featured Gin middlewares for slog logging.
It includes a logging and a panic recovery middlewares.

## Install

Compatibility: go >= 1.21

**Download it:**

```
go get github.com/FabienMht/ginslog
```

**Add the following import:**

```go
ginlogger "github.com/FabienMht/ginslog/logger"
ginrecovery "github.com/FabienMht/ginslog/recovery"
```

## Usage

### Basic text handler

```go
package main

import (
    "log/slog"

    ginlogger "github.com/FabienMht/ginslog/logger"
    ginrecovery "github.com/FabienMht/ginslog/recovery"
    "github.com/gin-gonic/gin"
)

func main() {
    // Create a new logger instance with a text handler.
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

    r := gin.New()
    // Add the ginslog middleware to all routes.
    // Default attributes are:
    // - time
    // - level
    // - msg
    // - ip
    // - status
    // - method
    // - path
    // - user-agent
    // - latency
    // - request-id
    r.Use(ginlogger.New(logger))
    r.Use(ginrecovery.New(logger))

    r.GET("/test", func(c *gin.Context) {
        c.String(200, "Hello world!")
    })

    r.GET("/panic", func(c *gin.Context) {
        panic("Panic!")
    })

    r.Run(":80")
}
```

**Output:**

```bash
# Log incoming request
$ curl 127.0.0.1:80/test
time=2023-01-01T00:00:00.000+02:00 level=INFO msg="Incoming request" ip="127.0.0.1" status=200 method=GET path=/test user-agent=curl/7.86.0 latency=13.877µs request-id=52fdfc07-2182-454f-963f-5f0f9a621d72
# Log panic recovered with stack trace and request
$ curl 127.0.0.1:80/panic
time=2023-01-01T00:00:00.000+02:00 level=ERROR msg="Panic recovered" error="Unexpected error" request="GET /panic HTTP/1.1\r\nHost: 127.0.0.1:8080\r\nAccept: */*\r\nUser-Agent: curl/7.86.0\r\n\r\n" stack="goroutine 19 [running]:\nruntime/debug.Stack()\n\t/usr/lib/go/src/runtime/debug/stack.go:24 +0x5e\n...\ncreated by net/http.(*Server).Serve in goroutine 1\n\t/usr/lib/go/src/net/http/server.go:3086 +0x5cb\n"
time=2023-01-01T00:00:00.000+02:00 level=ERROR msg="Incoming request" ip=127.0.0.1 status=500 method=GET path=/panic user-agent=curl/7.86.0 latency=220.331µs
```

### Basic JSON handler

```go
package main

import (
    "log/slog"

    ginlogger "github.com/FabienMht/ginslog/logger"
    ginrecovery "github.com/FabienMht/ginslog/recovery"
    "github.com/gin-gonic/gin"
)

func main() {
    // Create a new logger instance with a JSON handler.
    logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

    r := gin.New()

    // Add the ginslog middleware to all routes.
    // Default attributes are:
    // - time
    // - level
    // - msg
    // - ip
    // - status
    // - method
    // - path
    // - user-agent
    // - latency
    // - request-id
    r.Use(ginlogger.New(logger))
    r.Use(ginrecovery.New(logger))

    r.GET("/test", func(c *gin.Context) {
        c.String(200, "Hello world!")
    })

    r.GET("/panic", func(c *gin.Context) {
        panic("Panic!")
    })

    r.Run(":80")
}
```

**Output:**

```bash
# Log incoming request
$ curl 127.0.0.1:80/test
{"time":"2023-01-01T00:00:00.000+02:00","level":"INFO","msg":"Incoming request","ip":"127.0.0.1","status":200,"method":"GET","path":"/test","user-agent":"curl/7.86.0","latency":43750,"request-id":"52fdfc07-2182-454f-963f-5f0f9a621d72"}
# Log panic recovered with stack trace and request
$ curl 127.0.0.1:80/panic
{"time":"2023-01-01T00:00:00.000+02:00","level":"ERROR","msg":"Panic recovered","error":"Unexpected error","request":"GET /panic HTTP/1.1\r\nHost: 127.0.0.1:8080\r\nAccept: */*\r\nUser-Agent: curl/7.86.0\r\n\r\n","stack":"goroutine 6 [running]:\nruntime/debug.Stack()\n\t/usr/lib/go/src/runtime/debug/stack.go:24 +0x5e\n...\ncreated by net/http.(*Server).Serve in goroutine 1\n\t/usr/lib/go/src/net/http/server.go:3086 +0x5cb\n"}
{"time":"2023-01-01T00:00:00.000+02:00","level":"ERROR","msg":"Incoming request","ip":"127.0.0.1","status":500,"method":"GET","path":"/panic","user-agent":"curl/7.86.0","latency":209845,"request-id":"52fdfc07-2182-454f-963f-5f0f9a621d72"}
```

## Contributing

Contributions are welcome ! Please open an issue or submit a pull request.

```bash
# Install task
$ go install github.com/go-task/task/v3/cmd/task@v3

# Install dev dependencies
$ task dev

# Run linter
$ task lint

# Run tests
$ task test
```

package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/idproxy/httpserver/pkg/hctx"
	"github.com/mattn/go-isatty"
)

// LoggerConfig defines the config for Logger middleware.
type LoggerConfig struct {
	// Optional. Default value is gin.defaultLogFormatter
	Formatter LogFormatter

	// Output is a writer where logs are written.
	// Optional. Default value is gin.DefaultWriter.
	Output io.Writer

	// SkipPaths is an url path array which logs are not written.
	// Optional.
	SkipPaths []string
}

// Logger is a Logger middleware that will write the logs to defaultWriter.
// By default, defaultWriter = os.Stdout.
func Logger() hctx.HandlerFunc {
	return LoggerWithConfig(LoggerConfig{})
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf LoggerConfig) hctx.HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultLogFormatter
	}

	out := conf.Output
	if out == nil {
		out = DefaultWriter
	}

	notlogged := conf.SkipPaths

	isTerm := true

	if w, ok := out.(*os.File); !ok || os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd())) {
		isTerm = false
	}

	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(hctx hctx.Context) {
		// Start timer
		start := time.Now()
		path := hctx.GetRequestPath()
		raw := hctx.GetRawQuery()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			param := LogFormatterParams{
				Request: hctx.GetRequest(),
				isTerm:  isTerm,
				//Keys:    gctx.Keys,
			}

			// Stop timer
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)

			param.ClientIP = hctx.ClientIP()
			param.Method = hctx.GetMethod()
			param.StatusCode = hctx.GetStatus()
			//param.ErrorMessage = hctx.
			//param.ErrorMessage = gctx.Errors.ByType(ErrorTypePrivate).String()

			//param.BodySize = gctx.Writer.Size()

			if raw != "" {
				path = path + "?" + raw
			}

			param.Path = path

			fmt.Fprint(out, formatter(param))
		}
	}
}

package log

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
)

type ContextKey string

const LoggerCtxKey ContextKey = "logger"

type LogResponseWriter struct {
	http.ResponseWriter
	statusCode int
	buf        bytes.Buffer
}

func (w *LogResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *LogResponseWriter) Write(body []byte) (int, error) {
	w.buf.Write(body)

	return w.ResponseWriter.Write(body)
}

func NewLogResponseWriter(w http.ResponseWriter) *LogResponseWriter {
	return &LogResponseWriter{ResponseWriter: w}
}

func LoggingMiddleware(logger *logrus.Entry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			newEntry := logger.WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
				"time":   time.Now(),
			})

			defer func() {
				if r := recover(); r != nil {
					w.WriteHeader(http.StatusInternalServerError)

					var err error

					switch x := r.(type) {
					case string:
						err = errors.New(x)
					case error:
						err = x
					default:
						err = errors.New("unknown panic")
					}

					logger.Errorf(
						"err=%v, trace=%v",
						err.Error(),
						string(debug.Stack()),
					)
				}
			}()

			ctx := context.WithValue(r.Context(), LoggerCtxKey, newEntry)

			startTime := time.Now()
			logRespWriter := NewLogResponseWriter(w)
			next.ServeHTTP(logRespWriter, r.WithContext(ctx))

			var bodyLog string
			contentType := logRespWriter.Header().Get("Content-Type")

			if contentType == "application/json; charset=utf-8" {
				bodyLog = logRespWriter.buf.String()
			}

			newEntry.Infof(
				"duration=%s status=%d body=%s",
				time.Since(startTime).String(),
				logRespWriter.statusCode,
				bodyLog,
			)
		})
	}
}

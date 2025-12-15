package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	data *responseData
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.data.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.data.size += size
	return size, err
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		data := &responseData{}
		lw := &loggingResponseWriter{ResponseWriter: w, data: data}

		next.ServeHTTP(lw, r)

		duration := time.Since(start)

		logrus.WithFields(logrus.Fields{
			"method":   r.Method,
			"uri":      r.RequestURI,
			"duration": duration,
			"status":   data.status,
			"size":     data.size,
		}).Info("request completed")
	})
}

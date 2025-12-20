package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "failed to decompress gzip body", http.StatusBadRequest)
				return
			}
			r.Body = gz
			defer gz.Close()
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gzw := gzip.NewWriter(w)
			defer gzw.Close()

			gzwResponse := gzipResponseWriter{Writer: gzw, ResponseWriter: w}
			next.ServeHTTP(gzwResponse, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

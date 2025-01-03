package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

// GzipMiddleware compresses HTTP responses with gzip when accepted by the client.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip content", http.StatusBadRequest)
				return
			}
			defer gr.Close()
			r.Body = io.NopCloser(gr)
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz := gzip.NewWriter(w)
			defer gz.Close()

			gzw := gzipResponseWriter{ResponseWriter: w, writer: gz}
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			next.ServeHTTP(gzw, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// Package middleware provides HTTP middleware components for the URL shortener service.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipResponseWriter wraps the http.ResponseWriter and allows
// writing compressed response data.
type gzipResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

// Write overrides the default Write method to use the Gzip writer
// for compressing response data.
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

// GzipMiddleware is an HTTP middleware that compresses the response
// using gzip if the client supports it (indicated by the "Accept-Encoding" header).
// It also decompresses Gzip-compressed request bodies if needed.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Decompress the request body if it is gzip-compressed.
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip content", http.StatusBadRequest)
				return
			}
			defer gr.Close()
			r.Body = io.NopCloser(gr)
		}

		// Compress the response using gzip if the client accepts it.
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

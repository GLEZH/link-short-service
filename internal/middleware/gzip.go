package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter  *gzip.Writer
	statusCode  int
	wroteHeader bool
	canCompress bool
}

func WithGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = io.NopCloser(gz)
		}

		gzipWriter := &gzipResponseWriter{
			ResponseWriter: w,
			canCompress:    strings.Contains(r.Header.Get("Accept-Encoding"), "gzip"),
		}

		next.ServeHTTP(gzipWriter, r)

		if !gzipWriter.wroteHeader {
			gzipWriter.writeHeader()
		}

		if gzipWriter.gzipWriter != nil {
			_ = gzipWriter.gzipWriter.Close()
		}
	})
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.writeHeader()
	}

	if w.gzipWriter != nil {
		return w.gzipWriter.Write(data)
	}

	return w.ResponseWriter.Write(data)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *gzipResponseWriter) writeHeader() {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	if w.canCompress && isCompressibleContentType(w.Header().Get("Content-Type")) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		w.gzipWriter = gzip.NewWriter(w.ResponseWriter)
	}

	w.ResponseWriter.WriteHeader(w.statusCode)
	w.wroteHeader = true
}

func isCompressibleContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "text/html")
}

package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (w *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.responseData.size += size

	if w.responseData.status == 0 {
		w.responseData.status = http.StatusOK
	}

	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

func WithLogging(next http.Handler, log *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{}

		loggingWriter := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(loggingWriter, r)

		if responseData.status == 0 {
			responseData.status = http.StatusOK
		}

		log.Infow(
			"request handled",
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", time.Since(start),
			"status", responseData.status,
			"size", responseData.size,
		)
	})
}

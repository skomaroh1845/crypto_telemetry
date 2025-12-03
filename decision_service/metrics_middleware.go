package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func MetricsMiddleware(m *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()

			// обёртка чтобы получить статус ответа
			sw := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(sw, r)

			duration := time.Since(start).Seconds()

			// получаем маршрут из chi
			route := "unknown"
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				route = rctx.RoutePattern()
			}

			isError := sw.status >= 400

			m.RecordRequest(r.Context(), route, duration, isError)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

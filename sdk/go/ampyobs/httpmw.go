package ampyobs

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func HTTPServerMiddleware(hdl *Handle) func(next http.Handler) http.Handler {
	prop := otel.GetTextMapPropagator()
	tr := hdl.Tracer("http.server")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := prop.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := tr.Start(ctx, r.Method+" "+r.URL.Path)
			defer span.End()

			start := time.Now()
			ww := &respWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(ww, r.WithContext(ctx))

			hdl.Logger.Info(ctx, "http.request",
				F("method", r.Method),
				F("path", r.URL.Path),
				F("status", ww.status),
				F("latency_ms", time.Since(start).Milliseconds()),
			)
		})
	}
}

type respWriter struct {
	http.ResponseWriter
	status int
}

func (w *respWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

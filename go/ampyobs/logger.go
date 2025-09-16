package ampyobs

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
)

var rootLogger *slog.Logger

func setupSlog(_ any) {
	// JSON handler, info level default
	rootLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Force ISO8601 for time
			if a.Key == slog.TimeKey {
				if t := a.Value.Time(); !t.IsZero() {
					a.Value = slog.StringValue(t.UTC().Format(time.RFC3339Nano))
				}
			}
			return a
		},
	}))
}

// L returns a *slog.Logger without context.
func L() *slog.Logger {
	if rootLogger == nil {
		setupSlog(nil)
	}
	return rootLogger.With(
		slog.String("service", globalCfg.ServiceName),
		slog.String("env", globalCfg.Environment),
		slog.String("service_version", globalCfg.ServiceVersion),
	)
}

// C returns a context-aware logger that enriches with trace/span if present.
func C(ctx context.Context) *slog.Logger {
	l := L()
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		l = l.With(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return l
}

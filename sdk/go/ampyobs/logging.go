package ampyobs

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	With(kv ...zap.Field) Logger
	Info(ctx context.Context, msg string, kv ...zap.Field)
	Warn(ctx context.Context, msg string, kv ...zap.Field)
	Error(ctx context.Context, msg string, kv ...zap.Field)
	Debug(ctx context.Context, msg string, kv ...zap.Field)
}

type zapLogger struct {
	base *zap.Logger
	meta []zap.Field // static fields: service, env, version
}

func newLogger(cfg Config) Logger {
	encCfg := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		MessageKey:    "message",
		CallerKey:     "caller",
		StacktraceKey: "stack",

		EncodeTime:   func(t time.Time, enc zapcore.PrimitiveArrayEncoder) { enc.AppendString(t.UTC().Format(time.RFC3339Nano)) },
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
		LineEnding:   zapcore.DefaultLineEnding,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encCfg), zapcore.AddSync(os.Stdout), zap.DebugLevel)
	z := zap.New(core, zap.AddCaller())

	meta := []zap.Field{
		zap.String("service", cfg.ServiceName),
		zap.String("env", cfg.Environment),
		zap.String("service_version", cfg.ServiceVersion),
	}
	return &zapLogger{base: z, meta: meta}
}

func (l *zapLogger) With(kv ...zap.Field) Logger {
	return &zapLogger{base: l.base, meta: append(append([]zap.Field{}, l.meta...), kv...)}
}

func (l *zapLogger) Info(ctx context.Context, msg string, kv ...zap.Field)  { l.log(ctx, zap.InfoLevel, msg, kv...) }
func (l *zapLogger) Warn(ctx context.Context, msg string, kv ...zap.Field)  { l.log(ctx, zap.WarnLevel, msg, kv...) }
func (l *zapLogger) Error(ctx context.Context, msg string, kv ...zap.Field) { l.log(ctx, zap.ErrorLevel, msg, kv...) }
func (l *zapLogger) Debug(ctx context.Context, msg string, kv ...zap.Field) { l.log(ctx, zap.DebugLevel, msg, kv...) }

func (l *zapLogger) log(ctx context.Context, level zapcore.Level, msg string, kv ...zap.Field) {
	fields := append([]zap.Field{}, l.meta...)

	// Attach trace/span ids if present
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		fields = append(fields,
			zap.String("trace_id", sc.TraceID().String()),
			zap.String("span_id", sc.SpanID().String()),
		)
	}

	// Attach domain context fields (if present)
	if dc, ok := FromDomainContext(ctx); ok {
		fields = append(fields, dc.toZapFields()...)
	}

	fields = append(fields, kv...)

	switch level {
	case zap.DebugLevel:
		l.base.Debug(msg, fields...)
	case zap.InfoLevel:
		l.base.Info(msg, fields...)
	case zap.WarnLevel:
		l.base.Warn(msg, fields...)
	case zap.ErrorLevel:
		l.base.Error(msg, fields...)
	default:
		l.base.Info(msg, fields...)
	}
}

// Helper so callers can pass fields without importing zap directly (optional).
func F(k string, v any) zap.Field { return zap.Any(k, v) }

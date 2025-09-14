package ampyobs

import (
	"context"

	"go.uber.org/zap"
)

type domainKey struct{}

type DomainContext struct {
	RunID         string
	AsOfISO       string
	UniverseID    string
	MessageID     string
	ClientOrderID string
	Symbol        string
	MIC           string
}

func WithDomainContext(ctx context.Context, dc DomainContext) context.Context {
	return context.WithValue(ctx, domainKey{}, dc)
}

func FromDomainContext(ctx context.Context) (DomainContext, bool) {
	v := ctx.Value(domainKey{})
	if v == nil {
		return DomainContext{}, false
	}
	dc, ok := v.(DomainContext)
	return dc, ok
}

func (d DomainContext) toZapFields() []zap.Field {
	out := make([]zap.Field, 0, 8)
	if d.RunID != "" {
		out = append(out, zap.String("run_id", d.RunID))
	}
	if d.AsOfISO != "" {
		out = append(out, zap.String("as_of", d.AsOfISO))
	}
	if d.UniverseID != "" {
		out = append(out, zap.String("universe_id", d.UniverseID))
	}
	if d.MessageID != "" {
		out = append(out, zap.String("message_id", d.MessageID))
	}
	if d.ClientOrderID != "" {
		out = append(out, zap.String("client_order_id", d.ClientOrderID))
	}
	if d.Symbol != "" {
		out = append(out, zap.String("symbol", d.Symbol))
	}
	if d.MIC != "" {
		out = append(out, zap.String("mic", d.MIC))
	}
	return out
}

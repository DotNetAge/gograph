package cypher

import (
	"context"
)

// defaultLogger is a Logger implementation using slog.
// (It's not really no-op but it's the default, let's keep it here if it's considered "hollow")
// Actually, it uses slog, so it's fine.

// defaultTracer is a no-op Tracer implementation.
type defaultTracer struct{}

func (t *defaultTracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &defaultSpan{name: name}
}

func (t *defaultTracer) RecordError(ctx context.Context, err error, attrs ...Attribute) {}

// defaultSpan is a no-op Span implementation.
type defaultSpan struct {
	name string
}

func (s *defaultSpan) End() {}

func (s *defaultSpan) AddEvent(name string, attrs ...Attribute) {}

func (s *defaultSpan) SetAttributes(attrs ...Attribute) {}

// defaultMeter is a no-op Meter implementation.
type defaultMeter struct{}

func (m *defaultMeter) RecordHistogram(ctx context.Context, name string, value float64, attrs ...Attribute) {
}

func (m *defaultMeter) RecordCounter(ctx context.Context, name string, value float64, attrs ...Attribute) {
}

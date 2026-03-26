package cypher

import (
	"context"
	"log/slog"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type Tracer interface {
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
	RecordError(ctx context.Context, err error, attrs ...Attribute)
}

type Span interface {
	End()
	AddEvent(name string, attrs ...Attribute)
	SetAttributes(attrs ...Attribute)
}

type SpanOption func(*spanConfig)

type spanConfig struct {
	attrs []Attribute
}

type Attribute struct {
	Key   string
	Value any
}

type Meter interface {
	RecordHistogram(ctx context.Context, name string, value float64, attrs ...Attribute)
	RecordCounter(ctx context.Context, name string, value float64, attrs ...Attribute)
}

type defaultLogger struct{}

func (l *defaultLogger) Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func (l *defaultLogger) Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func (l *defaultLogger) Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func (l *defaultLogger) Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

type defaultTracer struct{}

func (t *defaultTracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &defaultSpan{name: name}
}

func (t *defaultTracer) RecordError(ctx context.Context, err error, attrs ...Attribute) {}

type defaultSpan struct {
	name string
}

func (s *defaultSpan) End() {}

func (s *defaultSpan) AddEvent(name string, attrs ...Attribute) {}

func (s *defaultSpan) SetAttributes(attrs ...Attribute) {}

type defaultMeter struct{}

func (m *defaultMeter) RecordHistogram(ctx context.Context, name string, value float64, attrs ...Attribute) {}

func (m *defaultMeter) RecordCounter(ctx context.Context, name string, value float64, attrs ...Attribute) {}

type Observability struct {
	Logger Logger
	Tracer Tracer
	Meter  Meter
}

type ObservabilityOption func(*Observability)

func WithLogger(l Logger) ObservabilityOption {
	return func(o *Observability) {
		o.Logger = l
	}
}

func WithTracer(t Tracer) ObservabilityOption {
	return func(o *Observability) {
		o.Tracer = t
	}
}

func WithMeter(m Meter) ObservabilityOption {
	return func(o *Observability) {
		o.Meter = m
	}
}

func NewObservability(opts ...ObservabilityOption) *Observability {
	o := &Observability{
		Logger: &defaultLogger{},
		Tracer: &defaultTracer{},
		Meter:  &defaultMeter{},
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

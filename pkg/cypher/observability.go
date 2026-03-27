// Package cypher provides Cypher query parsing and execution capabilities for gograph.
package cypher

import (
	"context"
	"log/slog"
)

// Logger defines the interface for logging operations.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Tracer defines the interface for distributed tracing.
type Tracer interface {
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
	RecordError(ctx context.Context, err error, attrs ...Attribute)
}

// Span represents a tracing span for observing execution.
type Span interface {
	End()
	AddEvent(name string, attrs ...Attribute)
	SetAttributes(attrs ...Attribute)
}

// SpanOption configures span creation.
type SpanOption func(*spanConfig)

type spanConfig struct {
	attrs []Attribute
}

// Attribute represents a key-value attribute for tracing and metrics.
type Attribute struct {
	Key   string
	Value any
}

// Meter defines the interface for recording metrics.
type Meter interface {
	RecordHistogram(ctx context.Context, name string, value float64, attrs ...Attribute)
	RecordCounter(ctx context.Context, name string, value float64, attrs ...Attribute)
}

// defaultLogger is a Logger implementation using slog.
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

// Observability contains the logging, tracing, and metering interfaces
// used throughout the Cypher execution pipeline.
type Observability struct {
	Logger Logger
	Tracer Tracer
	Meter  Meter
}

// ObservabilityOption configures optional parameters for observability.
type ObservabilityOption func(*Observability)

// WithLogger sets the Logger implementation.
func WithLogger(l Logger) ObservabilityOption {
	return func(o *Observability) {
		o.Logger = l
	}
}

// WithTracer sets the Tracer implementation.
func WithTracer(t Tracer) ObservabilityOption {
	return func(o *Observability) {
		o.Tracer = t
	}
}

// WithMeter sets the Meter implementation.
func WithMeter(m Meter) ObservabilityOption {
	return func(o *Observability) {
		o.Meter = m
	}
}

// NewObservability creates a new Observability instance with default (no-op) implementations.
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

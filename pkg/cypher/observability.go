// Package cypher provides Cypher query parsing and execution capabilities for gograph.
package cypher

import (
	"context"
	"log/slog"
)

// Logger defines the interface for logging operations.
// Implementations can use different logging backends (slog, zap, logrus, etc.).
type Logger interface {
	// Debug logs a debug message with optional key-value pairs.
	Debug(msg string, args ...any)

	// Info logs an informational message with optional key-value pairs.
	Info(msg string, args ...any)

	// Warn logs a warning message with optional key-value pairs.
	Warn(msg string, args ...any)

	// Error logs an error message with optional key-value pairs.
	Error(msg string, args ...any)
}

// Tracer defines the interface for distributed tracing.
// Implementations can use different tracing backends (OpenTelemetry, Jaeger, etc.).
type Tracer interface {
	// StartSpan starts a new span with the given name and options.
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

	// RecordError records an error on the current span.
	RecordError(ctx context.Context, err error, attrs ...Attribute)
}

// Span represents a tracing span for observing execution.
type Span interface {
	// End completes the span.
	End()

	// AddEvent adds an event to the span.
	AddEvent(name string, attrs ...Attribute)

	// SetAttributes sets attributes on the span.
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
// Implementations can use different metrics backends (Prometheus, StatsD, etc.).
type Meter interface {
	// RecordHistogram records a histogram observation.
	RecordHistogram(ctx context.Context, name string, value float64, attrs ...Attribute)

	// RecordCounter increments a counter.
	RecordCounter(ctx context.Context, name string, value float64, attrs ...Attribute)
}

// defaultLogger is a Logger implementation using slog.
type defaultLogger struct{}

// Debug logs a debug message using slog.
func (l *defaultLogger) Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

// Info logs an informational message using slog.
func (l *defaultLogger) Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Warn logs a warning message using slog.
func (l *defaultLogger) Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

// Error logs an error message using slog.
func (l *defaultLogger) Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

// defaultTracer is a no-op Tracer implementation.
type defaultTracer struct{}

// StartSpan starts a no-op span.
func (t *defaultTracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &defaultSpan{}
}

// RecordError is a no-op.
func (t *defaultTracer) RecordError(ctx context.Context, err error, attrs ...Attribute) {}

// defaultSpan is a no-op Span implementation.
type defaultSpan struct{}

// End is a no-op.
func (s *defaultSpan) End() {}

// AddEvent is a no-op.
func (s *defaultSpan) AddEvent(name string, attrs ...Attribute) {}

// SetAttributes is a no-op.
func (s *defaultSpan) SetAttributes(attrs ...Attribute) {}

// defaultMeter is a no-op Meter implementation.
type defaultMeter struct{}

// RecordHistogram is a no-op.
func (m *defaultMeter) RecordHistogram(ctx context.Context, name string, value float64, attrs ...Attribute) {
}

// RecordCounter is a no-op.
func (m *defaultMeter) RecordCounter(ctx context.Context, name string, value float64, attrs ...Attribute) {
}

// Observability contains the logging, tracing, and metering interfaces
// used throughout the Cypher execution pipeline.
//
// It provides a unified interface for observability concerns, allowing
// users to plug in their own implementations or use the defaults.
//
// Example:
//
//	// Use default observability (uses slog for logging)
//	obs := cypher.NewObservability()
//
//	// Use custom logger
//	obs := cypher.NewObservability(cypher.WithLogger(myLogger))
//
//	// Use custom tracer and meter
//	obs := cypher.NewObservability(
//	    cypher.WithTracer(myTracer),
//	    cypher.WithMeter(myMeter),
//	)
type Observability struct {
	// Logger is used for logging operations.
	Logger Logger

	// Tracer is used for distributed tracing.
	Tracer Tracer

	// Meter is used for recording metrics.
	Meter Meter
}

// ObservabilityOption configures optional parameters for observability.
type ObservabilityOption func(*Observability)

// WithLogger sets the Logger implementation.
//
// Parameters:
//   - l: The Logger implementation to use
//
// Returns an ObservabilityOption that can be passed to NewObservability.
//
// Example:
//
//	obs := cypher.NewObservability(cypher.WithLogger(myLogger))
func WithLogger(l Logger) ObservabilityOption {
	return func(o *Observability) {
		o.Logger = l
	}
}

// WithTracer sets the Tracer implementation.
//
// Parameters:
//   - t: The Tracer implementation to use
//
// Returns an ObservabilityOption that can be passed to NewObservability.
//
// Example:
//
//	obs := cypher.NewObservability(cypher.WithTracer(myTracer))
func WithTracer(t Tracer) ObservabilityOption {
	return func(o *Observability) {
		o.Tracer = t
	}
}

// WithMeter sets the Meter implementation.
//
// Parameters:
//   - m: The Meter implementation to use
//
// Returns an ObservabilityOption that can be passed to NewObservability.
//
// Example:
//
//	obs := cypher.NewObservability(cypher.WithMeter(myMeter))
func WithMeter(m Meter) ObservabilityOption {
	return func(o *Observability) {
		o.Meter = m
	}
}

// NewObservability creates a new Observability instance with default implementations.
// By default, it uses slog for logging and no-op implementations for tracing and metrics.
//
// Parameters:
//   - opts: Optional configuration options
//
// Returns a new Observability instance.
//
// Example:
//
//	// Default observability
//	obs := cypher.NewObservability()
//
//	// With custom options
//	obs := cypher.NewObservability(
//	    cypher.WithLogger(myLogger),
//	    cypher.WithTracer(myTracer),
//	)
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

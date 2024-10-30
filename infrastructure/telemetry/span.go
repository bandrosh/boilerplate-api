package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Span(ctx context.Context, name string, kind trace.SpanKind) (context.Context, trace.Span) {
	ctx, span := otel.GetTracerProvider().Tracer(ctx.Value("service-name").(string)).Start(
		ctx, name, trace.WithSpanKind(kind),
	)

	return ctx, span
}

func ErrorSpan(span trace.Span, err error) {
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}

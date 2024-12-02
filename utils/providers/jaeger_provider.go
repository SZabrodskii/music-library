package providers

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type JaegerProviderConfig struct {
	Endpoint string
}

func NewJaegerProviderConfig() *JaegerProviderConfig {
	return &JaegerProviderConfig{
		Endpoint: "http://localhost:4317",
	}
}

func NewJaegerProvider(logger *zap.Logger, config *JaegerProviderConfig) (*trace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpoint(config.Endpoint))
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

func ShutdownJaegerProvider(tp *trace.TracerProvider) {
	_ = tp.Shutdown(context.Background())
}

func JaegerProviderModule() fx.Option {
	return fx.Options(
		fx.Provide(NewJaegerProviderConfig, NewJaegerProvider),
		fx.Invoke(ShutdownJaegerProvider),
	)
}

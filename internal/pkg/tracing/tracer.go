package tracing

import (
	"fmt"
	"io"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-client-go/zipkin"
)

// InitTracer initializes a new OpenTracing tracer with Tempo as the backend
// We're still using the Jaeger client as Tempo supports the Jaeger protocol
func InitTracer(serviceName string, tempoHostPort string) (opentracing.Tracer, io.Closer, error) {
	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  tempoHostPort, // Now points to Tempo instead of Jaeger
		},
	}

	// Initialize tracer with zipkin propagation format
	jLogger := jaegerlog.StdLogger
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.ZipkinSharedRPCSpan(true),
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot initialize OpenTracing Tracer: %w", err)
	}

	opentracing.SetGlobalTracer(tracer)
	return tracer, closer, nil
}

package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// MetadataTextMap adapts gRPC metadata to opentracing TextMap interface
type MetadataTextMap metadata.MD

// Set implements opentracing.TextMapWriter
func (m MetadataTextMap) Set(key, val string) {
	metadata.MD(m).Set(key, val)
}

// ForeachKey implements opentracing.TextMapReader
func (m MetadataTextMap) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range metadata.MD(m) {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// UnaryServerInterceptor returns a grpc.UnaryServerInterceptor for OpenTracing
func UnaryServerInterceptor(tracer opentracing.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		spanContext, err := tracer.Extract(opentracing.HTTPHeaders, MetadataTextMap(md))
		var span opentracing.Span
		if err != nil {
			// Create a new span if no parent span was found
			span = tracer.StartSpan(info.FullMethod)
		} else {
			// Create a child span if parent span was found
			span = tracer.StartSpan(info.FullMethod, ext.RPCServerOption(spanContext))
		}
		defer span.Finish()

		// Set span tags
		ext.SpanKindRPCServer.Set(span)
		ext.Component.Set(span, "gRPC")

		// Store span in context
		ctx = opentracing.ContextWithSpan(ctx, span)

		// Call the handler with the new context
		resp, err := handler(ctx, req)
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}
		return resp, err
	}
}

// UnaryClientInterceptor returns a grpc.UnaryClientInterceptor for OpenTracing
func UnaryClientInterceptor(tracer opentracing.Tracer) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var parentSpanCtx opentracing.SpanContext
		if parent := opentracing.SpanFromContext(ctx); parent != nil {
			parentSpanCtx = parent.Context()
		}

		span := tracer.StartSpan(
			method,
			opentracing.ChildOf(parentSpanCtx),
			ext.SpanKindRPCClient,
			opentracing.Tag{Key: "component", Value: "gRPC"},
		)
		defer span.Finish()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		mdWriter := MetadataTextMap(md)
		err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, mdWriter)
		if err != nil {
			span.LogFields(log.String("event", "inject-error"), log.Error(err))
		}

		ctx = metadata.NewOutgoingContext(ctx, metadata.MD(mdWriter))
		err = invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}
		return err
	}
}

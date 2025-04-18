package tracing

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// GinMiddleware returns a gin middleware for OpenTracing
func GinMiddleware(tracer opentracing.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the parent span context from the HTTP headers
		spanCtx, err := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header),
		)

		var span opentracing.Span
		if err != nil {
			// Create a new span if no parent span was found
			span = tracer.StartSpan(c.Request.URL.Path)
		} else {
			// Create a child span if parent span was found
			span = tracer.StartSpan(
				c.Request.URL.Path,
				ext.RPCServerOption(spanCtx),
			)
		}
		defer span.Finish()

		// Set span tags
		ext.HTTPMethod.Set(span, c.Request.Method)
		ext.HTTPUrl.Set(span, c.Request.URL.String())
		ext.SpanKindRPCServer.Set(span)
		ext.Component.Set(span, "gin")

		// Store span in context
		c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), span))

		// Process request
		c.Next()

		// Set status code after request is processed
		ext.HTTPStatusCode.Set(span, uint16(c.Writer.Status()))
		if c.Writer.Status() >= http.StatusInternalServerError {
			ext.Error.Set(span, true)
		}
	}
}
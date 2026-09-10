package reqlog

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"go.opentelemetry.io/otel/codes"
	otel_trace "go.opentelemetry.io/otel/trace"

	"github.com/Southclaws/storyden/app/transports/http/middleware/origin"
	"github.com/Southclaws/storyden/internal/infrastructure/instrumentation/kv"
)

type Middleware struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Middleware {
	return &Middleware{logger: logger}
}

type withStatus struct {
	http.ResponseWriter
	statusCode int
}

func (w *withStatus) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (lrw *withStatus) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (m *Middleware) WithLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ctx := r.Context()
			span := otel_trace.SpanFromContext(ctx)
			spanContext := span.SpanContext()

			title := r.Method + " " + r.URL.Path

			attrs := kv.Attrs{
				kv.String("http.request.header.origin", origin.GetOrigin(ctx)),
				kv.String("client.address", r.RemoteAddr),
				kv.String("http.request.method", r.Method),
				kv.String("url.query", r.URL.Query().Encode()),
				kv.Int("http.request.body.size", int(r.ContentLength)),
			}

			span.SetAttributes(attrs.ToAttributes()...)
			ctx = fctx.WithMeta(ctx, attrs.ToFault()...)

			logger := m.logger.With(attrs.ToSlog()...).With(
				slog.String("trace_id", spanContext.TraceID().String()),
				slog.String("span_id", spanContext.SpanID().String()),
			)

			wr := &withStatus{ResponseWriter: w}

			defer func() {
				span.SetAttributes(kv.Attrs{
					kv.Duration("duration", time.Since(start)),
					kv.Int("http.response.status_code", wr.statusCode),
				}.ToAttributes()...)

				logger.Info(title)

				if recovery := recover(); recovery != nil {
					err := func(v any) error {
						if e, ok := v.(error); ok {
							return e
						}
						return fmt.Errorf("%v", v)
					}(recovery)

					stack := debug.Stack()
					errorlog := title + ": " + err.Error()

					span.SetStatus(codes.Error, errorlog)
					span.RecordError(err)

					wrapped := fault.Wrap(err,
						fctx.With(fctx.WithMeta(ctx, "trace", string(stack))),
						fmsg.With(errorlog),
					)

					logger.Error(errorlog, slog.String("error", wrapped.Error()))

					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}()

			next.ServeHTTP(wr, r.WithContext(ctx))
		})
	}
}

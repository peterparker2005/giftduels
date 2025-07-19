package interceptors

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// CorrelationInterceptorUnary — для Unary‑RPC.
func CorrelationInterceptorUnary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx = injectCorrelation(ctx)
		return handler(ctx, req)
	}
}

// CorrelationInterceptorStream — для Stream‑RPC.
func CorrelationInterceptorStream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Создаем новый контекст с request-id / trace-id
		ctx := injectCorrelation(ss.Context())
		// Оборачиваем stream, чтобы Context() вернул наш новый контекст
		wrapped := &correlatedStream{
			ServerStream: ss,
			ctx:          ctx,
		}
		return handler(srv, wrapped)
	}
}

// correlatedStream оборачивает исходный ServerStream, подменяя Context().
type correlatedStream struct {
	grpc.ServerStream

	ctx context.Context
}

// Context возвращает наш скорректированный контекст.
func (c *correlatedStream) Context() context.Context {
	return c.ctx
}

// injectCorrelation читает/генерирует ID и клонирует metadata в новый контекст.
func injectCorrelation(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)

	// 1) Request‑ID
	var reqID string
	if vals := md.Get("x-request-id"); len(vals) > 0 {
		reqID = vals[0]
	} else {
		reqID = uuid.NewString()
	}

	// 2) Trace‑ID
	var traceID string
	if vals := md.Get("x-trace-id"); len(vals) > 0 {
		traceID = vals[0]
	} else {
		traceID = uuid.NewString()
	}

	// Обновляем входящий metadata
	md.Set("x-request-id", reqID)
	md.Set("x-trace-id", traceID)
	ctx = metadata.NewIncomingContext(ctx, md)

	// И сразу подготавливаем исходящий для downstream‑вызовов
	outMD := metadata.Pairs(
		"x-request-id", reqID,
		"x-trace-id", traceID,
	)
	ctx = metadata.NewOutgoingContext(ctx, outMD)
	return ctx
}

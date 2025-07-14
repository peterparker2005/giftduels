package logger

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const requestIDKey contextKey = "x-request-id"

// LoggerWithRequestID создает логгер с request_id из контекста.
func LoggerWithRequestID(ctx context.Context, baseLogger *Logger) *Logger {
	if reqID, ok := GetRequestID(ctx); ok {
		return baseLogger.With(RequestID(reqID))
	}
	return baseLogger.With(RequestID(uuid.NewString()))
}

// WithRequestID добавляет request_id в контекст.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestID извлекает request_id из контекста.
func GetRequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

// RequestIDOrEmpty возвращает request_id из контекста или пустую строку.
func RequestIDOrEmpty(ctx context.Context) string {
	id, _ := GetRequestID(ctx)
	return id
}

// FromIncomingContext получает request_id из входящего metadata или генерирует новый.
func FromIncomingContext(ctx context.Context) context.Context {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-request-id"); len(vals) > 0 {
			return WithRequestID(ctx, vals[0])
		}
	}
	return WithRequestID(ctx, uuid.NewString())
}

// OutgoingContextWithRequestID добавляет request_id в исходящий metadata.
func OutgoingContextWithRequestID(ctx context.Context) context.Context {
	id, ok := GetRequestID(ctx)
	if !ok {
		id = uuid.NewString()
		ctx = WithRequestID(ctx, id)
	}
	return metadata.AppendToOutgoingContext(ctx, "x-request-id", id)
}

// WithRequestIDIfMissing устанавливает request_id если он не установлен.
func WithRequestIDIfMissing(ctx context.Context) context.Context {
	if _, ok := GetRequestID(ctx); !ok {
		return WithRequestID(ctx, uuid.NewString())
	}
	return ctx
}

// UnaryRequestIDInterceptor - интерцептор для унарных gRPC вызовов (сервер).
func UnaryRequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = FromIncomingContext(ctx)
		return handler(ctx, req)
	}
}

// StreamRequestIDInterceptor - интерцептор для потоковых gRPC вызовов (сервер).
func StreamRequestIDInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := FromIncomingContext(ss.Context())
		wrapped := &wrappedStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrapped)
	}
}

type wrappedStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// UnaryClientRequestIDInterceptor - интерцептор для унарных gRPC вызовов (клиент).
func UnaryClientRequestIDInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
	) error {
		ctx = OutgoingContextWithRequestID(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientRequestIDInterceptor - интерцептор для потоковых gRPC вызовов (клиент).
func StreamClientRequestIDInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
		method string, streamer grpc.Streamer, opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx = OutgoingContextWithRequestID(ctx)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

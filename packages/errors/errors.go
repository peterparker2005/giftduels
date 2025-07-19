package errors

import (
	"context"

	errorsv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/errors/v1"
	errdetails "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

// Option — функция‑опция для настройки ошибки.
type Option func(*errConfig)

type errConfig struct {
	grpcCode     codes.Code
	errorCode    errorsv1.ErrorCode
	message      string
	field        string
	metadata     map[string]string
	traceID      string
	resourceType string
	resourceID   string
}

func defaultConfig() *errConfig {
	return &errConfig{
		grpcCode:  codes.Internal,
		errorCode: errorsv1.ErrorCode_ERROR_CODE_UNSPECIFIED,
		message:   "",
		field:     "",
		metadata:  map[string]string{},
		traceID:   "",
	}
}

// --- Опции ---

// WithGRPCCode задаёт gRPC-код.
func WithGRPCCode(c codes.Code) Option {
	return func(cfg *errConfig) { cfg.grpcCode = c }
}

// WithErrorCode задаёт бизнес‑код из enum ErrorCode.
func WithErrorCode(ec errorsv1.ErrorCode) Option {
	return func(cfg *errConfig) { cfg.errorCode = ec }
}

// WithMessage задаёт сообщение.
func WithMessage(m string) Option {
	return func(cfg *errConfig) { cfg.message = m }
}

// WithField указывает поле (для валидации).
func WithField(f string) Option {
	return func(cfg *errConfig) { cfg.field = f }
}

// WithMetadata добавляет пару ключ–значение.
func WithMetadata(k, v string) Option {
	return func(cfg *errConfig) { cfg.metadata[k] = v }
}

// WithResource указывает тип и ID ресурса.
func WithResource(t, id string) Option {
	return func(cfg *errConfig) {
		cfg.resourceType = t
		cfg.resourceID = id
	}
}

// WithContext извлекает из ctx trace_id и помещает его и в metadata, и в отдельное поле.
func WithContext(ctx context.Context) Option {
	return func(cfg *errConfig) {
		if tid, ok := ctx.Value("x-trace-id").(string); ok && tid != "" {
			cfg.traceID = tid
			cfg.metadata["x-trace-id"] = tid
		}
	}
}

// NewError builds status.Status with details:
// 1) errdetails.BadRequest (if cfg.field is set)
// 2) errdetails.ErrorInfo (always)
// 3) errorsv1.ErrorContext (if any of: traceID, resourceType, resourceID, metadata is set).
func NewError(opts ...Option) error {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}

	st := status.New(cfg.grpcCode, cfg.message)

	// Now we collect a slice of protoadapt.MessageV1
	var details []protoadapt.MessageV1

	// a) BadRequest (if cfg.field is set)
	if cfg.field != "" {
		br := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{{
				Field:       cfg.field,
				Description: cfg.message,
			}},
		}
		details = append(details, protoadapt.MessageV1Of(br))
	}

	// b) ErrorInfo (always)
	ei := &errdetails.ErrorInfo{
		Reason:   cfg.errorCode.String(),
		Domain:   cfg.resourceType,
		Metadata: cfg.metadata,
	}
	details = append(details, protoadapt.MessageV1Of(ei))

	// c) Your ErrorContext (if any of: traceID, resourceType, resourceID, metadata is set)
	if cfg.traceID != "" || cfg.resourceType != "" || cfg.resourceID != "" || len(cfg.metadata) > 0 {
		ec := &errorsv1.ErrorContext{
			TraceId:      cfg.traceID,
			ResourceType: cfg.resourceType,
			ResourceId:   cfg.resourceID,
			Metadata:     cfg.metadata,
		}
		details = append(details, protoadapt.MessageV1Of(ec))
	}

	// We pass exactly []protoadapt.MessageV1 → WithDetails accepts them without errors
	st2, err := st.WithDetails(details...)
	if err != nil {
		// If something went wrong — we return just without details
		return status.Error(cfg.grpcCode, cfg.message)
	}
	return st2.Err()
}

// Wrap transforms any err into a status:
//
//	– if it's already a status — returns it as is
//	– otherwise — wraps it in Internal with a message and attaches the context
func Wrap(ctx context.Context, err error) error {
	if st, ok := status.FromError(err); ok {
		return st.Err()
	}
	return NewError(
		WithGRPCCode(codes.Internal),
		WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
		WithMessage(err.Error()),
		WithContext(ctx),
	)
}

// IsCode checks if the error has the specified ErrorCode in its details.
func IsCode(err error, code errorsv1.ErrorCode) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	for _, d := range st.Details() {
		if ei, infoOk := d.(*errdetails.ErrorInfo); infoOk && ei.GetReason() == code.String() {
			return true
		}
	}
	return false
}

package grpc

import (
	"context"
	"slices"
	"strconv"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoyauthv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/token"
	authctx "github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	"go.uber.org/zap"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type IdentityEnvoyHandler struct {
	envoyauthv3.UnimplementedAuthorizationServer

	tokenSvc token.Service
	logger   *logger.Logger
}

func NewIdentityEnvoyHandler(ts token.Service, lg *logger.Logger) envoyauthv3.AuthorizationServer {
	return &IdentityEnvoyHandler{tokenSvc: ts, logger: lg}
}

func (h *IdentityEnvoyHandler) Check(
	_ context.Context,
	req *envoyauthv3.CheckRequest,
) (*envoyauthv3.CheckResponse, error) {
	httpReq := req.GetAttributes().GetRequest().GetHttp()
	path := httpReq.GetPath()
	method := httpReq.GetMethod()
	hdrs := httpReq.GetHeaders()

	h.logger.Info("ext_authz.check",
		zap.String("method", method),
		zap.String("path", path),
		zap.Any("headers", hdrs),
	)

	publicMethods := []string{
		identityv1.IdentityPublicService_Authorize_FullMethodName,
		envoyauthv3.Authorization_Check_FullMethodName,
	}

	if slices.Contains(publicMethods, path) {
		return okResponse(nil), nil
	}

	auth := hdrs["authorization"]
	if auth == "" {
		return denyResponse(codes.Unauthenticated, "missing Authorization header"), nil
	}
	//nolint:mnd // just split header by space
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return denyResponse(codes.Unauthenticated, "invalid bearer format"), nil
	}

	claims, err := h.tokenSvc.Validate(parts[1])
	if err != nil {
		h.logger.Debug("token invalid", zap.Error(err))
		return denyResponse(codes.Unauthenticated, "invalid or expired token"), nil
	}

	h.logger.Debug("token OK", zap.Int64("telegram_user_id", claims.TelegramUserID))

	hdr := &corev3.HeaderValueOption{
		Header: &corev3.HeaderValue{
			Key:   authctx.TelegramUserIDKey.String(),
			Value: strconv.FormatInt(claims.TelegramUserID, 10),
		},
		Append: wrapperspb.Bool(false),
	}
	return okResponse([]*corev3.HeaderValueOption{hdr}), nil
}

func okResponse(hdrs []*corev3.HeaderValueOption) *envoyauthv3.CheckResponse {
	return &envoyauthv3.CheckResponse{
		Status: &rpcstatus.Status{Code: int32(codes.OK)},
		HttpResponse: &envoyauthv3.CheckResponse_OkResponse{
			OkResponse: &envoyauthv3.OkHttpResponse{Headers: hdrs},
		},
	}
}

func denyResponse(code codes.Code, msg string) *envoyauthv3.CheckResponse {
	return &envoyauthv3.CheckResponse{
		//nolint:gosec // ok
		Status: &rpcstatus.Status{Code: int32(code), Message: msg},
		HttpResponse: &envoyauthv3.CheckResponse_DeniedResponse{
			DeniedResponse: &envoyauthv3.DeniedHttpResponse{
				//nolint:gosec // ok
				Status: &typev3.HttpStatus{Code: typev3.StatusCode(code)},
			},
		},
	}
}

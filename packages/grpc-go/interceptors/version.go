package interceptors

import (
	"github.com/peterparker2005/giftduels/packages/version-go"
	"google.golang.org/grpc"
)

func VersionInterceptorUnary() grpc.UnaryServerInterceptor {
	return version.UnaryInterceptor()
}

func VersionInterceptorStream() grpc.StreamServerInterceptor {
	return version.StreamInterceptor()
}

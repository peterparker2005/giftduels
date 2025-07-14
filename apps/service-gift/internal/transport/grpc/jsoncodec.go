package grpc

import (
	// Required for ConnectRPC.
	_ "github.com/peterparker2005/giftduels/packages/grpc-go/codec/json"
	_ "google.golang.org/grpc/encoding/gzip"
)

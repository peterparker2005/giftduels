package grpc

import (
	// JSON codec is used for ConnectRPC.
	_ "github.com/peterparker2005/giftduels/packages/grpc-go/codec/json"
	_ "google.golang.org/grpc/encoding/gzip"
)

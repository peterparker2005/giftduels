package codec

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

//nolint:gochecknoinits // required for connectrpc
func init() {
	encoding.RegisterCodec(Codec{})
}

type Codec struct{}

func (Codec) Name() string {
	return "json"
}

func (codec Codec) Marshal(message interface{}) ([]byte, error) {
	if protoMessage, ok := message.(proto.Message); ok {
		return protojson.Marshal(protoMessage)
	}
	return json.Marshal(message)
}

func (codec Codec) Unmarshal(data []byte, message interface{}) error {
	if protoMessage, ok := message.(proto.Message); ok {
		return protojson.Unmarshal(data, protoMessage)
	}
	return json.Unmarshal(data, message)
}

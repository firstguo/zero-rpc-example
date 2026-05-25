package common

import (
	"fmt"

	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const JSONCodecName = "json"

type jsonCodec struct{}

var marshalOptions = protojson.MarshalOptions{
	EmitDefaultValues: true,
}

func (c *jsonCodec) Marshal(v any) ([]byte, error) {
	m, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("jsonCodec: Marshal expects proto.Message, got %T", v)
	}
	return marshalOptions.Marshal(m)
}

func (c *jsonCodec) Unmarshal(data []byte, v any) error {
	m, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("jsonCodec: Unmarshal expects proto.Message, got %T", v)
	}
	return protojson.Unmarshal(data, m)
}

func (c *jsonCodec) Name() string {
	return JSONCodecName
}

func RegisterJSONCodec() {
	encoding.RegisterCodec(&jsonCodec{})
}

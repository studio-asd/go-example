package proto

import (
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func ToJson(t *testing.T, message proto.Message) string {
	t.Helper()
	out, err := protojson.MarshalOptions{Multiline: true}.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

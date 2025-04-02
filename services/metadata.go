package services

import (
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
)

const (
	// MetadataUserAgent is being set from X-User-Agent header and only available from incoming context.
	MetadataUserAgent = "User-Agent"
	// MetadataAuthorization is being set from Authorization header and only available from incoming context.
	MetadataAuthorization = "Authorization"
	// MetatdataUserID is being set from authroization process and being set from the token information.
	MetadataUserID = "UserID"
	// MetadataUserEmail is being set from authroization process and being set from the token information.
	MetadataUserEmail = "UserEmail"
)

type GRPCMetadata struct {
	md metadata.MD
}

// SetGRPCMetadataToContext sets the kev value pairs as a gRPC metadata to the context. We are using the IncomingContext
// metadata to set the context. So, metadata.FromIncomingContext needs to be called to get the metadata from the context.
func SetGRPCMetadataToContext(ctx context.Context, kv map[string]string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(kv)
	} else {
		for k, v := range kv {
			md.Append(k, v)
		}
	}
	return metadata.NewIncomingContext(ctx, md)
}

func NewGRPCMetadataRetriever(ctx context.Context) (*GRPCMetadata, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("grpc metadata not found")
	}
	return &GRPCMetadata{
		md: md,
	}, nil
}

func (ret *GRPCMetadata) UserAgent() string {
	values := ret.md.Get(MetadataUserAgent)
	return values[0]
}

func (ret *GRPCMetadata) Authorization() string {
	values := ret.md.Get(MetadataAuthorization)
	return values[0]
}

func (ret *GRPCMetadata) UserID() string {
	values := ret.md.Get(MetadataAuthorization)
	return values[0]
}

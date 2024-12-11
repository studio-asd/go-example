package api

// GRPC is the grpc server implementation of the API. The methods in the struct should only be invoked from the
// rpc framework as interceptor and other parts of the gRPC stacks won't be available via direct method call.
type GRPC struct {
	api *API
}

func newGRPC(api *API) *GRPC {
	return &GRPC{
		api: api,
	}
}

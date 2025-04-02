package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	userv1 "github.com/studio-asd/go-example/proto/api/user/v1"
	"github.com/studio-asd/go-example/services"
	userapi "github.com/studio-asd/go-example/services/user/api"
)

type serviceAuth struct {
	userapi *userapi.API
	// noAuthMethods stores the http patterns that don't require authentication. PLEASE be careful on adding more methods
	// here as we need to make sure that the method is really doesn't require authentication.
	//
	// In practice, we should write a description of each method why it doesn't require authentication.
	noAuthPatterns map[string]string
}

func (s *serviceAuth) middleware(hf runtime.HandlerFunc) runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		// Retrieve the http pattern name from the context. Since we are using grpc-gateway, the pattern is not available via r.Pattern
		// and both runtime.HTTPPathPattern and runtime.RPCMethod will reteurn an empty value.
		// Unfortunately the HTTP pattern is not available in grpc gateway generated code so we need to type it by our own.
		ptrn, ok := runtime.HTTPPattern(r.Context())
		// This means the request is coming from non-gateway handler, so we can't handle it.
		// We will immediately return internal server error(500) in this case.
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Check whether the user is allowed to access the resource or not.
		ctx, err := s.authorize(r.Context(), ptrn.String(), r.Method, r.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Inject the context provided by the authorization to the subsequent handlers.
		r = r.WithContext(ctx)
		hf(w, r, pathParams)
	}
}

func (s *serviceAuth) authorize(ctx context.Context, httpPathPattern, reqHttpMethod, authHeader string) (context.Context, error) {
	// Check whether we can go on with the request and whether the request need further authentication or not.
	// If the pattern and method is a match then we should go without authentication.
	httpMethod, ok := s.noAuthPatterns[httpPathPattern]
	if ok && reqHttpMethod == httpMethod {
		return ctx, nil
	}

	// Get authorization header from the request to determine whether the request is authenticated or not.
	if authHeader == "" {
		return nil, errors.New("authorization header is missing")
	}
	if !strings.HasPrefix(authHeader, "Bearer") {
		return nil, errors.New("invalid type format for authorization")
	}
	authToken := strings.TrimPrefix(authHeader, "Bearer ")
	// Authenticate the user token as the session is stored within the user domain.
	resp, err := s.userapi.AuthorizeUser(ctx, &userv1.AuthorizationRequest{
		SessionToken: authToken,
	})
	if err != nil {
		return nil, err
	}
	// Set all the metadata to the incoming context.
	newCtx := services.SetGRPCMetadataToContext(ctx, map[string]string{
		services.MetadataUserID:    resp.GetUserId(),
		services.MetadataUserEmail: resp.GetEmail(),
	})
	return newCtx, nil
}

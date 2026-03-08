package grpcerr

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Translate maps downstream gRPC errors to gateway-safe errors.
// This is the single place where you decide what information reaches clients.
// Audit this function when reviewing what your API leaks.
func Translate(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error at all — network failure, context cancelled, etc.
		return status.Error(codes.Internal, "upstream service unavailable")
	}
	switch st.Code() {
	case codes.NotFound:
		return status.Error(codes.NotFound, st.Message()) // safe to forward
	case codes.AlreadyExists:
		return status.Error(codes.AlreadyExists, st.Message()) // safe to forward
	case codes.InvalidArgument:
		// Downstream rejected OUR request — that's our bug, not the client's.
		// The message reveals our internal request format — hide it.
		return status.Error(codes.Internal, "internal error")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

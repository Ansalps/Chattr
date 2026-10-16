package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func RequestIDClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {

		// 1. Get request_id from context
		reqID, ok := ctx.Value("request_id").(string)

		if ok && reqID != "" {
			// 2. Add to metadata
			md, _ := metadata.FromOutgoingContext(ctx)
			md = md.Copy()

			md.Set("x-request-id", reqID)

			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		// 3. Call gRPC
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
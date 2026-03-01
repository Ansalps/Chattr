package middleware

import (
	"context"
	"time"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/logger"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LoggerInterceptor(baseLogger logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		start := time.Now()

		// extract request_id from metadata
		md, _ := metadata.FromIncomingContext(ctx)

		var requestID string
		if ids := md.Get("x-request-id"); len(ids) > 0 {
			requestID = ids[0]
		}

		if requestID == "" {
			requestID = uuid.New().String()
		}

		reqLogger := baseLogger.With(
			logger.Field{Key: "request_id", Value: requestID},
			logger.Field{Key: "method", Value: info.FullMethod},
		)

		ctx = utils.SetLogger(ctx, reqLogger)

		resp, err := handler(ctx, req)

		latency := time.Since(start)

		if err != nil {
			reqLogger.Error("grpc request failed",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "latency", Value: latency.String()},
			)
		} else {
			reqLogger.Info("grpc request completed",
				logger.Field{Key: "latency", Value: latency.String()},
			)
		}

		return resp, err
	}
}

package middleware

import (
	"context"
	"errors"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/logger"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func LoggerInterceptor(baseLogger logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		//start := time.Now()

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

		// latency := time.Since(start)

		// if err != nil {
		// 	reqLogger.Error("grpc request failed",
		// 		logger.Field{Key: "error", Value: err},
		// 		logger.Field{Key: "latency", Value: latency.String()},
		// 	)
		// } else {
		// 	reqLogger.Info("grpc request completed",
		// 		logger.Field{Key: "latency", Value: latency.String()},
		// 	)
		// }
		log := utils.GetLogger(ctx)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrDatabaseConnectionTimeOut):
				return nil, status.Error(codes.Unavailable, "database timeout")

			case errors.Is(err, domain.ErrUserNotFound):
				return nil, status.Error(codes.NotFound, "user not found")

			case errors.Is(err, domain.ErrInvalidCredentials):
				return nil, status.Error(codes.Unauthenticated, "invalid credentials")

			case errors.Is(err, domain.ErrDatabase):
				log.Error("Database Error", logger.Field{Key: "details", Value: err.Error()})
				return nil, status.Error(codes.Internal, "internal database error")

			default:
				log.Error("Unexpected Error", logger.Field{Key: "details", Value: err.Error()})
				return nil, status.Error(codes.Internal, "internal server error")
			}
		}

		return resp, err
	}
}

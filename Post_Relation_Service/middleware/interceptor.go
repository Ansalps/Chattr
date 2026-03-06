package middleware

import (
	"context"
	"errors"

	"github.com/Ansalps/Chattr_Post_Relation_Service/infrastructure/logger"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/utils"
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
		log.Error("Client/server Error", logger.Field{Key: "details", Value: err.Error()})
		if err != nil {
			switch{
			case errors.Is(err,domain.ErrForeignKeyViolationCommentPost):
				return nil,status.Error(codes.NotFound,domain.ErrForeignKeyViolationCommentPost.Error())
			case errors.Is(err,domain.ErrRecursiveComment):
				return nil,status.Error(codes.PermissionDenied,domain.ErrRecursiveComment.Error())
			case errors.Is(err,domain.ErrCommentIdNotFound):
				return nil,status.Error(codes.NotFound,domain.ErrCommentIdNotFound.Error())
			case errors.Is(err,domain.ErrPostLikeNotFound):
				return nil,status.Error(codes.NotFound,domain.ErrPostLikeNotFound.Error())
			case errors.Is(err, domain.ErrForeignKeyViolationCommentPost):
				return nil,status.Error(codes.NotFound,domain.ErrForeignKeyViolationCommentPost.Error())
			case errors.Is(err,domain.ErrPostNotFound):
				return nil,status.Error(codes.NotFound,domain.ErrPostNotFound.Error())
			case errors.Is(err,domain.ErrUsertNotFound):
				return nil,status.Error(codes.NotFound,domain.ErrUsertNotFound.Error())
			case errors.Is(err,domain.ErrDatabase):
				return nil,status.Error(codes.Internal,domain.ErrDatabase.Error())
			}
		}

		return resp, err
	}
}

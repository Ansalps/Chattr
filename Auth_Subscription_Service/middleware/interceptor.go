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
		log.Error("Client/server Error", logger.Field{Key: "details", Value: err.Error()})
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrDatabaseConnectionTimeOut):
				return nil, status.Error(codes.Unavailable, "database timeout")

			case errors.Is(err, domain.ErrUserNotFound):
				return nil, status.Error(codes.NotFound, "user not found")

			case errors.Is(err, domain.ErrInvalidCredentials):
				return nil, status.Error(codes.Unauthenticated, "invalid credentials")
			case errors.Is(err,domain.ErrUserAlreadyExistsByEmail):
				return nil,status.Error(codes.AlreadyExists,domain.ErrUserAlreadyExistsByEmail.Error())
			case errors.Is(err,domain.ErrUserAlreadyExistsByUsername):
				return nil,status.Error(codes.AlreadyExists,domain.ErrUserAlreadyExistsByUsername.Error())
			case errors.Is(err,domain.ErrVerifyOtpTokenFail):
				return nil,status.Error(codes.Internal,domain.ErrVerifyOtpTokenFail.Error())

			case errors.Is(err,domain.ErrOtpExpired):
				return nil,status.Error(codes.FailedPrecondition,domain.ErrOtpExpired.Error())
			case errors.Is(err,domain.ErrUserAccessTokenFail):
				return nil,status.Error(codes.Internal,domain.ErrUserAccessTokenFail.Error())
			case errors.Is(err,domain.ErrUserRefreshTokenFail):
				return nil,status.Error(codes.Internal,domain.ErrUserRefreshTokenFail.Error())

			case errors.Is(err,domain.ErrAdminAccessTokenFail):
				return nil,status.Error(codes.Internal,domain.ErrAdminAccessTokenFail.Error())
			case errors.Is(err,domain.ErrAdminRefreshTokenFail):
				return nil,status.Error(codes.Internal,domain.ErrAdminRefreshTokenFail.Error())

			case errors.Is(err,domain.ErrUserNotActive):
				return nil,status.Error(codes.FailedPrecondition,domain.ErrUserNotActive.Error())

			case errors.Is(err,domain.ErrUserNotBlocked):
				return nil,status.Error(codes.FailedPrecondition,domain.ErrUserNotBlocked.Error())

			case errors.Is(err,domain.ErrBlockedLogin):
				return nil,status.Error(codes.PermissionDenied,domain.ErrBlockedLogin.Error())
			case errors.Is(err,domain.ErrPendingLogin):
				return nil,status.Error(codes.PermissionDenied,domain.ErrPendingLogin.Error())
			
			case errors.Is(err,domain.ErrInvalidRequest):
				return nil,status.Error(codes.InvalidArgument,domain.ErrInvalidRequest.Error())
			case errors.Is(err,domain.ErrExternalService):
				return nil,status.Error(codes.Internal,domain.ErrExternalService.Error())
			case errors.Is(err,domain.ErrServiceUnavailable):
				return nil,status.Error(codes.Unavailable,domain.ErrServiceUnavailable.Error())
			case errors.Is(err,domain.ErrUnknown):
				return nil,status.Error(codes.Internal,domain.ErrUnknown.Error())

			case errors.Is(err,domain.ErrSubPlanNotFound):
				return nil,status.Error(codes.NotFound,domain.ErrSubPlanNotFound.Error())
			case errors.Is(err,domain.ErrSubscriptionPlanAlreadyActive):
				return nil,status.Error(codes.FailedPrecondition,domain.ErrSubscriptionPlanAlreadyActive.Error())
			case errors.Is(err,domain.ErrInvalidResponseRazorpay):
				return nil,status.Error(codes.InvalidArgument,domain.ErrInvalidResponseRazorpay.Error())

			case errors.Is(err,domain.ErrSubscriptionPlanAlreadyDeactive):
				return nil,status.Error(codes.FailedPrecondition,domain.ErrSubscriptionPlanAlreadyDeactive.Error())

			case errors.Is(err,domain.ErrNotEligible):
				return nil,status.Error(codes.FailedPrecondition,domain.ErrNotEligible.Error())

			case errors.Is(err,domain.ErrNoActiveSubscription):
				return nil,status.Error(codes.FailedPrecondition,domain.ErrNoActiveSubscription.Error())
			case errors.Is(err,domain.ErrSubCompleted):
				return nil,status.Error(codes.FailedPrecondition,domain.ErrSubCompleted.Error())
			case errors.Is(err,domain.ErrSubCancelled):
				return nil,status.Error(codes.FailedPrecondition,domain.ErrSubCancelled.Error())
			case errors.Is(err,domain.ErrSubCancelled):
				return nil,status.Error(codes.FailedPrecondition,domain.ErrSubCancelled.Error())
			case errors.Is(err,domain.ErrRazorpayCancel):
				return nil,status.Error(codes.Internal,domain.ErrRazorpayCancel.Error())

			case errors.Is(err,domain.ErrContentTypeNil):
				return nil,status.Error(codes.Internal,domain.ErrContentTypeNil.Error())
			case errors.Is(err,domain.ErrS3UploadFail):
				return nil,status.Error(codes.Internal,domain.ErrS3UploadFail.Error())

			case errors.Is(err, domain.ErrDatabase):
				//log.Error("Database Error", logger.Field{Key: "details", Value: err.Error()})
				return nil, status.Error(codes.Internal, "internal database error")
		
			default:
				//log.Error("Unexpected Error", logger.Field{Key: "details", Value: err.Error()})
				return nil, status.Error(codes.Internal, "internal server error")
			}
		}

		return resp, err
	}
}

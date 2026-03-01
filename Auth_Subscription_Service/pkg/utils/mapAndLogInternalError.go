package utils

// import (
// 	"context"
// 	"errors"

// 	"github.com/Ansalps/Chattr_Auth_Subscription_Service/logger"
// 	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// // In a file like utils/error_mapper.go
// func MapAndLogInternalError(ctx context.Context, err error) error {
// 	log := GetLogger(ctx)

// 	// Map domain errors to gRPC status codes
// 	switch {
// 	case errors.Is(err, domain.ErrDatabaseConnectionTimeOut):
// 		return status.Error(codes.Unavailable, "database timeout")

// 	case errors.Is(err, domain.ErrUserNotFound):
// 		return status.Error(codes.NotFound, "user not found")

// 	case errors.Is(err, domain.ErrInvalidCredentials):
// 		return status.Error(codes.Unauthenticated, "invalid credentials")
// 	case errors.Is(err,domain.ErrDatabase):
// 		log.Error("Internal Server Error", logger.Field{Key: "details", Value: err.Error()})
// 		return status.Error(codes.Internal,domain.ErrDatabase.Error())

// 	// For any "Database" or unexpected errors, log them as ERROR and return Internal
// 	default:
// 		log.Error("Internal Server Error", logger.Field{Key: "details", Value: err.Error()})
// 		return status.Error(codes.Internal, "internal server error")
// 	}
// }

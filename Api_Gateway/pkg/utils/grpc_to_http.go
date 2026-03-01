package utils

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCtoHTTP(err error) (int, string) {
	//fmt.Println("error printing", err.Error())
	if st, ok := status.FromError(err); ok {
		//fmt.Println("code printing",st.Code())
		switch st.Code() {
		case codes.Internal:
			return http.StatusInternalServerError,st.Message()
		case codes.Unauthenticated:
			return http.StatusUnauthorized,st.Message()
		case codes.DeadlineExceeded:
			return http.StatusGatewayTimeout,"Gateway timed out"
		case codes.Unavailable:
			return http.StatusServiceUnavailable, "Service Unavailable"
		case codes.NotFound:
			return http.StatusNotFound, st.Message()
		case codes.FailedPrecondition:
			return http.StatusPreconditionFailed, st.Message()
		case codes.InvalidArgument:
			return http.StatusBadRequest, st.Message()
		}
	}
	return http.StatusInternalServerError, "An unexpected error occured"
}

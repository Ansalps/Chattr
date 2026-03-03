package utils

import (
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCtoHTTP(err error) (int, string) {
	//fmt.Println("error printing", err.Error())
	if st, ok := status.FromError(err); ok {
		//fmt.Println("code printing",st.Code())
		switch st.Code() {
		case codes.AlreadyExists:
			return http.StatusConflict, st.Message()
		case codes.Internal:
			return http.StatusInternalServerError, st.Message()
		case codes.Unauthenticated:
			return http.StatusUnauthorized, st.Message()
		case codes.DeadlineExceeded:
			return http.StatusGatewayTimeout, "Gateway timed out"
		case codes.Unavailable:
			return http.StatusServiceUnavailable, "Service Unavailable"
		case codes.NotFound:
			return http.StatusNotFound, st.Message()
		case codes.FailedPrecondition:
			// If the message contains "otp expired", we might want to be more specific
			if strings.Contains(st.Message(), "otp expired") {
				return http.StatusGone, st.Message() // 410
			} else if strings.Contains(st.Message(),"Cannot block user, email not verified or user alreday blocked"){
				return http.StatusConflict,st.Message()
			} else if strings.Contains(st.Message(),"Cannnot unblock user, unblock allowed for users who are alreday in blocked state"){
				return http.StatusConflict, st.Message()
			}
			return http.StatusPreconditionFailed, st.Message()
		case codes.PermissionDenied:
			return http.StatusConflict,st.Message()
		case codes.InvalidArgument:
			return http.StatusBadRequest, st.Message()
		}
	}
	return http.StatusInternalServerError, "An unexpected error occured"
}

package utils

import (
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCtoHTTP(err error) (int, string) {
	fmt.Println("error printing", err.Error())
	if st, ok := status.FromError(err); ok {
		fmt.Println("code printing",st.Code())
		switch st.Code() {
		case codes.NotFound:
			return http.StatusNotFound, st.Message()
		case codes.InvalidArgument:
			return http.StatusBadRequest, st.Message()
		}
	}
	return http.StatusInternalServerError, "An unexpected error occured"
}

package utils

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	fmt.Println(t,timestamppb.New(t))
	return timestamppb.New(t)
}

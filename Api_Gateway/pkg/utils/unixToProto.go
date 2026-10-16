package utils
import (
    "time"
    "google.golang.org/protobuf/types/known/timestamppb"
)

// UnixToProto converts an int64 Unix timestamp to a google.protobuf.Timestamp
func UnixToProto(unixTime int64) *timestamppb.Timestamp {
    if unixTime == 0 {
        return nil
    }
    // Step 1: Convert int64 to time.Time
    t := time.Unix(unixTime, 0)
    
    // Step 2: Convert time.Time to Protobuf Timestamp
    return timestamppb.New(t)
}
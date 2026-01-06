package grpc

import (
	"time"

	"google.golang.org/genproto/googleapis/type/datetime"
	"google.golang.org/protobuf/types/known/durationpb"
)

func responseDateTime(t time.Time) *datetime.DateTime {
	// v1
	_, timeOffsetSec := t.Zone()
	tz := &datetime.DateTime_UtcOffset{
		UtcOffset: &durationpb.Duration{Seconds: int64(timeOffsetSec)},
	}

	// v2
	//timeZoneName, _ := t.Zone()
	//tz := &datetime.DateTime_TimeZone{TimeZone: &datetime.TimeZone{Id: timeZoneName}}

	return &datetime.DateTime{
		Year:       int32(t.Year()),
		Month:      int32(t.Month()),
		Day:        int32(t.Day()),
		Hours:      int32(t.Hour()),
		Minutes:    int32(t.Minute()),
		Seconds:    int32(t.Second()),
		Nanos:      int32(t.Nanosecond()),
		TimeOffset: tz,
	}
}

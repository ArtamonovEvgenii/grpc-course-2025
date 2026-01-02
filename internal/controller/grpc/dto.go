package grpc

import (
	"time"

	"google.golang.org/genproto/googleapis/type/datetime"
)

func responseDateTime(t time.Time) *datetime.DateTime {
	//_, timeOffset := entityNote.UpdatedAt.Zone()

	return &datetime.DateTime{
		Year:    int32(t.Year()),
		Month:   int32(t.Month()),
		Day:     int32(t.Day()),
		Hours:   int32(t.Hour()),
		Minutes: int32(t.Minute()),
		Seconds: int32(t.Second()),
		Nanos:   int32(t.Nanosecond()),
		//TimeOffset: datetime.TimeOffset{}, // todo: set time offset
	}
}

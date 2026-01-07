package grpc

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/genproto/googleapis/type/datetime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/ArtamonovEvgenii/grpc-course-2025/internal/entity"
	pb "github.com/ArtamonovEvgenii/grpc-course-2025/pkg/api/notes/v1"
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

func domainToTransportError(grpcCode codes.Code, err error) error {
	switch {
	case errors.Is(err, entity.ErrNoteNotFound):
		respStatus := status.New(grpcCode, entity.ErrNoteNotFound.Error())

		descriptionText := err.Error()
		fmt.Println(descriptionText)

		errDetails := &pb.ErrorDetails{
			Code:        pb.ErrorCode_ERROR_CODE_NOTE_NOT_FOUND,
			Description: err.Error(),
		}

		var errAddDetails error
		respStatus, errAddDetails = respStatus.WithDetails(errDetails)
		if errAddDetails != nil {
			return errors.Join(err, errAddDetails)
		}

		return respStatus.Err()

	default:
		respStatus := status.New(grpcCode, err.Error())

		return respStatus.Err()
	}
}

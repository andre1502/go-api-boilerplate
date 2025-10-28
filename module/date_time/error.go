package date_time

import "errors"

var (
	ErrInvalidDateTimeFormat = errors.New("invalid date time format")
	ErrLoadLocation          = errors.New("error when load timezone location")
	ErrEmptyStartTime        = errors.New("start time is empty")
	ErrEmptyEndTime          = errors.New("end time is empty")
	ErrStartTimeOverEndTime  = errors.New("start time is over end time")
)

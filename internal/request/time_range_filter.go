package request

import (
	"go-api-boilerplate/internal/exception"
	"go-api-boilerplate/internal/status_code"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/date_time"
	"time"
)

type TimeRangeFilter struct {
	StartTime       *string `json:"start_time" query:"start_time"`
	EndTime         *string `json:"end_time" query:"end_time"`
	Datetime        *bool   `json:"datetime" query:"datetime"`
	TimeOfStartTime *time.Time
	TimeOfEndTime   *time.Time
}

func (tf *TimeRangeFilter) ParseTimeRangeFilter(withDefault bool) error {
	var startTime time.Time
	var endTime time.Time
	var today *time.Time
	var err error
	format := time.DateOnly

	if tf.Datetime != nil && *tf.Datetime {
		format = time.DateTime
	}

	if tf.StartTime != nil && tf.EndTime != nil {
		startTime, err = time.ParseInLocation(format, *tf.StartTime, time.Local)
		if err != nil {
			return exception.Ex.Errors(status_code.INVALID_DATE_TIME_FORMAT_CODE, status_code.INVALID_DATE_TIME_FORMAT_MESSAGE, err)
		}

		endTime, err = time.ParseInLocation(format, *tf.EndTime, time.Local)
		if err != nil {
			return exception.Ex.Errors(status_code.INVALID_DATE_TIME_FORMAT_CODE, status_code.INVALID_DATE_TIME_FORMAT_MESSAGE, err)
		}
	} else if withDefault {
		startTime = time.Now().AddDate(0, 0, -7)
		endTime = time.Now()
	} else {
		return exception.Ex.Errors(status_code.START_TIME_OR_END_TIME_EMPTY_ERROR_CODE, status_code.START_TIME_OR_END_TIME_EMPTY_ERROR_MESSAGE, nil)
	}

	if withDefault || (tf.StartTime != nil && tf.EndTime != nil && !module.IsEmptyString(*tf.StartTime) && !module.IsEmptyString(*tf.EndTime)) {
		if startTime.After(endTime) {
			return exception.Ex.Errors(status_code.START_TIME_GREATER_THAN_END_TIME_ERROR_CODE, status_code.START_TIME_GREATER_THAN_END_TIME_ERROR_MESSAGE, nil)
		}

		now := time.Now().Add(time.Hour * 24)
		today, err = date_time.GetDateOnly(&now)
		if err != nil {
			return err
		}

		if startTime.After(*today) || endTime.After(*today) {
			return exception.Ex.Errors(status_code.TIME_FILTER_GREATER_THAN_TODAY_ERROR_CODE, status_code.TIME_FILTER_GREATER_THAN_TODAY_ERROR_MESSAGE, nil)
		}

		if tf.Datetime != nil && *tf.Datetime {
			endTime = endTime.Add(time.Second)
		} else {
			endTime = endTime.Add(time.Hour * 24)
		}

		startTimeStr := startTime.Format(format)
		endTimeStr := endTime.Format(format)

		tf.StartTime = &startTimeStr
		tf.EndTime = &endTimeStr
		tf.TimeOfStartTime = &startTime
		tf.TimeOfEndTime = &endTime
	}

	return nil
}

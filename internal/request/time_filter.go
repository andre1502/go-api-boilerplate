package request

import (
	"go-api-boilerplate/internal/exception"
	"go-api-boilerplate/internal/status_code"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/date_time"
	"time"
)

type TimeFilter struct {
	Time       *string `json:"time" query:"time"`
	TimeOfTime *time.Time
	Datetime   *bool
}

func (tf *TimeFilter) ParseTimeFilter(withDefault bool) error {
	var timeF time.Time
	var err error
	format := time.DateOnly

	if tf.Datetime != nil && *tf.Datetime {
		format = time.DateTime
	}

	if tf.Time != nil {
		timeF, err = time.ParseInLocation(format, *tf.Time, time.Local)
		if err != nil {
			return exception.Ex.Errors(status_code.INVALID_DATE_TIME_FORMAT_CODE, status_code.INVALID_DATE_TIME_FORMAT_MESSAGE, err)
		}
	} else if withDefault {
		timeF = time.Now()
	} else {
		return exception.Ex.Errors(status_code.TIME_FILTER_EMPTY_ERROR_CODE, status_code.TIME_FILTER_EMPTY_ERROR_MESSAGE, nil)
	}

	now := time.Now().Add(time.Hour * 24)
	today, err := date_time.GetDateOnly(&now)
	if err != nil {
		return err
	}

	if timeF.After(*today) {
		return exception.Ex.Errors(status_code.TIME_FILTER_GREATER_THAN_TODAY_ERROR_CODE, status_code.TIME_FILTER_GREATER_THAN_TODAY_ERROR_MESSAGE, nil)
	}

	if withDefault || (tf.Time != nil && !module.IsEmptyString(*tf.Time)) {
		if tf.Datetime != nil && *tf.Datetime {
			timeF = timeF.Add(time.Second * 1)
		} else {
			timeF = timeF.Add(time.Hour * 24)
		}

		timeFStr := timeF.Format(format)

		tf.Time = &timeFStr
		tf.TimeOfTime = &timeF
	}

	return nil
}

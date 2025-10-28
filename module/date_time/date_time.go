package date_time

import (
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/logger"
	"math"
	"time"
)

func ConvertToTimezone(dateTimeStr string, destTimezone string, dateTimeFormat string) (*time.Time, error) {
	if module.IsEmptyString(dateTimeFormat) {
		dateTimeFormat = time.DateTime
	}

	if module.IsEmptyString(dateTimeStr) {
		dateTimeStr = time.Now().UTC().Format(dateTimeFormat)
	}

	utcTime, err := time.ParseInLocation(dateTimeFormat, dateTimeStr, time.UTC)
	if err != nil {
		logger.Log.Errorf("error parsing datetime: %v", err)
		return nil, ErrInvalidDateTimeFormat
	}

	if !module.IsEmptyString(destTimezone) {
		destLoc, err := time.LoadLocation(destTimezone)
		if err != nil {
			logger.Log.Errorf("error loading destination timezone: %v", err)
			return nil, ErrLoadLocation
		}

		destTime := utcTime.In(destLoc)

		return &destTime, nil
	}

	destTime := utcTime.In(time.Local)

	return &destTime, nil
}

func DaysBetween(a, b time.Time, includeLastDay bool) int {
	if a.After(b) {
		a, b = b, a
	}

	days := (int)(math.Ceil(b.Sub(a).Hours() / 24.0))

	if includeLastDay {
		return days + 1
	}

	return days
}

func ValidateStartTimeAfterEndTime(startTimeStr, endTimeStr, formatTime string) error {
	var startTime time.Time
	var endTime time.Time
	var err error

	if module.IsEmptyString(formatTime) {
		formatTime = time.DateOnly
	}

	if module.IsEmptyString(endTimeStr) {
		return ErrEmptyStartTime
	}

	if module.IsEmptyString(endTimeStr) {
		return ErrEmptyEndTime
	}

	startTime, err = time.ParseInLocation(formatTime, startTimeStr, time.Local)
	if err != nil {
		logger.Log.Errorf("error parsing start time: %v", err)
		return ErrInvalidDateTimeFormat
	}

	endTime, err = time.ParseInLocation(formatTime, endTimeStr, time.Local)
	if err != nil {
		logger.Log.Errorf("error parsing end time: %v", err)
		return ErrInvalidDateTimeFormat
	}

	if startTime.After(endTime) {
		return ErrStartTimeOverEndTime
	}

	return nil
}

func GetDateOnly(currentTime *time.Time) (*time.Time, error) {
	if currentTime == nil {
		nowTime := time.Now()
		currentTime = &nowTime
	}

	currentDate, err := time.ParseInLocation(time.DateOnly, currentTime.Format(time.DateOnly), time.Local)
	if err != nil {
		logger.Log.Errorf("error parsing date: %v", err)
		return nil, ErrInvalidDateTimeFormat
	}

	return &currentDate, nil
}

// GetMondayOfCurrentWeek returns the Monday of the week for the given time.
// This effectively acts as "get the start day of the week" for a Monday-based week.
func GetMondayOfCurrentWeek(t time.Time) time.Time {
	// Normalize the time to the start of its day (00:00:00) to ensure consistency.
	// This zeroes out hours, minutes, seconds, and nanoseconds.
	normalizedTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	// Calculate how many days to subtract to get to Monday.
	// time.Weekday() returns Sunday=0, Monday=1, ..., Saturday=6.
	// We want to move back to Monday (1).
	// Example:
	// If it's Monday (1): 1 - 1 = 0 days back.
	// If it's Tuesday (2): 2 - 1 = 1 day back.
	// If it's Sunday (0): 0 - 1 = -1. This needs special handling.
	daysToMonday := int(normalizedTime.Weekday() - time.Monday)

	// If the calculation results in a negative number (meaning the current day
	// is chronologically "before" Monday in the Weekday enum, like Sunday),
	// add 7 to wrap around to the previous week's Monday.
	// Example: Sunday (0) - Monday (1) = -1. Adding 7 makes it 6.
	// So from Sunday, we subtract 6 days to get the previous Monday.
	if daysToMonday < 0 {
		daysToMonday += 7
	}

	// Subtract the calculated days from the normalized time.
	monday := normalizedTime.AddDate(0, 0, -daysToMonday)

	return monday
}

func GetStartOfMonth() time.Time {
	now := time.Now()

	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
}

func GetStartOfDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

func GetEndOfDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
}

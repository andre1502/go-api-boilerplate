package db

import (
	"errors"
)

var (
	ErrEmptyConfig     = errors.New("empty database config")
	ErrOpenConnection  = errors.New("error when open database connection")
	ErrScanTotalRecord = errors.New("error when get total record")
)

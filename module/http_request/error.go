package http_request

import "errors"

var (
	ErrNewRequest     = errors.New("error when create new request")
	ErrDoRequest      = errors.New("error when do request")
	ErrIOReadResponse = errors.New("error on io read response")
)

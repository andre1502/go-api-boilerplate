package geolocation

import "errors"

var (
	ErrGeoLocationClientNotInit = errors.New("geolocation client is not initialized")
	ErrInvalidIPAddress         = errors.New("invalid IP Address")
	ErrLookupIPAddress          = errors.New("lookup error for IP Address")
)

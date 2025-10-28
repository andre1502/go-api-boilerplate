package status_code

const (
	NEW_REQUEST_ERROR_CODE    = 6000
	NEW_REQUEST_ERROR_MESSAGE = "Error when create new request"

	DO_REQUEST_ERROR_CODE    = 6001
	DO_REQUEST_ERROR_MESSAGE = "Error when do request"

	IO_READ_RESPONSE_ERROR_CODE    = 6002
	IO_READ_RESPONSE_ERROR_MESSAGE = "Error on io read response"

	HTTP_REQUEST_ERROR_CODE    = 6999
	HTTP_REQUEST_ERROR_MESSAGE = "Http request error"
)

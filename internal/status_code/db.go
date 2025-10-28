package status_code

const (
	EMPTY_DB_CONFIG_ERROR_CODE    = 1000
	EMPTY_DB_CONFIG_ERROR_MESSAGE = "Empty database config"

	OPEN_DB_CONNECTION_ERROR_CODE    = 1001
	OPEN_DB_CONNECTION_ERROR_MESSAGE = "Error when open database connection"

	GET_TOTAL_RECORD_ERROR_CODE    = 1002
	GET_TOTAL_RECORD_ERROR_MESSAGE = "Error when get total record"

	DB_ERROR_CODE    = 1999
	DB_ERROR_MESSAGE = "DB error"
)

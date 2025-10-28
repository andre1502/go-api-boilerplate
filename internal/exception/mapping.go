package exception

import (
	"errors"
	"go-api-boilerplate/internal/status_code"
	"go-api-boilerplate/module/cryptography"
	"go-api-boilerplate/module/date_time"
	"go-api-boilerplate/module/db"
	"go-api-boilerplate/module/elastic"
	"go-api-boilerplate/module/http_request"
	"go-api-boilerplate/module/redis"
)

func (ex *Exception) MappingErrorCryptography(err error) *Exception {
	if errors.Is(err, cryptography.ErrNewCipher) {
		return ex.Errors(status_code.NEW_CIPHER_ERROR_CODE, status_code.NEW_CIPHER_ERROR_MESSAGE, err)
	}

	if errors.Is(err, cryptography.ErrInvalidPaddingBlockSize) {
		return ex.Errors(status_code.INVALID_PADDING_BLOCK_SIZE_ERROR_CODE, status_code.INVALID_PADDING_BLOCK_SIZE_ERROR_MESSAGE, err)
	}

	if errors.Is(err, cryptography.ErrDecryption) {
		return ex.Errors(status_code.DECRYPTION_ERROR_CODE, status_code.DECRYPTION_ERROR_MESSAGE, err)
	}

	if errors.Is(err, cryptography.ErrHashPassword) {
		return ex.Errors(status_code.HASH_PASSWORD_ERROR_CODE, status_code.HASH_PASSWORD_ERROR_MESSAGE, err)
	}

	return ex.Errors(status_code.CRYPTOGRAPHY_ERROR_CODE, status_code.CRYPTOGRAPHY_ERROR_MESSAGE, err)
}

func (ex *Exception) MappingErrorDateTime(err error) *Exception {
	if errors.Is(err, date_time.ErrInvalidDateTimeFormat) {
		return ex.Errors(status_code.INVALID_DATE_TIME_FORMAT_CODE, status_code.INVALID_DATE_TIME_FORMAT_MESSAGE, err)
	}

	if errors.Is(err, date_time.ErrLoadLocation) {
		return ex.Errors(status_code.LOAD_TIMEZONE_LOCATION_ERROR_CODE, status_code.LOAD_TIMEZONE_LOCATION_ERROR_MESSAGE, err)
	}

	if errors.Is(err, date_time.ErrEmptyStartTime) || errors.Is(err, date_time.ErrEmptyEndTime) {
		return ex.Errors(status_code.START_TIME_OR_END_TIME_EMPTY_ERROR_CODE, status_code.START_TIME_OR_END_TIME_EMPTY_ERROR_MESSAGE, err)
	}

	if errors.Is(err, date_time.ErrStartTimeOverEndTime) {
		return ex.Errors(status_code.START_TIME_GREATER_THAN_END_TIME_ERROR_CODE, status_code.START_TIME_GREATER_THAN_END_TIME_ERROR_MESSAGE, err)
	}

	return ex.Errors(status_code.DATE_TIME_ERROR_CODE, status_code.DATE_TIME_ERROR_MESSAGE, err)
}

func (ex *Exception) MappingErrorDB(err error) *Exception {
	if errors.Is(err, db.ErrEmptyConfig) {
		return ex.Errors(status_code.EMPTY_DB_CONFIG_ERROR_CODE, status_code.EMPTY_DB_CONFIG_ERROR_MESSAGE, err)
	}

	if errors.Is(err, db.ErrOpenConnection) {
		return ex.Errors(status_code.OPEN_DB_CONNECTION_ERROR_CODE, status_code.OPEN_DB_CONNECTION_ERROR_MESSAGE, err)
	}

	if errors.Is(err, db.ErrScanTotalRecord) {
		return ex.Errors(status_code.GET_TOTAL_RECORD_ERROR_CODE, status_code.GET_TOTAL_RECORD_ERROR_MESSAGE, err)
	}

	return ex.Errors(status_code.DB_ERROR_CODE, status_code.DB_ERROR_MESSAGE, err)
}

func (ex *Exception) MappingErrorElastic(err error) *Exception {
	if errors.Is(err, elastic.ErrGetIndexSetting) {
		return ex.Errors(status_code.GET_INDEX_SETTING_ERROR_CODE, status_code.GET_INDEX_SETTING_ERROR_MESSAGE, err)
	}

	if errors.Is(err, elastic.ErrSearchIndex) {
		return ex.Errors(status_code.SEARCH_INDEX_ERROR_CODE, status_code.SEARCH_INDEX_ERROR_MESSAGE, err)
	}

	if errors.Is(err, elastic.ErrCastDataType) {
		return ex.Errors(status_code.CAST_DATA_TYPE_ERROR_CODE, status_code.CAST_DATA_TYPE_ERROR_MESSAGE, err)
	}

	return ex.Errors(status_code.ELASTIC_ERROR_CODE, status_code.ELASTIC_ERROR_MESSAGE, err)
}

func (ex *Exception) MappingErrorHttpRequest(err error) *Exception {
	if errors.Is(err, http_request.ErrNewRequest) {
		return ex.Errors(status_code.NEW_REQUEST_ERROR_CODE, status_code.NEW_REQUEST_ERROR_MESSAGE, err)
	}

	if errors.Is(err, http_request.ErrDoRequest) {
		return ex.Errors(status_code.DO_REQUEST_ERROR_CODE, status_code.DO_REQUEST_ERROR_MESSAGE, err)
	}

	if errors.Is(err, http_request.ErrIOReadResponse) {
		return ex.Errors(status_code.IO_READ_RESPONSE_ERROR_CODE, status_code.IO_READ_RESPONSE_ERROR_MESSAGE, err)
	}

	return ex.Errors(status_code.HTTP_REQUEST_ERROR_CODE, status_code.HTTP_REQUEST_ERROR_MESSAGE, err)
}

func (ex *Exception) MappingErrorRedis(err error) *Exception {
	if errors.Is(err, redis.ErrEmptyConfig) {
		return ex.Errors(status_code.EMPTY_REDIS_CONFIG_ERROR_CODE, status_code.EMPTY_REDIS_CONFIG_ERROR_MESSAGE, err)
	}

	if errors.Is(err, redis.ErrInvalidDB) {
		return ex.Errors(status_code.INVALID_REDIS_DB_ERROR_CODE, status_code.INVALID_REDIS_DB_ERROR_MESSAGE, err)
	}

	if errors.Is(err, redis.ErrInvalidConnectionDB) {
		return ex.Errors(status_code.REDIS_CONNECTION_ERROR_CODE, status_code.REDIS_CONNECTION_ERROR_MESSAGE, err)
	}

	if errors.Is(err, redis.ErrSetNX) {
		return ex.Errors(status_code.SET_NX_ERROR_CODE, status_code.SET_NX_ERROR_MESSAGE, err)
	}

	if errors.Is(err, redis.ErrSetMNX) {
		return ex.Errors(status_code.SET_MNX_ERROR_CODE, status_code.SET_MNX_ERROR_MESSAGE, err)
	}

	if errors.Is(err, redis.ErrSetCache) {
		return ex.Errors(status_code.SET_CACHE_ERROR_CODE, status_code.SET_CACHE_ERROR_MESSAGE, err)
	}

	if errors.Is(err, redis.ErrGetCache) {
		return ex.Errors(status_code.GET_CACHE_ERROR_CODE, status_code.GET_CACHE_ERROR_MESSAGE, err)
	}

	if errors.Is(err, redis.ErrDelCacheEmptyKeys) {
		return ex.Errors(status_code.DELETE_CACHE_EMPTY_KEYS_ERROR_CODE, status_code.DELETE_CACHE_EMPTY_KEYS_ERROR_MESSAGE, err)
	}

	if errors.Is(err, redis.ErrDelCache) {
		return ex.Errors(status_code.DELETE_CACHE_ERROR_CODE, status_code.DELETE_CACHE_ERROR_MESSAGE, err)
	}

	if errors.Is(err, redis.ErrFetchFunction) {
		return ex.Errors(status_code.CALLBACK_FUNCTION_ERROR_CODE, status_code.CALLBACK_FUNCTION_MESSAGE, err)
	}

	if errors.Is(err, redis.ErrIncrementCache) {
		return ex.Errors(status_code.INCREMENT_CACHE_ERROR_CODE, status_code.INCREMENT_CACHE_ERROR_MESSAGE, err)
	}

	if errors.Is(err, redis.ErrDecrementCache) {
		return ex.Errors(status_code.DECREMENT_CACHE_ERROR_CODE, status_code.DECREMENT_CACHE_ERROR_MESSAGE, err)
	}

	return ex.Errors(status_code.REDIS_ERROR_CODE, status_code.REDIS_ERROR_MESSAGE, err)
}

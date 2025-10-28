package redis

import "errors"

var (
	ErrEmptyConfig         = errors.New("empty redis config")
	ErrInvalidDB           = errors.New("invalid redis db")
	ErrConnectionFailed    = errors.New("connection redis failed")
	ErrInvalidConnectionDB = errors.New("invalid connection redis db")
	ErrSetNX               = errors.New("error when set nx cache")
	ErrSetMNX              = errors.New("error when set multiple nx cache")
	ErrSetCache            = errors.New("error when set cache")
	ErrGetCache            = errors.New("error when get cache")
	ErrDelCacheEmptyKeys   = errors.New("keys required to remove cache")
	ErrDelCache            = errors.New("error when remove cache")
	ErrFetchFunction       = errors.New("error when call callback function")
	ErrIncrementCache      = errors.New("error when increment cache")
	ErrDecrementCache      = errors.New("error when decrement cache")
)

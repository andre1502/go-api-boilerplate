package elastic

import "errors"

var (
	ErrGetIndexSetting = errors.New("error on get index setting")
	ErrSearchIndex     = errors.New("error when search index")
	ErrCastDataType    = errors.New("error when cast data type")
)

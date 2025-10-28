package cryptography

import "errors"

var (
	ErrNewCipher               = errors.New("error when init new cipher")
	ErrInvalidPaddingBlockSize = errors.New("invalid padding block size")
	ErrDecryption              = errors.New("error when decrypt text")

	ErrHashPassword = errors.New("error when hash password")
)

package cryptography

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"go-api-boilerplate/module"
)

type Encryption struct {
	SKey   string
	Key    []byte
	IV     []byte
	Chiper cipher.Block
}

func NewEncryption(key string) (*Encryption, error) {
	bkey := make([]byte, 32)
	copy(bkey, []byte(key))

	iv := []byte(module.Reverse(key))

	cipher, err := aes.NewCipher(bkey)
	if err != nil {
		return nil, ErrNewCipher
	}

	return &Encryption{
		SKey:   key,
		Key:    bkey,
		IV:     iv,
		Chiper: cipher,
	}, nil
}

func (e *Encryption) pad(plainText []byte, blockSize int) ([]byte, error) {
	padding := blockSize - len(plainText)%blockSize
	padded := bytes.Repeat([]byte{byte(padding)}, padding)

	plainText = append(plainText, padded...)

	if len(plainText)%aes.BlockSize != 0 {
		return nil, ErrInvalidPaddingBlockSize
	}

	return plainText, nil
}

func (e *Encryption) unpad(padded []byte, blockSize int) ([]byte, error) {
	if len(padded)%blockSize != 0 {
		return nil, ErrInvalidPaddingBlockSize
	}

	bufLen := len(padded) - int(padded[len(padded)-1])
	buf := make([]byte, bufLen)
	copy(buf, padded[:bufLen])

	return buf, nil
}

func (e *Encryption) Encrypt(plainText string) (string, error) {
	buffer, err := e.pad([]byte(plainText), aes.BlockSize)
	if err != nil {
		return plainText, err
	}

	mode := cipher.NewCBCEncrypter(e.Chiper, e.IV)
	mode.CryptBlocks(buffer, buffer)

	return base64.StdEncoding.EncodeToString(buffer), nil
}

func (e *Encryption) Decrypt(encText string) (string, error) {
	buffer, err := base64.StdEncoding.DecodeString(encText)
	if err != nil {
		return encText, ErrDecryption
	}

	mode := cipher.NewCBCDecrypter(e.Chiper, e.IV)
	mode.CryptBlocks(buffer, buffer)

	buffer, err = e.unpad(buffer, aes.BlockSize)
	if err != nil {
		return encText, err
	}

	return string(buffer), nil
}

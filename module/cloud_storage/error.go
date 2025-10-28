package cloud_storage

import "errors"

var (
	ErrCloudStorageClientNotInit   = errors.New("cloud storage client is not initialized")
	ErrCloudStorageBucketNameEmpty = errors.New("empty cloud storage bucket name, please check .env setup")
	ErrOpenUploadedFile            = errors.New("error opening uploaded file")
	ErrFolderName                  = errors.New("invalid folder name: '..' is not allowed")
	ErrFileSizeTooLarge            = errors.New("file size too large")
	ErrNoFileSelected              = errors.New("no files selected for upload")
	ErrTooManyFileSelected         = errors.New("too many files selected")
	ErrInvalidMimeType             = errors.New("invalid mime type")
	ErrReadObjectFromBucket        = errors.New("error when read object from bucket")
	ErrDeleteObjectFromBucket      = errors.New("error when delete object from bucket")
	ErrIORead                      = errors.New("error on io read")
	ErrIOCopy                      = errors.New("error on io copy")
	ErrWriterWriteToBucket         = errors.New("error on writer write to bucket")
	ErrWriterClose                 = errors.New("error on writer close")
	ErrUnmarshal                   = errors.New("error json unmarshal")
	ErrMarshalIndent               = errors.New("error marshal indent")
)

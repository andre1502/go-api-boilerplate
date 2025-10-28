package cloud_storage

import (
	"context"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/redis"
)

type CloudStorage struct {
	ctx                   context.Context
	config                *config.Config
	redis                 *redis.RedisConnection
	maxFileSizeBytes      int64
	maxFileCount          int
	AllowedMimeTypes      map[string]bool
	AllowedImageMimeTypes map[string]bool
}

type UploadFile struct {
	Filename string `json:"filename"`
	Error    error  `json:"error,omitempty"`
}

func NewCloudStorage(
	cfg *config.Config,
	rds *redis.RedisConnection,
) *CloudStorage {
	maxFileSizeMB := module.StrToIntDefault(cfg.MAX_FILE_SIZE_MB, 5)
	maxFileCount := module.StrToIntDefault(cfg.MAX_FILE_COUNT, 10)

	AllowedMimeTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"text/plain":      true, // Can cover some CSVs, but generic
		"text/csv":        true, // The preferred and official MIME type for CSV
		"application/csv": true, // Also commonly seen for CSV
	}

	AllowedImageMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	return &CloudStorage{
		ctx:                   context.Background(),
		config:                cfg,
		redis:                 rds,
		maxFileSizeBytes:      int64(maxFileSizeMB) * 1024 * 1024,
		maxFileCount:          maxFileCount,
		AllowedMimeTypes:      AllowedMimeTypes,
		AllowedImageMimeTypes: AllowedImageMimeTypes,
	}
}

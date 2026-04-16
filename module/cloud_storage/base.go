package cloud_storage

import (
	"context"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/redis"
	"regexp"
	"strings"
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
	maxFileSizeMB := module.StrToIntDefault(cfg.MAX_FILE_SIZE_MB, 2)
	maxFileCount := module.StrToIntDefault(cfg.MAX_FILE_COUNT, 10)

	AllowedMimeTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/webp":      true,
		"text/plain":      true, // Can cover some CSVs, but generic
		"text/csv":        true, // The preferred and official MIME type for CSV
		"application/csv": true, // Also commonly seen for CSV
	}

	AllowedImageMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
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

func (cs *CloudStorage) cleanFolderName(folderName string) (string, error) {
	if module.IsEmptyString(folderName) {
		return "", nil
	}

	// Replace any backslashes with forward slashes for compatibility
	folderName = strings.ReplaceAll(folderName, "\\", "/")
	// Remove any leading/trailing slashes
	folderName = strings.Trim(folderName, "/")
	// Replace multiple slashes with a single slash
	folderName = regexp.MustCompile(`/{2,}`).ReplaceAllString(folderName, "/")

	// Optional: Further restrict characters if needed. allows many chars,
	// but for user-inputted paths, it's often safer to restrict.
	// For example, to allow only alphanumeric, hyphens, underscores, and forward slashes:
	// folderName = regexp.MustCompile(`[^a-zA-Z0-9/\-_.]`).ReplaceAllString(folderName, "")

	// Prevent ".." for directory traversal attempts
	if strings.Contains(folderName, "..") {
		return "", ErrFolderName
	}

	if !module.IsEmptyString(folderName) {
		folderName += "/"
	}

	return folderName, nil
}

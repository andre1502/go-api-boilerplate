package cloud_storage

import (
	"context"
	"encoding/json"
	"fmt"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/logger"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GoogleClient struct {
	*CloudStorage
	client *storage.Client
}

func NewGoogleClient(cloudStorage *CloudStorage) (*GoogleClient, error) {
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if module.IsEmptyString(credentialsPath) {
		msg := "service account GOOGLE_APPLICATION_CREDENTIALS environment variable is not set"
		fmt.Println(msg)
		logger.Log.Warn(msg)

		return nil, nil
	}

	client, err := storage.NewClient(cloudStorage.ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		msg := "error when init google cloud storage with credentials file: '%s'. %v"
		fmt.Println(fmt.Errorf(msg, credentialsPath, err))
		logger.Log.Errorf(msg, credentialsPath, err)

		return nil, err
	}

	return &GoogleClient{
		CloudStorage: cloudStorage,
		client:       client,
	}, nil
}

func (g *GoogleClient) cleanFolderName(folderName string) (string, error) {
	if module.IsEmptyString(folderName) {
		return "", nil
	}

	// Replace any backslashes with forward slashes for GCS compatibility
	folderName = strings.ReplaceAll(folderName, "\\", "/")
	// Remove any leading/trailing slashes
	folderName = strings.Trim(folderName, "/")
	// Replace multiple slashes with a single slash
	folderName = regexp.MustCompile(`/{2,}`).ReplaceAllString(folderName, "/")

	// Optional: Further restrict characters if needed. GCS allows many chars,
	// but for user-inputted paths, it's often safer to restrict.
	// For example, to allow only alphanumeric, hyphens, underscores, and forward slashes:
	// folderName = regexp.MustCompile(`[^a-zA-Z0-9/\-_.]`).ReplaceAllString(folderName, "")

	// Prevent ".." for directory traversal attempts (GCS handles this, but good practice)
	if strings.Contains(folderName, "..") {
		return "", ErrFolderName
	}

	if !module.IsEmptyString(folderName) {
		folderName += "/"
	}

	return folderName, nil
}

// readAndUpdateJSON reads a JSON file from GCS, updates a specified field, and
// re-uploads it to the same bucket.
func (g *GoogleClient) ReadAndUpdateJSONFile(ctx context.Context, bucketName, objectName string, updateKey, updateValue string) (bool, error) {
	if module.IsEmptyString(bucketName) {
		return false, ErrCloudStorageBucketNameEmpty
	}

	if g.client == nil {
		return false, ErrCloudStorageClientNotInit
	}

	// Get a reference to the bucket.
	bucket := g.client.Bucket(bucketName)

	// Create a new reader to read the object's content.
	obj := bucket.Object(objectName)
	rc, err := obj.NewReader(ctx)
	if err != nil {
		logger.Log.Errorf("Object(%q).NewReader: %v", objectName, err)
		return false, ErrReadObjectFromBucket
	}

	defer rc.Close()

	// Read the content into a byte slice.
	data, err := io.ReadAll(rc)
	if err != nil {
		logger.Log.Errorf("io.ReadAll: %v", err)
		return false, ErrIORead
	}

	// Parse the JSON data into a map. Using a map allows us to handle
	// dynamic JSON structures without defining a specific struct.
	var parsedJSON map[string]interface{}
	if err := json.Unmarshal(data, &parsedJSON); err != nil {
		logger.Log.Errorf("json.Unmarshal: %v", err)
		return false, ErrUnmarshal
	}

	// Update the specific field.
	parsedJSON[updateKey] = updateValue

	// Marshal the updated map back into a byte slice.
	updatedData, err := json.MarshalIndent(parsedJSON, "", "  ")
	if err != nil {
		logger.Log.Errorf("json.MarshalIndent: %v", err)
		return false, ErrMarshalIndent
	}

	// Re-upload the updated content.
	wc := obj.NewWriter(ctx)
	if _, err = wc.Write(updatedData); err != nil {
		logger.Log.Errorf("Writer.Write: %v", err)
		return false, ErrWriterWriteToBucket
	}
	if err := wc.Close(); err != nil {
		logger.Log.Errorf("Error writer close %s: %v", objectName, err)

		return false, ErrWriterClose
	}

	return true, nil
}

func (g *GoogleClient) UploadRawContent(ctx context.Context, bucketName, objectName, content string) (string, error) {
	if module.IsEmptyString(bucketName) {
		return "", ErrCloudStorageBucketNameEmpty
	}

	if g.client == nil {
		return "", ErrCloudStorageClientNotInit
	}

	// Create a GCS object writer
	// This writer is connected directly to GCS. As you write to it, data streams to GCS.
	writer := g.client.Bucket(bucketName).Object(objectName).NewWriter(ctx) // Use request context for GCS operation

	// Write the content from the in-memory string
	if _, err := io.Copy(writer, strings.NewReader(content)); err != nil {
		logger.Log.Errorf("Error io.Copy %s: %v", objectName, err)

		// Try to close the writer even on error to release resources
		if closeErr := writer.Close(); closeErr != nil {
			logger.Log.Errorf("Error closing GCS writer after copy error for '%s': %v", objectName, closeErr)
		}

		return "", ErrIOCopy
	}

	// Close the GCS writer to finalize the upload
	if err := writer.Close(); err != nil {
		logger.Log.Errorf("Error writer close %s: %v", objectName, err)

		return "", ErrWriterClose
	}

	return objectName, nil
}

func (g *GoogleClient) UploadFileWeb(ctx context.Context, folderName string, formFile *multipart.FileHeader) (string, error) {
	if module.IsEmptyString(g.config.GCP_CLOUD_STORAGE_BUCKET_NAME) {
		return "", ErrCloudStorageBucketNameEmpty
	}

	if g.client == nil {
		return "", ErrCloudStorageClientNotInit
	}

	folderName, err := g.cleanFolderName(folderName)
	if err != nil {
		return "", err
	}

	if formFile.Size > g.maxFileSizeBytes {
		return "", ErrFileSizeTooLarge
	}

	contentType := formFile.Header.Get("Content-Type")
	if !g.AllowedMimeTypes[contentType] {
		return "", ErrInvalidMimeType
	}

	// Open the uploaded file (it's in memory or a temp file managed by Echo)
	srcFile, err := formFile.Open()
	if err != nil {
		return "", ErrOpenUploadedFile
	}

	defer srcFile.Close() // Ensure the source file is closed

	// Define the destination object name in GCS
	// We'll use a timestamped filename to avoid clashes
	timestamp := time.Now().UnixMicro()
	objectName := fmt.Sprintf("%s%d_%s", folderName, timestamp, filepath.Base(formFile.Filename))

	// Create a GCS object writer
	// This writer is connected directly to GCS. As you write to it, data streams to GCS.
	writer := g.client.Bucket(g.config.GCP_CLOUD_STORAGE_BUCKET_NAME).Object(objectName).NewWriter(ctx) // Use request context for GCS operation

	// Optional: Set content type if known or desired (Echo provides file.Header.Get("Content-Type"))
	writer.ContentType = formFile.Header.Get("Content-Type")

	// Stream the file content directly from the uploaded file to GCS
	if _, err = io.Copy(writer, srcFile); err != nil {
		logger.Log.Errorf("Error io.Copy %s: %v", formFile.Filename, err)

		// Try to close the writer even on error to release resources
		if closeErr := writer.Close(); closeErr != nil {
			logger.Log.Errorf("Error closing GCS writer after copy error for '%s': %v", formFile.Filename, closeErr)
		}

		return "", ErrIOCopy
	}

	// Close the GCS writer to finalize the upload
	if err := writer.Close(); err != nil {
		logger.Log.Errorf("Error writer close %s: %v", formFile.Filename, err)

		return "", ErrWriterClose
	}

	return objectName, nil
}

func (g *GoogleClient) UploadMultipleFileWeb(ctx context.Context, folderName string, forms *multipart.Form) ([]UploadFile, error) {
	results := []UploadFile{}

	if module.IsEmptyString(g.config.GCP_CLOUD_STORAGE_BUCKET_NAME) {
		return results, ErrCloudStorageBucketNameEmpty
	}

	if g.client == nil {
		return results, ErrCloudStorageClientNotInit
	}

	folderName, err := g.cleanFolderName(folderName)
	if err != nil {
		return results, err
	}

	formFiles := forms.File["files"]

	if len(formFiles) == 0 {
		return results, ErrNoFileSelected
	}

	if len(formFiles) > g.maxFileCount {
		return results, ErrTooManyFileSelected
	}

	var wg sync.WaitGroup                                // WaitGroup to wait for all goroutines
	resultsChan := make(chan UploadFile, len(formFiles)) // Channel to collect results from goroutines

	// Iterate over each file and launch a goroutine for upload
	for _, fileHeader := range formFiles {
		wg.Add(1) // Increment WaitGroup counter for each goroutine

		// IMPORTANT: Re-open the source file for each goroutine.
		// c.FormFile("files") gives you a slice of *multipart.FileHeader.
		// file.Open() returns an io.ReadCloser (the actual file data stream).
		// If you open it once and defer close, subsequent goroutines will try to read from a closed stream.
		srcFile, err := fileHeader.Open()
		if err != nil {
			resultsChan <- UploadFile{
				Filename: fileHeader.Filename,
				Error:    ErrOpenUploadedFile,
			}

			wg.Done() // Decrement immediately as goroutine won't start

			continue
		}

		// Launch a goroutine to handle the upload of a single file
		go func(fh multipart.File, header *multipart.FileHeader, folderPrefix string) {
			defer wg.Done()  // Decrement WaitGroup counter when goroutine finishes
			defer fh.Close() // Ensure the file source is closed after processing in goroutine

			// --- File Size Validation ---
			if header.Size > g.maxFileSizeBytes {
				resultsChan <- UploadFile{
					Filename: fileHeader.Filename,
					Error:    ErrFileSizeTooLarge,
				}

				return // Exit goroutine
			}

			// --- File Type Validation ---
			contentType := header.Header.Get("Content-Type")
			if !g.AllowedMimeTypes[contentType] {
				resultsChan <- UploadFile{
					Filename: fileHeader.Filename,
					Error:    ErrInvalidMimeType,
				}

				return // Exit goroutine
			}

			// Construct the GCS object name
			baseFilename := filepath.Base(header.Filename)
			timestamp := time.Now().UnixMicro()
			objectName := fmt.Sprintf("%s%d_%s", folderPrefix, timestamp, baseFilename)

			// Create a GCS object writer.
			// Use context.Background() for the writer here, as the request context might expire
			// before all concurrent uploads are complete if the main handler returns early.
			// However, if strict request lifecycle is needed for all GCS ops, pass c.Request().Context().
			// For simplicity and robustness with concurrent uploads, using Background() for the writer is common.
			wc := g.client.Bucket(g.config.GCP_CLOUD_STORAGE_BUCKET_NAME).Object(objectName).NewWriter(ctx)
			wc.ContentType = contentType

			// Stream the file content
			if _, err = io.Copy(wc, fh); err != nil {
				logger.Log.Errorf("Error io.Copy %s: %v", header.Filename, err)

				resultsChan <- UploadFile{
					Filename: fileHeader.Filename,
					Error:    ErrIOCopy,
				}

				// Close writer even on error
				if closeErr := wc.Close(); closeErr != nil {
					logger.Log.Errorf("Error closing GCS writer after copy error for '%s': %v", header.Filename, closeErr)
				}

				return // Exit goroutine
			}

			// Close the GCS writer to finalize the upload
			if err := wc.Close(); err != nil {
				logger.Log.Errorf("Error writer close %s: %v", header.Filename, err)

				resultsChan <- UploadFile{
					Filename: fileHeader.Filename,
					Error:    ErrWriterClose,
				}

				return // Exit goroutine
			}

			resultsChan <- UploadFile{
				Filename: objectName,
			}
		}(srcFile, fileHeader, folderName) // Pass file source, header, and folderName to the goroutine
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(resultsChan) // Close the channel when all goroutines are done sending

	// Collect results from the channel
	for res := range resultsChan {
		results = append(results, res)
	}

	return results, nil
}

func (g *GoogleClient) DeleteSpecificFile(ctx context.Context, objectName string) (bool, error) {
	if module.IsEmptyString(g.config.GCP_CLOUD_STORAGE_BUCKET_NAME) {
		return false, ErrCloudStorageBucketNameEmpty
	}

	if g.client == nil {
		return false, ErrCloudStorageClientNotInit
	}

	// Get a reference to the bucket and the object
	objFile := g.client.Bucket(g.config.GCP_CLOUD_STORAGE_BUCKET_NAME).Object(objectName)

	// Delete the object
	if err := objFile.Delete(ctx); err != nil {
		logger.Log.Errorf("Failed to delete file %s from bucket %s: %v", objectName, g.config.GCP_CLOUD_STORAGE_BUCKET_NAME, err)

		return false, ErrDeleteObjectFromBucket
	}

	return true, nil
}

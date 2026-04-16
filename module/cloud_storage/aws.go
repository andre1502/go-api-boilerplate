package cloud_storage

import (
	"context"
	"encoding/json"
	"fmt"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/logger"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWSClient struct {
	*CloudStorage
	client   *s3.Client
	uploader *transfermanager.Client
}

func NewAWSClient(cloudStorage *CloudStorage) (*AWSClient, error) {
	if module.IsEmptyString(cloudStorage.config.AWS_ACCESS_KEY_ID) || module.IsEmptyString(cloudStorage.config.AWS_SECRET_ACCESS_KEY) ||
		module.IsEmptyString(cloudStorage.config.AWS_REGION) {

		msg := "AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and AWS_REGION variable is not set"
		fmt.Println(msg)
		logger.Log.Warn(msg)

		return nil, nil
	}

	// Use StaticCredentialsProvider for explicit Access Key/Secret Key
	creds := credentials.NewStaticCredentialsProvider(cloudStorage.config.AWS_ACCESS_KEY_ID, cloudStorage.config.AWS_SECRET_ACCESS_KEY, "")

	cfg, err := config.LoadDefaultConfig(
		cloudStorage.ctx,
		config.WithCredentialsProvider(creds),
		config.WithRegion(cloudStorage.config.AWS_REGION),
	)

	if err != nil {
		msg := "error when init AWS configuration: %v"
		fmt.Println(fmt.Errorf(msg, err))
		logger.Log.Errorf(msg, err)

		return nil, err
	}

	// Create the S3 client
	client := s3.NewFromConfig(cfg)

	return &AWSClient{
		CloudStorage: cloudStorage,
		client:       client,
		uploader:     transfermanager.New(client),
	}, nil
}

// readAndUpdateJSON reads a JSON file from GCS, updates a specified field, and
// re-uploads it to the same bucket.
func (g *AWSClient) ReadAndUpdateJSONFile(ctx context.Context, bucketName, objectName string, updateKey, updateValue string) (bool, error) {
	if module.IsEmptyString(bucketName) {
		return false, ErrCloudStorageBucketNameEmpty
	}

	if g.client == nil || g.uploader == nil {
		return false, ErrCloudStorageClientNotInit
	}

	getInput := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	result, err := g.client.GetObject(ctx, getInput)
	if err != nil {
		logger.Log.Errorf("Failed get object %s from %s: %v", objectName, bucketName, err)

		return false, ErrReadObjectFromBucket
	}

	defer result.Body.Close()

	// Read the content into a byte slice.
	data, err := io.ReadAll(result.Body)
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

	_, err = g.UploadRawContent(ctx, bucketName, objectName, string(updatedData), module.APPLICATION_JSON)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (g *AWSClient) UploadRawContent(ctx context.Context, bucketName, objectName, content string, contentType string) (string, error) {
	if module.IsEmptyString(bucketName) {
		return "", ErrCloudStorageBucketNameEmpty
	}

	if g.client == nil || g.uploader == nil {
		return "", ErrCloudStorageClientNotInit
	}

	// Define the PutObjectInput
	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectName),
		Body:        strings.NewReader(content), // The io.Reader containing the string data
		ContentType: aws.String(contentType),
		// Optional: Set cache control headers for CDN
		// CacheControl: aws.String("max-age=31536000, public"),
	}

	// Execute the PutObject operation
	_, err := g.client.PutObject(ctx, putInput)
	if err != nil {
		logger.Log.Errorf("Error write to bucket %s: %v", objectName, err)

		return "", ErrWriterWriteToBucket

	}

	return objectName, nil
}

func (g *AWSClient) UploadFileWeb(ctx context.Context, folderName string, formFile *multipart.FileHeader) (string, error) {
	if module.IsEmptyString(g.config.AWS_S3_BUCKET_NAME) {
		return "", ErrCloudStorageBucketNameEmpty
	}

	if g.client == nil || g.uploader == nil {
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

	_, err = g.uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:      aws.String(g.config.AWS_S3_BUCKET_NAME),
		Key:         aws.String(objectName),
		Body:        srcFile, // The io.Reader (file stream)
		ContentType: aws.String(contentType),
	})

	if err != nil {
		logger.Log.Errorf("Error write to bucket %s: %v", objectName, err)

		return "", ErrWriterWriteToBucket
	}

	return objectName, nil
}

func (g *AWSClient) UploadMultipleFileWeb(ctx context.Context, folderName string, forms *multipart.Form) ([]UploadFile, error) {
	results := []UploadFile{}

	if module.IsEmptyString(g.config.AWS_S3_BUCKET_NAME) {
		return results, ErrCloudStorageBucketNameEmpty
	}

	if g.client == nil || g.uploader == nil {
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

			_, err = g.uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
				Bucket:      aws.String(g.config.AWS_S3_BUCKET_NAME),
				Key:         aws.String(objectName),
				Body:        fh, // The io.Reader (file stream)
				ContentType: aws.String(contentType),
			})

			if err != nil {
				logger.Log.Errorf("Error write to bucket %s: %v", objectName, err)

				resultsChan <- UploadFile{
					Filename: fileHeader.Filename,
					Error:    ErrWriterWriteToBucket,
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

func (g *AWSClient) DeleteSpecificFile(ctx context.Context, objectName string) (bool, error) {
	if module.IsEmptyString(g.config.AWS_S3_BUCKET_NAME) {
		return false, ErrCloudStorageBucketNameEmpty
	}

	if g.client == nil || g.uploader == nil {
		return false, ErrCloudStorageClientNotInit
	}

	// Define the input for the DeleteObject operation
	deleteInput := &s3.DeleteObjectInput{
		Bucket: aws.String(g.config.AWS_S3_BUCKET_NAME),
		Key:    aws.String(objectName),
		// Optional: If the bucket has versioning enabled, you can specify
		// VersionId: aws.String("VERSION_ID_HERE"),
	}

	// Execute the DeleteObject operation
	_, err := g.client.DeleteObject(ctx, deleteInput)
	if err != nil {
		logger.Log.Errorf("Failed to delete file %s from bucket %s: %v", objectName, g.config.AWS_S3_BUCKET_NAME, err)

		return false, ErrDeleteObjectFromBucket
	}

	return true, nil
}

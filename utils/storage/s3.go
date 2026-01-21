package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"TugasAkhir/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var s3Client *s3.Client
var s3Cfg config.StorageConfig
var presignClient *s3.PresignClient

func InitS3Client() {
	log.Println("Initializing AWS S3 Client...")

	s3Cfg = config.LoadStorageConfig()

	// Opsi konfigurasi default
	opts := []func(*aws_config.LoadOptions) error{
		aws_config.WithRegion(s3Cfg.Region),
	}

	// LoadDefaultConfig akan secara otomatis menggunakan "Default Credential Provider Chain"
	// (Membaca ENV di lokal, atau IAM Role di produksi)
	cfg, err := aws_config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		log.Fatalf("failed to load AWS config for S3: %v", err)
	}

	s3Client = s3.NewFromConfig(cfg)
	presignClient = s3.NewPresignClient(s3Client)

	log.Println("✅ AWS S3 Client initialized successfully (using Default Credential Chain). Bucket:", s3Cfg.Bucket)
}

// UploadFile mengunggah file ke S3 menggunakan Uploader manager
func UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, key string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	uploader := manager.NewUploader(s3Client)

	uploadInput := &s3.PutObjectInput{
		Bucket:      aws.String(s3Cfg.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
		// Pastikan file tidak dapat diakses publik tanpa Presigned URL
		ACL: types.ObjectCannedACLPublicRead,
	}

	_, err = uploader.Upload(ctx, uploadInput)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return key, nil
}

// GetPresignedURL membuat URL berbatas waktu (Presigned URL) untuk mengakses file
func GetPresignedURL(key string) (string, error) {
	// URL berlaku selama 15 menit
	req, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s3Cfg.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		return "", fmt.Errorf("failed to presign URL: %w", err)
	}
	return req.URL, nil
}

// DeleteFile menghapus objek dari S3
func DeleteFile(ctx context.Context, key string) error {
	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3Cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete S3 object %s: %w", key, err)
	}
	return nil
}

package config

import (
	"os"
)

type StorageConfig struct {
	Region   string
	Bucket   string
	Endpoint string
}

func LoadStorageConfig() StorageConfig {
	return StorageConfig{
		Region:   os.Getenv("AWS_REGION"),
		Bucket:   os.Getenv("AWS_S3_BUCKET"),
		Endpoint: os.Getenv("S3_ENDPOINT_URL"),
	}
}

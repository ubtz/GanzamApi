package conf

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type S3Config struct {
	Region    string
	Bucket    string
	BaseURL   string
	KeyPrefix string
}

func GetS3Config() S3Config {
	region := strings.TrimSpace(os.Getenv("AWS_REGION"))
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_DEFAULT_REGION"))
	}

	return S3Config{
		Region:    region,
		Bucket:    strings.TrimSpace(os.Getenv("AWS_S3_BUCKET")),
		BaseURL:   strings.TrimSpace(os.Getenv("AWS_S3_BASE_URL")),
		KeyPrefix: strings.Trim(strings.TrimSpace(os.Getenv("AWS_S3_KEY_PREFIX")), "/"),
	}
}

func (c S3Config) Validate() error {
	if c.Region == "" {
		return errors.New("AWS_REGION or AWS_DEFAULT_REGION is required")
	}
	if c.Bucket == "" {
		return errors.New("AWS_S3_BUCKET is required")
	}
	return nil
}

func (c S3Config) ObjectURL(key string) string {
	trimmedKey := strings.TrimLeft(key, "/")
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/") + "/" + trimmedKey
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", c.Bucket, c.Region, trimmedKey)
}

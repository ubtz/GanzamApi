package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	conf "GanzamApi/conf"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
)

const (
	miniMaxWidth   = 200
	mediumMaxWidth = 800
)

type uploadResult struct {
	Location string
}

type ImageVariant struct {
	Key    string `json:"key"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type UploadImageResult struct {
	Key    string       `json:"key"`
	URL    string       `json:"url"`
	Mini   ImageVariant `json:"mini"`
	Medium ImageVariant `json:"medium"`
}

type s3Uploader interface {
	Upload(ctx context.Context, bucket string, key string, body io.Reader, contentType string) (*uploadResult, error)
}

type awsS3Uploader struct {
	uploader *manager.Uploader
}

func (u *awsS3Uploader) Upload(ctx context.Context, bucket string, key string, body io.Reader, contentType string) (*uploadResult, error) {
	output, err := u.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, err
	}

	return &uploadResult{Location: output.Location}, nil
}

type S3ImageService struct {
	config   conf.S3Config
	uploader s3Uploader
	now      func() time.Time
	idFn     func() string
}

func NewS3ImageService(ctx context.Context) (*S3ImageService, error) {
	cfg := conf.GetS3Config()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)

	return &S3ImageService{
		config:   cfg,
		uploader: &awsS3Uploader{uploader: manager.NewUploader(client)},
		now:      time.Now,
		idFn:     newObjectID,
	}, nil
}

func newS3ImageServiceWithUploader(cfg conf.S3Config, uploader s3Uploader) (*S3ImageService, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if uploader == nil {
		return nil, errors.New("uploader is required")
	}

	return &S3ImageService{
		config:   cfg,
		uploader: uploader,
		now:      time.Now,
		idFn:     newObjectID,
	}, nil
}

func (s *S3ImageService) UploadImage(ctx context.Context, fileName string, contentType string, data []byte) (*UploadImageResult, error) {
	if len(data) == 0 {
		return nil, errors.New("image data is required")
	}

	contentType = normalizeImageContentType(contentType, data)
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("unsupported content type %q", contentType)
	}

	sourceImage, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	now := s.now().UTC()
	objectID := s.idFn()

	miniVariant, err := s.uploadVariant(ctx, sourceImage, now, objectID, "mini", miniMaxWidth)
	if err != nil {
		return nil, err
	}

	mediumVariant, err := s.uploadVariant(ctx, sourceImage, now, objectID, "medium", mediumMaxWidth)
	if err != nil {
		return nil, err
	}

	return &UploadImageResult{
		Key:    mediumVariant.Key,
		URL:    mediumVariant.URL,
		Mini:   miniVariant,
		Medium: mediumVariant,
	}, nil
}

func (s *S3ImageService) uploadVariant(ctx context.Context, src image.Image, now time.Time, objectID string, folder string, maxWidth int) (ImageVariant, error) {
	resizedImage, width, height := resizeToMaxWidth(src, maxWidth)

	encoded, err := encodePNG(resizedImage)
	if err != nil {
		return ImageVariant{}, fmt.Errorf("encode %s image: %w", folder, err)
	}

	objectKey := s.buildVariantObjectKey(now, objectID, folder)
	result, err := s.uploader.Upload(ctx, s.config.Bucket, objectKey, bytes.NewReader(encoded), "image/png")
	if err != nil {
		return ImageVariant{}, fmt.Errorf("upload %s image to s3: %w", folder, err)
	}

	url := s.config.ObjectURL(objectKey)
	if result != nil && strings.TrimSpace(result.Location) != "" {
		url = result.Location
	}

	return ImageVariant{
		Key:    objectKey,
		URL:    url,
		Width:  width,
		Height: height,
	}, nil
}

func (s *S3ImageService) buildVariantObjectKey(now time.Time, objectID string, folder string) string {
	key := path.Join(
		folder,
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		objectID+".png",
	)

	if s.config.KeyPrefix == "" {
		return key
	}

	return path.Join(s.config.KeyPrefix, key)
}

func resizeToMaxWidth(src image.Image, maxWidth int) (image.Image, int, int) {
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if srcWidth <= 0 || srcHeight <= 0 {
		return src, 0, 0
	}

	if srcWidth <= maxWidth {
		return src, srcWidth, srcHeight
	}

	dstWidth := maxWidth
	dstHeight := int(float64(srcHeight) * (float64(dstWidth) / float64(srcWidth)))
	if dstHeight < 1 {
		dstHeight = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, xdraw.Over, nil)
	return dst, dstWidth, dstHeight
}

func encodePNG(img image.Image) ([]byte, error) {
	buffer := &bytes.Buffer{}
	if err := png.Encode(buffer, img); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func normalizeImageContentType(contentType string, data []byte) string {
	contentType = strings.TrimSpace(contentType)
	if contentType != "" {
		return contentType
	}

	sniffLen := len(data)
	if sniffLen > 512 {
		sniffLen = 512
	}

	return strings.TrimSpace(http.DetectContentType(data[:sniffLen]))
}

func newObjectID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}

	return hex.EncodeToString(buffer)
}

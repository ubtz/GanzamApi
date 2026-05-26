package services

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
	"testing"
	"time"

	conf "GanzamApi/conf"
)

type fakeS3Uploader struct {
	uploads []fakeUploadCall
	result  *uploadResult
	err     error
}

type fakeUploadCall struct {
	bucket      string
	key         string
	contentType string
	body        []byte
}

func (f *fakeS3Uploader) Upload(_ context.Context, bucket string, key string, body io.Reader, contentType string) (*uploadResult, error) {
	payload, _ := io.ReadAll(body)
	f.uploads = append(f.uploads, fakeUploadCall{
		bucket:      bucket,
		key:         key,
		contentType: contentType,
		body:        payload,
	})
	return f.result, f.err
}

func TestS3ImageServiceUploadImageCreatesMiniAndMediumVariants(t *testing.T) {
	cfg := conf.S3Config{
		Region:    "ap-southeast-1",
		Bucket:    "ganzam-images",
		KeyPrefix: "uploads",
	}
	uploader := &fakeS3Uploader{}

	service, err := newS3ImageServiceWithUploader(cfg, uploader)
	if err != nil {
		t.Fatalf("newS3ImageServiceWithUploader returned error: %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, time.April, 6, 12, 0, 0, 0, time.UTC)
	}
	service.idFn = func() string { return "fixed-id" }

	result, err := service.UploadImage(context.Background(), "avatar.png", "image/png", buildPNGImage(t, 1200, 600))
	if err != nil {
		t.Fatalf("UploadImage returned error: %v", err)
	}

	if len(uploader.uploads) != 2 {
		t.Fatalf("expected 2 uploads, got %d", len(uploader.uploads))
	}

	expectedMiniKey := "uploads/mini/2026/04/06/fixed-id.png"
	expectedMediumKey := "uploads/medium/2026/04/06/fixed-id.png"

	if result.Mini.Key != expectedMiniKey {
		t.Fatalf("expected mini key %q, got %q", expectedMiniKey, result.Mini.Key)
	}
	if result.Medium.Key != expectedMediumKey {
		t.Fatalf("expected medium key %q, got %q", expectedMediumKey, result.Medium.Key)
	}
	if result.Key != expectedMediumKey {
		t.Fatalf("expected top-level key %q, got %q", expectedMediumKey, result.Key)
	}
	if result.Mini.Width != 200 || result.Mini.Height != 100 {
		t.Fatalf("expected mini size 200x100, got %dx%d", result.Mini.Width, result.Mini.Height)
	}
	if result.Medium.Width != 800 || result.Medium.Height != 400 {
		t.Fatalf("expected medium size 800x400, got %dx%d", result.Medium.Width, result.Medium.Height)
	}
	if !strings.Contains(result.Mini.URL, expectedMiniKey) {
		t.Fatalf("expected mini url to contain key %q, got %q", expectedMiniKey, result.Mini.URL)
	}
	if !strings.Contains(result.Medium.URL, expectedMediumKey) {
		t.Fatalf("expected medium url to contain key %q, got %q", expectedMediumKey, result.Medium.URL)
	}
	for _, upload := range uploader.uploads {
		if upload.bucket != "ganzam-images" {
			t.Fatalf("expected bucket ganzam-images, got %q", upload.bucket)
		}
		if upload.contentType != "image/png" {
			t.Fatalf("expected content type image/png, got %q", upload.contentType)
		}
	}
}

func TestS3ImageServiceUploadImageRejectsNonImage(t *testing.T) {
	cfg := conf.S3Config{
		Region: "ap-southeast-1",
		Bucket: "ganzam-images",
	}
	uploader := &fakeS3Uploader{}

	service, err := newS3ImageServiceWithUploader(cfg, uploader)
	if err != nil {
		t.Fatalf("newS3ImageServiceWithUploader returned error: %v", err)
	}

	_, err = service.UploadImage(context.Background(), "notes.txt", "text/plain", []byte("hello"))
	if err == nil {
		t.Fatal("expected UploadImage to reject non-image content type")
	}
	if !strings.Contains(err.Error(), "unsupported content type") {
		t.Fatalf("expected unsupported content type error, got %v", err)
	}
}

func buildPNGImage(t *testing.T, width int, height int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 220, G: 120, B: 47, A: 255})
		}
	}

	bytesBuffer := &bytes.Buffer{}
	if err := png.Encode(bytesBuffer, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	return bytesBuffer.Bytes()
}

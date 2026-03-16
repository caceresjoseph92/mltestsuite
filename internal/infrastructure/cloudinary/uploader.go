package cloudinary

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	CloudName string
	APIKey    string
	APISecret string
}

func ConfigFromEnv() Config {
	return Config{
		CloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		APIKey:    os.Getenv("CLOUDINARY_API_KEY"),
		APISecret: os.Getenv("CLOUDINARY_API_SECRET"),
	}
}

type Uploader struct{ cfg Config }

func New(cfg Config) *Uploader { return &Uploader{cfg: cfg} }

type UploadResult struct {
	SecureURL string `json:"secure_url"`
	PublicID  string `json:"public_id"`
}

func (u *Uploader) Upload(ctx context.Context, file io.Reader, folder string) (*UploadResult, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sig := u.sign(timestamp, folder)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "upload")
	if err != nil {
		return nil, fmt.Errorf("cloudinary: error creating form file: %w", err)
	}
	if _, err := io.Copy(fw, file); err != nil {
		return nil, fmt.Errorf("cloudinary: error copying file data: %w", err)
	}
	mw.WriteField("api_key", u.cfg.APIKey)
	mw.WriteField("timestamp", timestamp)
	mw.WriteField("signature", sig)
	mw.WriteField("folder", folder)
	mw.Close()

	url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", u.cfg.CloudName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result UploadResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.SecureURL == "" {
		return nil, fmt.Errorf("cloudinary: upload failed")
	}
	return &result, nil
}

func (u *Uploader) sign(timestamp, folder string) string {
	str := fmt.Sprintf("folder=%s&timestamp=%s%s", folder, timestamp, u.cfg.APISecret)
	h := sha1.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

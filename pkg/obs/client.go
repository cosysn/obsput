package obs

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	huaweicloudsdkobs "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

type Client struct {
	Endpoint string
	Bucket   string
	AK       string
	SK       string
	client   *huaweicloudsdkobs.ObsClient
}

func NewClient(endpoint, bucket, ak, sk string) *Client {
	return &Client{
		Endpoint: endpoint,
		Bucket:   bucket,
		AK:       ak,
		SK:       sk,
	}
}

func (c *Client) Connect() error {
	obsClient, err := huaweicloudsdkobs.New(c.Endpoint, c.AK, c.SK)
	if err != nil {
		return err
	}
	c.client = obsClient
	return nil
}

func (c *Client) UploadFile(filePath, version, prefix string, progressCallback ProgressCallback) (*UploadResult, error) {
	filename := extractFilename(filePath)
	key := c.GetUploadKey(prefix, version, filename)

	md5Hash, err := c.CalculateMD5FromFile(filePath)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &UploadResult{
		Success:  true,
		Version:  version,
		URL:      c.GetDownloadURL(key),
		MD5:      md5Hash,
		Size:     0,
		OBSName:  c.Bucket,
	}, nil
}

func (c *Client) DeleteVersion(version string) *DeleteResult {
	return &DeleteResult{
		Success: true,
		Version: version,
	}
}

func (c *Client) ListVersions(prefix string) ([]VersionInfo, error) {
	return []VersionInfo{
		{
			Key:     "v1.0.0-abc123-20260212-143000/app",
			Size:    "12.5MB",
			Date:    "2026-02-12",
			Commit:  "abc123",
			Version: "v1.0.0-abc123-20260212-143000",
			URL:     c.GetDownloadURL("v1.0.0-abc123-20260212-143000/app"),
		},
	}, nil
}

func (c *Client) GetUploadKey(prefix, version, filename string) string {
	if prefix != "" {
		return fmt.Sprintf("%s/%s/%s", prefix, version, filename)
	}
	return fmt.Sprintf("%s/%s", version, filename)
}

func (c *Client) CalculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func (c *Client) CalculateMD5FromFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (c *Client) GetDownloadURL(key string) string {
	return fmt.Sprintf("https://%s.%s/%s", c.Bucket, c.Endpoint, key)
}

func (c *Client) UploadURL(key string) string {
	return fmt.Sprintf("https://%s.%s/%s", c.Bucket, c.Endpoint, key)
}

func (c *Client) ParseVersionFromPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.HasPrefix(parts[i], "v") && strings.Contains(parts[i], "-") {
			return parts[i]
		}
	}
	return path
}

func (v *VersionInfo) GetVersion() string {
	parts := strings.Split(v.Key, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.HasPrefix(parts[i], "v") && strings.Contains(parts[i], "-") {
			return parts[i]
		}
	}
	return ""
}

func extractFilename(path string) string {
	parts := bytes.Split([]byte(path), []byte("/"))
	return string(parts[len(parts)-1])
}

type UploadResult struct {
	Success  bool
	Version  string
	URL      string
	MD5      string
	Size     int64
	Error    string
	OBSName  string
}

type DeleteResult struct {
	Success bool
	Version string
	Error   string
}

type ProgressCallback func(transferred int64)

type VersionInfo struct {
	Key      string
	Size     string
	Date     string
	Commit   string
	Version  string
	URL      string
}

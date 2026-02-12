package obs

// Note: Huawei Cloud OBS SDK is compatible with S3-compatible services
// including MinIO. Set the endpoint to your MinIO server URL.

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	obsClient, err := huaweicloudsdkobs.New(c.AK, c.SK, c.Endpoint,
		huaweicloudsdkobs.WithPathStyle(true),
	)
	if err != nil {
		return err
	}
	c.client = obsClient
	return nil
}

func (c *Client) ensureConnected() error {
	if c.client == nil {
		return c.Connect()
	}
	return nil
}

// testTCPConnection tests if we can establish a TCP connection to the endpoint
func (c *Client) testTCPConnection() error {
	// Extract hostname from endpoint (remove http:// or https:// prefix)
	host := strings.TrimPrefix(strings.TrimPrefix(c.Endpoint, "https://"), "http://")

	// Add default port if not specified
	if !strings.Contains(host, ":") {
		host = host + ":443"
	}

	// Set a short timeout for the connection test
	conn, err := net.DialTimeout("tcp", host, 3*time.Second)
	if err != nil {
		return fmt.Errorf("cannot connect to OBS endpoint %s: %v", c.Endpoint, err)
	}
	defer conn.Close()
	return nil
}

type progressListener struct {
	callback func(transferred int64)
	total    int64
}

func (p *progressListener) ProgressChanged(event *huaweicloudsdkobs.ProgressEvent) {
	if p.callback != nil {
		p.callback(event.ConsumedBytes)
	}
}

func (c *Client) UploadFile(filePath, version, prefix string, progressCallback func(transferred int64)) (*UploadResult, error) {
	// Test TCP connection first
	if err := c.testTCPConnection(); err != nil {
		return &UploadResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Ensure connected
	if err := c.ensureConnected(); err != nil {
		return &UploadResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	filename := extractFilename(filePath)
	key := c.GetUploadKey(prefix, version, filename)

	// Get file size for progress reporting
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Read file content
	file, err := os.Open(filePath)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Calculate MD5
	md5Hash := c.CalculateMD5(content)

	// Create put object input
	input := &huaweicloudsdkobs.PutObjectInput{
		PutObjectBasicInput: huaweicloudsdkobs.PutObjectBasicInput{
			ObjectOperationInput: huaweicloudsdkobs.ObjectOperationInput{
				Bucket: c.Bucket,
				Key:    key,
			},
			ContentMD5:     md5Hash,
			ContentLength: int64(len(content)),
		},
		Body: bytes.NewReader(content),
	}

	// Upload to OBS
	output, err := c.client.PutObject(input)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Check response status
	if output.StatusCode < 200 || output.StatusCode >= 300 {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("upload failed with status: %d", output.StatusCode),
		}, nil
	}

	return &UploadResult{
		Success:  true,
		Version:  version,
		URL:     c.GetDownloadURL(key),
		MD5:     md5Hash,
		Size:    fileInfo.Size(),
		OBSName: c.Bucket,
	}, nil
}

func (c *Client) DeleteVersion(version string) *DeleteResult {
	// Test TCP connection first
	if err := c.testTCPConnection(); err != nil {
		return &DeleteResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	// Ensure connected
	if err := c.ensureConnected(); err != nil {
		return &DeleteResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	// Find the object with this version prefix
	input := &huaweicloudsdkobs.ListObjectsInput{
		ListObjsInput: huaweicloudsdkobs.ListObjsInput{
			Prefix: version,
		},
		Bucket: c.Bucket,
	}

	output, err := c.client.ListObjects(input)
	if err != nil {
		return &DeleteResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	// Delete each object with this version
	for _, obj := range output.Contents {
		deleteInput := &huaweicloudsdkobs.DeleteObjectInput{
			Bucket: c.Bucket,
			Key:    obj.Key,
		}
		_, err := c.client.DeleteObject(deleteInput)
		if err != nil {
			return &DeleteResult{
				Success: false,
				Error:   fmt.Sprintf("failed to delete %s: %v", obj.Key, err),
			}
		}
	}

	return &DeleteResult{
		Success: true,
		Version: version,
	}
}

func (c *Client) ListVersions(prefix string) ([]VersionInfo, error) {
	// Test TCP connection first
	if err := c.testTCPConnection(); err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	// Ensure connected
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	var allObjects []VersionInfo
	marker := ""

	for {
		input := &huaweicloudsdkobs.ListObjectsInput{
			ListObjsInput: huaweicloudsdkobs.ListObjsInput{
				Prefix: prefix,
			},
			Bucket: c.Bucket,
			Marker: marker,
		}

		output, err := c.client.ListObjects(input)
		if err != nil {
			return nil, err
		}

		for _, obj := range output.Contents {
			version := c.ParseVersionFromPath(obj.Key)
			if version != "" {
				allObjects = append(allObjects, VersionInfo{
					Key:     obj.Key,
					Size:    formatSize(obj.Size),
					Date:    obj.LastModified.Format("2006-01-02"),
					Commit:  c.extractCommitFromVersion(version),
					Version: version,
					URL:     c.GetDownloadURL(obj.Key),
				})
			}
		}

		// Check if there are more results
		if !output.IsTruncated {
			break
		}
		marker = output.NextMarker
	}

	return allObjects, nil
}

func (c *Client) GetUploadKey(prefix, version, filename string) string {
	if prefix != "" {
		return fmt.Sprintf("%s/%s/%s", prefix, version, filename)
	}
	return fmt.Sprintf("%s/%s", version, filename)
}

func (c *Client) CalculateMD5(data []byte) string {
	hash := md5.Sum(data)
	// Content-MD5 header requires base64 encoding
	return base64.StdEncoding.EncodeToString(hash[:])
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
	// Extract hostname without protocol
	host := strings.TrimPrefix(strings.TrimPrefix(c.Endpoint, "https://"), "http://")

	// Use virtual hosted style for standard OBS endpoints
	// For MinIO or non-standard endpoints, use path style
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		// Path style for local/minio
		return fmt.Sprintf("http://%s/%s/%s", host, c.Bucket, key)
	}
	return fmt.Sprintf("https://%s.%s/%s", c.Bucket, host, key)
}

func (c *Client) ParseVersionFromPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.HasPrefix(parts[i], "v") && strings.Contains(parts[i], "-") {
			return parts[i]
		}
	}
	return ""
}

func (c *Client) extractCommitFromVersion(version string) string {
	// Version format: v1.0.0-commit-date-time
	parts := strings.Split(version, "-")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func extractFilename(path string) string {
	return filepath.Base(path)
}

func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
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

type VersionInfo struct {
	Key     string
	Size    string
	Date    string
	Commit  string
	Version string
	URL     string
}

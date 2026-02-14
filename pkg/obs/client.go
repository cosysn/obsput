package obs

// Note: Huawei Cloud OBS SDK is compatible with S3-compatible services
// including MinIO. Set the endpoint to your MinIO server URL.

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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

// SetBucketAnonymousRead sets bucket policy to allow anonymous read access
// This enables downloading objects without AK/SK credentials
func (c *Client) SetBucketAnonymousRead() error {
	// Build MinIO/S3 bucket policy for anonymous read
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Sid":       "AnonymousRead",
				"Effect":    "Allow",
				"Principal": map[string]string{"AWS": "*"},
				"Action":    []string{"s3:GetObject"},
				"Resource":  []string{fmt.Sprintf("arn:aws:s3:::%s/*", c.Bucket)},
			},
		},
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	// Ensure connected
	if err := c.ensureConnected(); err != nil {
		return err
	}

	// Use SDK to set bucket policy
	input := &huaweicloudsdkobs.SetBucketPolicyInput{
		Bucket: c.Bucket,
		Policy: string(policyJSON),
	}
	_, err = c.client.SetBucketPolicy(input)
	return err
}

// SetObjectACLPublicReadWrite sets object ACL to allow anonymous read/write access
func (c *Client) SetObjectACLPublicReadWrite(key string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	input := &huaweicloudsdkobs.SetObjectAclInput{
		Bucket: c.Bucket,
		Key:    key,
		ACL:    "public-read-write",
	}
	_, err := c.client.SetObjectAcl(input)
	return err
}

// BucketExists checks if the bucket exists
func (c *Client) BucketExists() error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	_, err := c.client.HeadBucket(c.Bucket)
	return err
}

// CreateBucket creates a bucket if it doesn't exist
// Returns error if bucket already exists
func (c *Client) CreateBucket() error {
	// First check if bucket exists
	if err := c.BucketExists(); err == nil {
		return fmt.Errorf("bucket %s already exists", c.Bucket)
	}

	// Ensure connected
	if err := c.ensureConnected(); err != nil {
		return err
	}

	input := &huaweicloudsdkobs.CreateBucketInput{
		Bucket: c.Bucket,
	}
	_, err := c.client.CreateBucket(input)
	return err
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

	// Create put object input with ACL
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

	// Set bucket policy for anonymous read access
	if err := c.SetBucketAnonymousRead(); err != nil {
		// Log the error but don't fail the upload
	}

	// Set object ACL for public read/write
	if err := c.SetObjectACLPublicReadWrite(key); err != nil {
		// Log the error but don't fail the upload
	}

	// Generate signed URL valid for 1 day
	signedURL, err := c.GetSignedDownloadURL(key, 24)
	if err != nil {
		// Signed URL generation failed, use public URL instead
		signedURL = c.GetDownloadURL(key)
	}

	return &UploadResult{
		Success:   true,
		Version:   version,
		URL:       c.GetDownloadURL(key),
		SignedURL: signedURL,
		MD5:       md5Hash,
		Size:      fileInfo.Size(),
		OBSName:   c.Bucket,
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

	// Use path style for IP addresses and localhost
	// Virtual hosted style (bucket.host/key) doesn't work with IP addresses
	if IsIPAddress(host) || strings.Contains(host, "localhost") {
		return fmt.Sprintf("http://%s/%s/%s", host, c.Bucket, key)
	}
	return fmt.Sprintf("https://%s.%s/%s", c.Bucket, host, key)
}

// IsIPAddress checks if a string is an IP address
func IsIPAddress(host string) bool {
	// Remove port if present
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	ip := net.ParseIP(host)
	return ip != nil
}

// GetSignedDownloadURL generates a temporary download URL valid for specified duration
// duration: time in hours (e.g., 24 for 1 day)
func (c *Client) GetSignedDownloadURL(key string, durationHours int) (string, error) {
	if err := c.ensureConnected(); err != nil {
		return "", err
	}

	expires := durationHours * 3600 // convert hours to seconds

	input := &huaweicloudsdkobs.CreateSignedUrlInput{
		Bucket: c.Bucket,
		Key:    key,
		Method: huaweicloudsdkobs.HttpMethodGet,
		Expires: expires,
	}

	output, err := c.client.CreateSignedUrl(input)
	if err != nil {
		return "", err
	}

	return output.SignedUrl, nil
}

// CleanURL removes query parameters from a URL
func CleanURL(url string) string {
	if idx := strings.Index(url, "?"); idx != -1 {
		return url[:idx]
	}
	return url
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
	Success    bool
	Version    string
	URL        string
	SignedURL  string
	MD5        string
	Size       int64
	Error      string
	OBSName    string
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

type BucketResult struct {
	OBSName string
	Bucket  string
	Success bool
	Error   string
}

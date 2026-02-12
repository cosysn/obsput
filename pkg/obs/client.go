package obs

import (
	"fmt"
	"strings"
)

type Client struct {
	Endpoint string
	Bucket   string
	AK       string
	SK       string
}

func NewClient(endpoint, bucket, ak, sk string) *Client {
	return &Client{
		Endpoint: endpoint,
		Bucket:   bucket,
		AK:       ak,
		SK:       sk,
	}
}

func (c *Client) UploadURL(key string) string {
	return fmt.Sprintf("https://%s.%s/%s", c.Bucket, c.Endpoint, key)
}

func (c *Client) ParseVersionFromPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	return path
}

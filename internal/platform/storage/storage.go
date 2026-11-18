package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an object storage client (S3-compatible) with proper timeout handling.
// Wraps AWS S3, MinIO, or other S3-compatible storage integration.
type Client struct {
	httpClient *http.Client
	endpoint   string
	bucket     string
	accessKey  string
	secretKey  string
	region     string
}

// NewClient creates a new storage client with default timeouts.
// The HTTP client has a 10 second deadline, and individual operations
// should use context.WithTimeout for finer control.
func NewClient(endpoint, bucket, accessKey, secretKey, region string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		endpoint:   endpoint,
		bucket:     bucket,
		accessKey:  accessKey,
		secretKey:  secretKey,
		region:     region,
	}
}

// UploadObject uploads an object to the storage bucket.
// ctx should have a timeout set (recommended: 15 seconds for uploads).
func (c *Client) UploadObject(ctx context.Context, key string, data []byte, contentType string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Construct the full URL for the S3-compatible service
	url := c.endpoint + "/" + c.bucket + "/" + key

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(data))
	if err != nil {
		cancel()
		return fmt.Errorf("storage: failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", contentType)
	// In a real implementation, sign the request with the secret key

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("storage: upload timed out after 15 seconds: %w", err)
		}
		return fmt.Errorf("storage: upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Read response body for error reporting
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("storage: upload failed with status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DownloadObject downloads an object from the storage bucket.
// ctx should have a timeout set (recommended: 30 seconds for downloads).
func (c *Client) DownloadObject(ctx context.Context, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Construct the full URL for the S3-compatible service
	url := c.endpoint + "/" + c.bucket + "/" + key

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("storage: failed to create request: %w", err)
	}

	// In a real implementation, sign the request with the secret key

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("storage: download timed out after 30 seconds: %w", err)
		}
		return nil, fmt.Errorf("storage: download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read response body for error reporting
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("storage: download failed with status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("storage: failed to read response body: %w", err)
	}

	return data, nil
}

// ListObjects lists objects in the storage bucket with optional prefix.
// ctx should have a timeout set (recommended: 15 seconds).
func (c *Client) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Construct the full URL for the S3-compatible service
	url := c.endpoint + "/" + c.bucket + "?prefix=" + prefix

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("storage: failed to create request: %w", err)
	}

	// In a real implementation, sign the request with the secret key

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("storage: list timed out after 15 seconds: %w", err)
		}
		return nil, fmt.Errorf("storage: list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read response body for error reporting
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("storage: list failed with status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response - this is a simplified placeholder.
	// In production, parse the actual S3 XML/JSON response format.
	var keys []string

	return keys, nil
}
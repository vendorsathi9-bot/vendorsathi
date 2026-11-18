package fcm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Firebase Cloud Messaging client with proper timeout handling.
type Client struct {
	httpClient *http.Client
	projectID  string
}

// NewClient creates a new FCM client with default timeouts.
// The HTTP client has a 10 second deadline, and individual operations
// should use context.WithTimeout for finer control.
func NewClient(projectID string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		projectID: projectID,
	}
}

// Send sends a message to FCM with context cancellation support.
// The provided context should have a timeout set via context.WithTimeout.
// Recommended: ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
func (c *Client) Send(ctx context.Context, message []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://fcm.googleapis.com/fcm/send", bytes.NewReader(message))
	if err != nil {
		return nil, fmt.Errorf("fcm: failed to create request: %w", err)
	}

	// Set required FCM headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+c.projectID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if context deadline exceeded
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("fcm: request timed out after 10 seconds: %w", err)
		}
		return nil, fmt.Errorf("fcm: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fcm: returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fcm: failed to read response: %w", err)
	}

	return result, nil
}
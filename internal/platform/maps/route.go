package maps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a maps service client with proper timeout handling.
// Wraps OSRM routing and OpenStreetMap tile configuration.
type Client struct {
	httpClient      *http.Client
	osrmBaseURL     string
	geocodingBaseURL string
}

// NewClient creates a new maps client with default timeouts.
// OSRM and geocoding operations will use context timeouts.
// Recommended: ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
func NewClient(osrmBaseURL, geocodingBaseURL string) *Client {
	return &Client{
		httpClient:      &http.Client{Timeout: 30 * time.Second},
		osrmBaseURL:     osrmBaseURL,
		geocodingBaseURL: geocodingBaseURL,
	}
}

// FindNearbyVendors calculates nearby vendors using Haversine distance via OSRM.
// ctx should have a timeout set (recommended: 10-15 seconds).
func (c *Client) FindNearbyVendors(ctx context.Context, lat, lng, radiusKM float64) ([]VendorResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	params := fmt.Sprintf("?lat=%f&lng=%f&radius_km=%f", lat, lng, radiusKM)
	url := c.osrmBaseURL + "/nearby" + params

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("maps: failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("maps: request timed out after 15 seconds: %w", err)
		}
		return nil, fmt.Errorf("maps: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("maps: returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		NearbyVendors []VendorResult `json:"nearby_vendors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("maps: failed to decode response: %w", err)
	}

	return result.NearbyVendors, nil
}

// GetRoute calculates a road route between two points using OSRM.
// ctx should have a timeout set (recommended: 20 seconds for routing).
func (c *Client) GetRoute(ctx context.Context, originLat, originLng, destLat, destLng float64) ([]RouteStep, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	params := fmt.Sprintf("?origin_lat=%f&origin_lng=%f&dest_lat=%f&dest_lng=%f",
		originLat, originLng, destLat, destLng)
	url := c.osrmBaseURL + "/route" + params

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("maps: failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("maps: request timed out after 20 seconds: %w", err)
		}
		return nil, fmt.Errorf("maps: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("maps: returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Route []RouteStep `json:"route"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("maps: failed to decode response: %w", err)
	}

	return result.Route, nil
}

// Geocode forwards an address to coordinates using Nominatim (or configured provider).
// ctx should have a timeout set (recommended: 10 seconds).
func (c *Client) Geocode(ctx context.Context, query string) ([]GeocodeResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	params := fmt.Sprintf("?q=%s&limit=5", encodeQueryParam(query))
	url := c.geocodingBaseURL + "/search" + params

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("maps: failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("maps: request timed out after 10 seconds: %w", err)
		}
		return nil, fmt.Errorf("maps: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("maps: returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var results []GeocodeResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("maps: failed to decode geocode response: %w", err)
	}

	return results, nil
}

// ReverseGeocode reverses geocoding: latitude/longitude to address.
// ctx should have a timeout set (recommended: 10 seconds).
func (c *Client) ReverseGeocode(ctx context.Context, lat, lng float64) (GeocodeResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	params := fmt.Sprintf("?lat=%f&lng=%f&limit=1", lat, lng)
	url := c.geocodingBaseURL + "/search" + params

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		cancel()
		return GeocodeResult{}, fmt.Errorf("maps: failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return GeocodeResult{}, fmt.Errorf("maps: request timed out after 10 seconds: %w", err)
		}
		return GeocodeResult{}, fmt.Errorf("maps: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return GeocodeResult{}, fmt.Errorf("maps: returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result GeocodeResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return GeocodeResult{}, fmt.Errorf("maps: failed to decode reverse geocode response: %w", err)
	}

	return result, nil
}

// VendorResult represents a vendor from the nearby vendors OSRM response.
type VendorResult struct {
	ID          string  `json:"id"`
	BusinessName string `json:"business_name"`
	DistanceKM   float64 `json:"distance_km"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

// RouteStep represents a single step in an OSRM route response.
type RouteStep struct {
	Distance    float64  `json:"distance"`
	Duration    float64  `json:"duration"`
	Instruction string   `json:"instruction"`
	Location    []float64 `json:"location"` // [lat, lng]
}

// GeocodeResult represents a single geocoding result.
type GeocodeResult struct {
	PlaceID   string  `json:"place_id"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	display_name string `json:"display_name"`
}

// encodeQueryParam simple URL encoding for geocoding queries.
func encodeQueryParam(s string) string {
	// Basic encoding - in production use url.QueryEscape
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '.' || c == '_' || c == '~' {
			result += string(c)
		} else {
			result += fmt.Sprintf("%%%02x", byte(c))
		}
	}
	return result
}
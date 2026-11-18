package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a weather service client with proper timeout handling.
// Wraps external weather API integration.
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

// NewClient creates a new weather client with default timeouts.
// The HTTP client has a 10 second deadline, and individual operations
// should use context.WithTimeout for finer control.
func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// GetCurrentWeather fetches current weather for the given coordinates.
// ctx should have a timeout set (recommended: 10 seconds).
func (c *Client) GetCurrentWeather(ctx context.Context, lat, lng float64) (WeatherResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	params := fmt.Sprintf("?lat=%f&lng=%f&appid=%s", lat, lng, c.apiKey)
	url := c.baseURL + "/weather" + params

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		cancel()
		return WeatherResult{}, fmt.Errorf("weather: failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return WeatherResult{}, fmt.Errorf("weather: request timed out after 10 seconds: %w", err)
		}
		return WeatherResult{}, fmt.Errorf("weather: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return WeatherResult{}, fmt.Errorf("weather: returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result WeatherResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return WeatherResult{}, fmt.Errorf("weather: failed to decode response: %w", err)
	}

	return result, nil
}

// GetForecast fetches weather forecast for the given coordinates.
// ctx should have a timeout set (recommended: 15 seconds).
func (c *Client) GetForecast(ctx context.Context, lat, lng float64) (ForecastResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	params := fmt.Sprintf("?lat=%f&lng=%f&appid=%s&cnt=7", lat, lng, c.apiKey)
	url := c.baseURL + "/forecast" + params

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		cancel()
		return ForecastResult{}, fmt.Errorf("weather: failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ForecastResult{}, fmt.Errorf("weather: request timed out after 15 seconds: %w", err)
		}
		return ForecastResult{}, fmt.Errorf("weather: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return ForecastResult{}, fmt.Errorf("weather: returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result ForecastResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ForecastResult{}, fmt.Errorf("weather: failed to decode forecast response: %w", err)
	}

	return result, nil
}

// WeatherResult represents the current weather response.
type WeatherResult struct {
	ID          int     `json:"id"`
	Main        string  `json:"main"`
	Description string  `json:"description"`
	Temperature float64 `json:"temp"`
	Humidity    int     `json:"humidity"`
	Pressure    int     `json:"pressure"`
	WindSpeed   float64 `json:"wind_speed"`
	WindDeg     float64 `json:"wind_deg"`
}

// ForecastResult represents a 7-day weather forecast response.
type ForecastResult struct {
	List []ForecastEntry `json:"list"`
}

// ForecastEntry represents a single forecast entry (3-hour interval).
type ForecastEntry struct {
	DT        int     `json:"dt"`
	Main      MainEntry `json:"main"`
	Weather   []WeatherDesc `json:"weather"`
	Clouds    CloudsEntry `json:"clouds"`
	Wind      WindEntry `json:"wind"`
	Rain      *RainEntry `json:"rain,omitempty"`
	Snow      *SnowEntry `json:"snow,omitempty"`
}

// MainEntry represents the main weather data subset.
type MainEntry struct {
	Temp     float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin  float64 `json:"temp_min"`
	TempMax  float64 `json:"temp_max"`
	Pressure int     `json:"pressure"`
	Humidity int     `json:"humidity"`
}

// WeatherDesc represents a weather description.
type WeatherDesc struct {
	ID          int    `json:"id"`
	Main      string `json:"main"`
	Description string `json:"description"`
	Icon      string `json:"icon"`
}

// CloudsEntry represents cloud coverage data.
type CloudsEntry struct {
	All int `json:"all"`
}

// WindEntry represents wind data.
type WindEntry struct {
	Speed float64 `json:"speed"`
	Deg   float64 `json:"deg"`
}

// RainEntry represents rain volume (in the last 1h and 3h).
type RainEntry struct {
	Volume1H float64 `json:"1h"`
	Volume3H float64 `json:"3h"`
}

// SnowEntry represents snow volume (in the last 1h and 3h).
type SnowEntry struct {
	Volume1H float64 `json:"1h"`
	Volume3H float64 `json:"3h"`
}
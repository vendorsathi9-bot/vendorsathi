package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"vendorsaathi/internal/config"
	"vendorsaathi/internal/modules/mobility/model"
	"vendorsaathi/internal/modules/mobility/repository"
)

type MobilityService struct {
	repo          *repository.MobilityRepository
	mapCfg        config.MapConfig
	httpClient    *http.Client
	geocodeCache  sync.Map
	nominatimMu   sync.Mutex
	lastNominatim time.Time
}

func NewMobilityService(repo *repository.MobilityRepository, mapCfg config.MapConfig) *MobilityService {
	return &MobilityService{
		repo:   repo,
		mapCfg: mapCfg,
		httpClient: &http.Client{
			Timeout: 6 * time.Second,
		},
	}
}

// throttleNominatim enforces the OpenStreetMap Nominatim Usage Policy of maximum 1 request per second.
func (s *MobilityService) throttleNominatim(ctx context.Context) error {
	s.nominatimMu.Lock()
	defer s.nominatimMu.Unlock()

	elapsed := time.Since(s.lastNominatim)
	if elapsed < 1*time.Second {
		waitDuration := 1*time.Second - elapsed
		select {
		case <-time.After(waitDuration):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	s.lastNominatim = time.Now()
	return nil
}

// calculateHaversine calculates straight-line spherical distance in kilometers between two coordinates.
func calculateHaversine(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	rLat1 := lat1 * (math.Pi / 180.0)
	rLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(rLat1)*math.Cos(rLat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func (s *MobilityService) FindNearbyVendors(ctx context.Context, lat, lng, radiusKm float64) ([]model.NearbyVendor, error) {
	if radiusKm <= 0 {
		radiusKm = 5.0
	}
	if radiusKm > 50.0 {
		radiusKm = 50.0
	}
	return s.repo.FindNearbyVendors(ctx, lat, lng, radiusKm)
}

// OSRM response structs
type osrmRouteResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Routes  []struct {
		Distance float64 `json:"distance"` // in meters
		Duration float64 `json:"duration"` // in seconds
		Geometry struct {
			Type        string      `json:"type"`
			Coordinates [][]float64 `json:"coordinates"` // [[lng, lat], ...]
		} `json:"geometry"`
	} `json:"routes"`
}

// GetRoute requests a road route from OSRM using {longitude},{latitude} ordering.
func (s *MobilityService) GetRoute(ctx context.Context, req model.RouteRequest) (*model.RouteResponse, error) {
	origin := model.Coordinates{Latitude: req.OriginLat, Longitude: req.OriginLng}
	dest := model.Coordinates{Latitude: req.DestLat, Longitude: req.DestLng}

	osmDirectionsURL := fmt.Sprintf(
		"https://www.openstreetmap.org/directions?engine=fossgis_osrm_car&route=%f%%2C%f%%3B%f%%2C%f",
		req.OriginLat, req.OriginLng, req.DestLat, req.DestLng,
	)

	// Validate coordinates range
	if req.OriginLat < -90 || req.OriginLat > 90 || req.DestLat < -90 || req.DestLat > 90 ||
		req.OriginLng < -180 || req.OriginLng > 180 || req.DestLng < -180 || req.DestLng > 180 {
		return &model.RouteResponse{
			RoutingAvailable: false,
			Provider:         s.mapCfg.RoutingProvider,
			DistanceType:     "none",
			Origin:           origin,
			Destination:      dest,
			Attribution:      "© OpenStreetMap contributors, OSRM",
			Message:          "Invalid coordinate bounds provided",
			DirectionsURL:    osmDirectionsURL,
		}, nil
	}

	straightLineKm := math.Round(calculateHaversine(req.OriginLat, req.OriginLng, req.DestLat, req.DestLng)*100) / 100

	baseURL := s.mapCfg.OSRMBaseURL
	if baseURL == "" {
		baseURL = "https://router.project-osrm.org"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	// OSRM requires {longitude},{latitude};{longitude},{latitude}
	url := fmt.Sprintf(
		"%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		baseURL,
		req.OriginLng, req.OriginLat,
		req.DestLng, req.DestLat,
	)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return &model.RouteResponse{
			RoutingAvailable:          false,
			Provider:                  s.mapCfg.RoutingProvider,
			DistanceType:              "straight_line_estimate",
			ApproximateStraightLineKm: straightLineKm,
			Origin:                    origin,
			Destination:               dest,
			Attribution:               "© OpenStreetMap contributors, OSRM",
			Message:                   fmt.Sprintf("Failed to initialize routing request. Approximate straight-line distance is %.2f km (road travel ETA unavailable).", straightLineKm),
			DirectionsURL:             osmDirectionsURL,
		}, nil
	}
	httpReq.Header.Set("User-Agent", "VendorSaathi/1.0 (contact: admin@vendorsaathi.in)")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		// Timeout or network unreachable — return structured unavailable response without fabricated data
		return &model.RouteResponse{
			RoutingAvailable:          false,
			Provider:                  s.mapCfg.RoutingProvider,
			DistanceType:              "straight_line_estimate",
			ApproximateStraightLineKm: straightLineKm,
			Origin:                    origin,
			Destination:               dest,
			Attribution:               "© OpenStreetMap contributors, OSRM",
			Message:                   fmt.Sprintf("Road routing currently unavailable from provider (timeout or connection failed). Approximate straight-line distance is %.2f km (road travel ETA unavailable).", straightLineKm),
			DirectionsURL:             osmDirectionsURL,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &model.RouteResponse{
			RoutingAvailable:          false,
			Provider:                  s.mapCfg.RoutingProvider,
			DistanceType:              "straight_line_estimate",
			ApproximateStraightLineKm: straightLineKm,
			Origin:                    origin,
			Destination:               dest,
			Attribution:               "© OpenStreetMap contributors, OSRM",
			Message:                   fmt.Sprintf("Routing service returned non-200 status (%d). Approximate straight-line distance is %.2f km (road travel ETA unavailable).", resp.StatusCode, straightLineKm),
			DirectionsURL:             osmDirectionsURL,
		}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &model.RouteResponse{
			RoutingAvailable:          false,
			Provider:                  s.mapCfg.RoutingProvider,
			DistanceType:              "straight_line_estimate",
			ApproximateStraightLineKm: straightLineKm,
			Origin:                    origin,
			Destination:               dest,
			Attribution:               "© OpenStreetMap contributors, OSRM",
			Message:                   fmt.Sprintf("Failed to read routing provider response. Approximate straight-line distance is %.2f km.", straightLineKm),
			DirectionsURL:             osmDirectionsURL,
		}, nil
	}

	var osrmResp osrmRouteResponse
	if err := json.Unmarshal(body, &osrmResp); err != nil {
		return &model.RouteResponse{
			RoutingAvailable:          false,
			Provider:                  s.mapCfg.RoutingProvider,
			DistanceType:              "straight_line_estimate",
			ApproximateStraightLineKm: straightLineKm,
			Origin:                    origin,
			Destination:               dest,
			Attribution:               "© OpenStreetMap contributors, OSRM",
			Message:                   fmt.Sprintf("Invalid response format from routing provider. Approximate straight-line distance is %.2f km.", straightLineKm),
			DirectionsURL:             osmDirectionsURL,
		}, nil
	}

	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return &model.RouteResponse{
			RoutingAvailable:          false,
			Provider:                  s.mapCfg.RoutingProvider,
			DistanceType:              "straight_line_estimate",
			ApproximateStraightLineKm: straightLineKm,
			Origin:                    origin,
			Destination:               dest,
			Attribution:               "© OpenStreetMap contributors, OSRM",
			Message:                   fmt.Sprintf("No road route found: %s. Approximate straight-line distance is %.2f km (road travel ETA unavailable).", osrmResp.Message, straightLineKm),
			DirectionsURL:             osmDirectionsURL,
		}, nil
	}

	bestRoute := osrmResp.Routes[0]
	distMeters := math.Round(bestRoute.Distance*10) / 10
	durSeconds := math.Round(bestRoute.Duration*10) / 10
	distKm := math.Round((bestRoute.Distance/1000.0)*100) / 100
	durMinutes := math.Max(1.0, math.Round((bestRoute.Duration/60.0)*10)/10)

	geometry := model.GeoJSONGeometry{
		Type:        bestRoute.Geometry.Type,
		Coordinates: bestRoute.Geometry.Coordinates,
	}

	return &model.RouteResponse{
		RoutingAvailable:          true,
		Provider:                  s.mapCfg.RoutingProvider,
		DistanceType:              "road",
		DistanceMeters:            distMeters,
		DurationSeconds:           durSeconds,
		Geometry:                  &geometry,
		Origin:                    origin,
		Destination:               dest,
		Attribution:               "© OpenStreetMap contributors, OSRM",
		DistanceKm:                distKm,
		DurationMinutes:           durMinutes,
		ApproximateStraightLineKm: straightLineKm,
		DirectionsURL:             osmDirectionsURL,
	}, nil
}

// Geocode uses Nominatim to search addresses (with in-memory cache to respect usage policies).
// Enforces a strict rate-limit of maximum 1 request per second to the public Nominatim endpoint.
func (s *MobilityService) Geocode(ctx context.Context, query string) (*model.GeocodeResult, error) {
	cacheKey := "geo:" + strings.ToLower(strings.TrimSpace(query))
	if cached, ok := s.geocodeCache.Load(cacheKey); ok {
		if res, ok := cached.(*model.GeocodeResult); ok {
			return res, nil
		}
	}

	if err := s.throttleNominatim(ctx); err != nil {
		return nil, err
	}

	baseURL := s.mapCfg.GeocodingBaseURL
	if baseURL == "" {
		baseURL = "https://nominatim.openstreetmap.org"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	url := fmt.Sprintf("%s/search?q=%s&format=json&limit=1", baseURL, query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "VendorSaathi/1.0 (contact: admin@vendorsaathi.in)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("geocoding rate limit exceeded, please retry later")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding service returned status %d", resp.StatusCode)
	}

	var results []struct {
		DisplayName string `json:"display_name"`
		Lat         string `json:"lat"`
		Lon         string `json:"lon"`
		Type        string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no geocoding results found")
	}

	first := results[0]
	lat, _ := strconv.ParseFloat(first.Lat, 64)
	lng, _ := strconv.ParseFloat(first.Lon, 64)

	res := &model.GeocodeResult{
		DisplayName: first.DisplayName,
		Latitude:    lat,
		Longitude:   lng,
		Type:        first.Type,
		Attribution: "© OpenStreetMap contributors, Nominatim",
	}

	s.geocodeCache.Store(cacheKey, res)
	return res, nil
}

// ReverseGeocode uses Nominatim for reverse geocoding (with caching).
// Enforces a strict rate-limit of maximum 1 request per second to the public Nominatim endpoint.
func (s *MobilityService) ReverseGeocode(ctx context.Context, lat, lng float64) (*model.GeocodeResult, error) {
	cacheKey := fmt.Sprintf("rev:%.5f,%.5f", lat, lng)
	if cached, ok := s.geocodeCache.Load(cacheKey); ok {
		if res, ok := cached.(*model.GeocodeResult); ok {
			return res, nil
		}
	}

	if err := s.throttleNominatim(ctx); err != nil {
		return nil, err
	}

	baseURL := s.mapCfg.GeocodingBaseURL
	if baseURL == "" {
		baseURL = "https://nominatim.openstreetmap.org"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	url := fmt.Sprintf("%s/reverse?lat=%f&lon=%f&format=json", baseURL, lat, lng)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "VendorSaathi/1.0 (contact: admin@vendorsaathi.in)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("reverse geocoding rate limit exceeded, please retry later")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reverse geocoding service returned status %d", resp.StatusCode)
	}

	var result struct {
		DisplayName string `json:"display_name"`
		Type        string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	res := &model.GeocodeResult{
		DisplayName: result.DisplayName,
		Latitude:    lat,
		Longitude:   lng,
		Type:        result.Type,
		Attribution: "© OpenStreetMap contributors, Nominatim",
	}

	s.geocodeCache.Store(cacheKey, res)
	return res, nil
}

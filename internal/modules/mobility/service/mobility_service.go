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

type cachedTile struct {
	data        []byte
	contentType string
	etag        string
	cachedAt    time.Time
}

type MobilityService struct {
	repo          *repository.MobilityRepository
	mapCfg        config.MapConfig
	httpClient    *http.Client
	geocodeCache  sync.Map
	nominatimMu   sync.Mutex
	lastNominatim time.Time
	tileCacheMu   sync.RWMutex
	tileCache     map[string]cachedTile
}

func NewMobilityService(repo *repository.MobilityRepository, mapCfg config.MapConfig) *MobilityService {
	return &MobilityService{
		repo:   repo,
		mapCfg: mapCfg,
		httpClient: &http.Client{
			Timeout: 6 * time.Second,
		},
		tileCache: make(map[string]cachedTile),
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

// GetTile proxies and caches a map tile with proper User-Agent identification and caching headers.
// Respects OpenStreetMap tile usage policy by using identifying User-Agent, in-memory caching,
// and handling 304 Not Modified.
func (s *MobilityService) GetTile(ctx context.Context, z, x, y int, ifNoneMatch string) ([]byte, string, string, bool, error) {
	if z < 0 || z > 20 || x < 0 || y < 0 {
		return nil, "", "", false, fmt.Errorf("tile coordinates out of range")
	}
	maxCoord := 1 << z
	if x >= maxCoord || y >= maxCoord {
		return nil, "", "", false, fmt.Errorf("tile coordinates out of bounds for zoom level %d", z)
	}

	cacheKey := fmt.Sprintf("%d/%d/%d", z, x, y)

	// Check in-memory cache
	s.tileCacheMu.RLock()
	cached, found := s.tileCache[cacheKey]
	s.tileCacheMu.RUnlock()

	if found && time.Since(cached.cachedAt) < 24*time.Hour {
		if ifNoneMatch != "" && ifNoneMatch == cached.etag {
			return nil, "", cached.etag, true, nil
		}
		return cached.data, cached.contentType, cached.etag, false, nil
	}

	// Determine upstream URL. Raster proxy usage is optional; MapLibre clients
	// use the vector style URL from /mobility/config instead.
	template := s.mapCfg.UpstreamTileURL
	if template == "" {
		template = s.mapCfg.TileURL
	}
	if template == "" {
		return nil, "", "", false, fmt.Errorf("no raster tile provider configured: set MAP_UPSTREAM_TILE_URL or MAP_TILE_URL")
	}

	tileURL := strings.Replace(template, "{z}", strconv.Itoa(z), 1)
	tileURL = strings.Replace(tileURL, "{x}", strconv.Itoa(x), 1)
	tileURL = strings.Replace(tileURL, "{y}", strconv.Itoa(y), 1)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tileURL, nil)
	if err != nil {
		return nil, "", "", false, fmt.Errorf("failed to create tile request: %w", err)
	}

	userAgent := s.mapCfg.UserAgent
	if userAgent == "" {
		userAgent = "VendorSaathi/1.0 (+https://vendorsaathi.in; contact: admin@vendorsaathi.in)"
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "image/png,image/jpeg,image/*")
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", "", false, fmt.Errorf("failed to fetch tile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return nil, "", ifNoneMatch, true, nil
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, "", "", false, fmt.Errorf("tile provider access denied (HTTP 403): %s", resp.Status)
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, "", "", false, fmt.Errorf("tile provider rate limit exceeded (HTTP 429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, "", "", false, fmt.Errorf("tile provider returned HTTP %d", resp.StatusCode)
	}

	// Limit tile size to 1MB
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, "", "", false, fmt.Errorf("failed to read tile data: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}
	etag := resp.Header.Get("ETag")
	if etag == "" {
		etag = fmt.Sprintf(`"%x-%d"`, len(data), time.Now().Unix())
	}

	// Store in cache with capacity limit (max 2000 tiles)
	s.tileCacheMu.Lock()
	if len(s.tileCache) > 2000 {
		now := time.Now()
		for k, v := range s.tileCache {
			if now.Sub(v.cachedAt) > 12*time.Hour {
				delete(s.tileCache, k)
			}
		}
		if len(s.tileCache) > 2000 {
			s.tileCache = make(map[string]cachedTile)
		}
	}
	s.tileCache[cacheKey] = cachedTile{
		data:        data,
		contentType: contentType,
		etag:        etag,
		cachedAt:    time.Now(),
	}
	s.tileCacheMu.Unlock()

	return data, contentType, etag, false, nil
}

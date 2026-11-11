package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"vendorsaathi/internal/config"
)

func TestGetTile_OutOfBounds(t *testing.T) {
	svc := NewMobilityService(nil, config.MapConfig{
		Provider: "carto",
		TileURL:  "https://a.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}.png",
	})

	// Invalid zoom
	_, _, _, _, err := svc.GetTile(context.Background(), -1, 0, 0, "")
	if err == nil {
		t.Fatal("expected error for zoom < 0")
	}

	// Coordinates exceed zoom max (z=1 has max coord 2)
	_, _, _, _, err = svc.GetTile(context.Background(), 1, 5, 0, "")
	if err == nil {
		t.Fatal("expected error for x out of bounds")
	}
}

func TestGetTile_ProxyAndCache(t *testing.T) {
	requestsCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsCount++
		// Verify User-Agent was set
		if r.Header.Get("User-Agent") == "" {
			t.Errorf("expected User-Agent header to be set")
		}

		if r.Header.Get("If-None-Match") == `"test-etag"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("ETag", `"test-etag"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mock-png-data"))
	}))
	defer mockServer.Close()

	svc := NewMobilityService(nil, config.MapConfig{
		Provider:        "osm_proxy",
		UpstreamTileURL: mockServer.URL + "/{z}/{x}/{y}.png",
		UserAgent:       "VendorSaathi-Test/1.0",
	})

	// 1. Initial request (miss, fetches from upstream)
	data, contentType, etag, notMod, err := svc.GetTile(context.Background(), 2, 1, 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notMod {
		t.Fatal("expected notMod to be false on initial fetch")
	}
	if string(data) != "mock-png-data" {
		t.Fatalf("expected data 'mock-png-data', got '%s'", string(data))
	}
	if contentType != "image/png" {
		t.Fatalf("expected 'image/png', got '%s'", contentType)
	}
	if etag != `"test-etag"` {
		t.Fatalf("expected '\"test-etag\"', got '%s'", etag)
	}
	if requestsCount != 1 {
		t.Fatalf("expected 1 upstream request, got %d", requestsCount)
	}

	// 2. Second request from cache (should NOT hit upstream)
	dataCached, _, _, notMod2, err := svc.GetTile(context.Background(), 2, 1, 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notMod2 {
		t.Fatal("expected notMod to be false")
	}
	if string(dataCached) != "mock-png-data" {
		t.Fatalf("expected cached data")
	}
	if requestsCount != 1 {
		t.Fatalf("expected still 1 upstream request due to cache, got %d", requestsCount)
	}

	// 3. ETag 304 check
	_, _, etag304, notMod3, err := svc.GetTile(context.Background(), 2, 1, 1, `"test-etag"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !notMod3 {
		t.Fatal("expected notMod3 to be true for matching ETag")
	}
	if etag304 != `"test-etag"` {
		t.Fatalf("expected matching ETag")
	}
}

func TestGetTile_Upstream403(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Access blocked - violation of tile usage policy"))
	}))
	defer mockServer.Close()

	svc := NewMobilityService(nil, config.MapConfig{
		Provider:        "osm_proxy",
		UpstreamTileURL: mockServer.URL + "/{z}/{x}/{y}.png",
		UserAgent:       "VendorSaathi-Test/1.0",
	})

	_, _, _, _, err := svc.GetTile(context.Background(), 2, 1, 1, "")
	if err == nil {
		t.Fatal("expected error on upstream 403")
	}
	if err.Error() == "" {
		t.Fatal("expected descriptive error message")
	}
}

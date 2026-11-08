import React, { useEffect, useRef, useState } from 'react';
import * as maplibregl from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { ApiClient, DeliveryAssignment, DeliveryLocation, RouteData } from '../../core/api/apiClient';
import { X, Navigation, MapPin, Clock, AlertTriangle, RefreshCw } from 'lucide-react';

interface DeliveryMapModalProps {
  assignment: DeliveryAssignment;
  onClose: () => void;
}

export const DeliveryMapModal: React.FC<DeliveryMapModalProps> = ({ assignment, onClose }) => {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<maplibregl.Map | null>(null);
  const markersRef = useRef<maplibregl.Marker[]>([]);

  const [agentLoc, setAgentLoc] = useState<DeliveryLocation | null>(null);
  const [routeData, setRouteData] = useState<RouteData | null>(null);
  const [loading, setLoading] = useState(true);
  const [routeLoading, setRouteLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Vendor coords (fallback to Bangalore Majestic if unset)
  const vendorLat = assignment.vendor_lat && assignment.vendor_lat !== 0 ? assignment.vendor_lat : 12.9778;
  const vendorLng = assignment.vendor_lng && assignment.vendor_lng !== 0 ? assignment.vendor_lng : 77.5727;

  // Customer coords (fallback to nearby destination)
  const customerLat = assignment.delivery_lat && assignment.delivery_lat !== 0 ? assignment.delivery_lat : 12.9352;
  const customerLng = assignment.delivery_lng && assignment.delivery_lng !== 0 ? assignment.delivery_lng : 77.6245;

  const loadData = async () => {
    try {
      setLoading(true);
      setErrorMessage(null);

      // 1. Fetch live agent location if available
      try {
        const loc = await ApiClient.getDeliveryLocation(assignment.order_id);
        if (loc && loc.latitude && loc.longitude) {
          setAgentLoc(loc);
        }
      } catch {
        // Agent may not have broadcasted yet
      }

      // 2. Fetch OSRM Road Route
      setRouteLoading(true);
      try {
        const route = await ApiClient.getRoute(vendorLat, vendorLng, customerLat, customerLng);
        setRouteData(route);
        if (!route.routing_available) {
          setErrorMessage(route.message || 'Road routing currently unavailable from OSRM provider. Showing approximate straight-line estimate.');
        }
      } catch (e: any) {
        setErrorMessage('OSRM road routing currently unavailable. Displaying location markers.');
      } finally {
        setRouteLoading(false);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();

    // Poll for live agent location updates every 15s
    const pollInterval = setInterval(async () => {
      try {
        const loc = await ApiClient.getDeliveryLocation(assignment.order_id);
        if (loc && loc.latitude && loc.longitude) {
          setAgentLoc(loc);
        }
      } catch {
        // No update yet
      }
    }, 15000);

    return () => clearInterval(pollInterval);
  }, [assignment.order_id]);

  // Initialize and update MapLibre GL map
  useEffect(() => {
    if (!mapContainerRef.current) return;

    // Get configuration or use standard OSM tiles
    const tileUrl = 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';

    const map = new maplibregl.Map({
      container: mapContainerRef.current,
      style: {
        version: 8,
        sources: {
          'osm-tiles': {
            type: 'raster',
            tiles: [tileUrl],
            tileSize: 256,
            attribution: '© <a href="https://www.openstreetmap.org/copyright" target="_blank">OpenStreetMap</a> contributors, MapLibre',
          },
        },
        layers: [
          {
            id: 'osm-tiles-layer',
            type: 'raster',
            source: 'osm-tiles',
            minzoom: 0,
            maxzoom: 19,
          },
        ],
      },
      center: [vendorLng, vendorLat],
      zoom: 13,
    });

    mapRef.current = map;
    map.addControl(new maplibregl.NavigationControl(), 'top-right');

    map.on('load', () => {
      renderMapFeatures(map);
    });

    return () => {
      markersRef.current.forEach(m => m.remove());
      markersRef.current = [];
      map.remove();
      mapRef.current = null;
    };
  }, []);

  // Update map features when routeData or agentLoc changes
  useEffect(() => {
    if (mapRef.current && mapRef.current.isStyleLoaded()) {
      renderMapFeatures(mapRef.current);
    }
  }, [routeData, agentLoc]);

  const renderMapFeatures = (map: maplibregl.Map) => {
    // Clear old markers
    markersRef.current.forEach(m => m.remove());
    markersRef.current = [];

    const bounds = new maplibregl.LngLatBounds();

    // 1. Vendor Pickup Marker (Orange)
    const vendorEl = document.createElement('div');
    vendorEl.innerHTML = `
      <div style="background-color: #ea580c; color: white; width: 34px; height: 34px; border-radius: 50%; display: flex; align-items: center; justify-content: center; box-shadow: 0 2px 6px rgba(0,0,0,0.3); border: 2px solid white; cursor: pointer;">
        🏪
      </div>
    `;
    const vendorMarker = new maplibregl.Marker({ element: vendorEl })
      .setLngLat([vendorLng, vendorLat])
      .setPopup(new maplibregl.Popup({ offset: 25 }).setHTML(`<strong>Pickup:</strong> ${assignment.vendor_name || 'Vendor'}`))
      .addTo(map);
    markersRef.current.push(vendorMarker);
    bounds.extend([vendorLng, vendorLat]);

    // 2. Customer Destination Marker (Red)
    const customerEl = document.createElement('div');
    customerEl.innerHTML = `
      <div style="background-color: #dc2626; color: white; width: 34px; height: 34px; border-radius: 50%; display: flex; align-items: center; justify-content: center; box-shadow: 0 2px 6px rgba(0,0,0,0.3); border: 2px solid white; cursor: pointer;">
        📍
      </div>
    `;
    const customerMarker = new maplibregl.Marker({ element: customerEl })
      .setLngLat([customerLng, customerLat])
      .setPopup(new maplibregl.Popup({ offset: 25 }).setHTML(`<strong>Delivery:</strong> ${assignment.customer_name || 'Customer'}<br/>${assignment.delivery_address_line1 || ''}`))
      .addTo(map);
    markersRef.current.push(customerMarker);
    bounds.extend([customerLng, customerLat]);

    // 3. Delivery Agent Marker (Green/Blue) if location exists
    if (agentLoc && agentLoc.latitude && agentLoc.longitude) {
      const agentEl = document.createElement('div');
      agentEl.innerHTML = `
        <div style="background-color: #2563eb; color: white; width: 38px; height: 38px; border-radius: 50%; display: flex; align-items: center; justify-content: center; box-shadow: 0 3px 8px rgba(37,99,235,0.4); border: 3px solid white; cursor: pointer; animation: pulse 2s infinite;">
          🛵
        </div>
      `;
      const agentMarker = new maplibregl.Marker({ element: agentEl })
        .setLngLat([agentLoc.longitude, agentLoc.latitude])
        .setPopup(new maplibregl.Popup({ offset: 25 }).setHTML(`<strong>Agent Location</strong><br/>Updated: ${new Date(agentLoc.updated_at).toLocaleTimeString()}`))
        .addTo(map);
      markersRef.current.push(agentMarker);
      bounds.extend([agentLoc.longitude, agentLoc.latitude]);
    }

    // 4. Render OSRM Road Route Geometry
    if (routeData?.geometry?.coordinates && routeData.geometry.coordinates.length > 0) {
      if (map.getSource('osrm-route')) {
        (map.getSource('osrm-route') as maplibregl.GeoJSONSource).setData({
          type: 'Feature',
          properties: {},
          geometry: {
            type: 'LineString',
            coordinates: routeData.geometry.coordinates,
          },
        });
      } else {
        map.addSource('osrm-route', {
          type: 'geojson',
          data: {
            type: 'Feature',
            properties: {},
            geometry: {
              type: 'LineString',
              coordinates: routeData.geometry.coordinates,
            },
          },
        });

        map.addLayer({
          id: 'osrm-route-casing',
          type: 'line',
          source: 'osrm-route',
          layout: { 'line-join': 'round', 'line-cap': 'round' },
          paint: { 'line-color': '#1d4ed8', 'line-width': 7, 'line-opacity': 0.8 },
        });

        map.addLayer({
          id: 'osrm-route-inner',
          type: 'line',
          source: 'osrm-route',
          layout: { 'line-join': 'round', 'line-cap': 'round' },
          paint: { 'line-color': '#60a5fa', 'line-width': 4 },
        });
      }

      // Extend bounds to encompass all route waypoints
      routeData.geometry.coordinates.forEach(coord => {
        bounds.extend(coord as [number, number]);
      });
    }

    // Fit camera to bounds
    if (!bounds.isEmpty()) {
      map.fitBounds(bounds, { padding: 60, maxZoom: 16 });
    }
  };

  const osmUrl = `https://www.openstreetmap.org/directions?engine=fossgis_osrm_car&route=${vendorLat}%2C${vendorLng}%3B${customerLat}%2C${customerLng}`;

  return (
    <div style={{
      position: 'fixed',
      top: 0, left: 0, right: 0, bottom: 0,
      backgroundColor: 'rgba(15, 23, 42, 0.75)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 1100,
      padding: '1.5rem',
    }}>
      <div style={{
        backgroundColor: 'white',
        borderRadius: '1rem',
        width: '100%',
        maxWidth: '920px',
        maxHeight: '90vh',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
      }}>
        {/* Header */}
        <div style={{
          padding: '1.25rem 1.5rem',
          borderBottom: '1px solid #e2e8f0',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          backgroundColor: '#f8fafc',
        }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Navigation size={20} color="#2563eb" />
              <h2 style={{ fontSize: '1.25rem', fontWeight: 700, color: '#0f172a', margin: 0 }}>
                Live Delivery Route & MapLibre Tracking
              </h2>
            </div>
            <p style={{ margin: '0.25rem 0 0 0', fontSize: '0.85rem', color: '#64748b' }}>
              Order #{assignment.order_id.slice(0, 8)} • OpenStreetMap & OSRM Road Routing
            </p>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            {loading && (
              <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Updating...</span>
            )}
            <button
              onClick={loadData}
              title="Refresh"
              disabled={loading}
              style={{
                backgroundColor: 'white',
                border: '1px solid #cbd5e1',
                borderRadius: '0.375rem',
                padding: '0.5rem',
                cursor: loading ? 'not-allowed' : 'pointer',
              }}
            >
              <RefreshCw size={16} color={loading ? '#2563eb' : '#64748b'} />
            </button>
            <button
              onClick={onClose}
              style={{
                backgroundColor: 'transparent',
                border: 'none',
                cursor: 'pointer',
                padding: '0.5rem',
                borderRadius: '0.375rem',
                color: '#64748b',
              }}
            >
              <X size={20} />
            </button>
          </div>
        </div>

        {/* Route Stats Bar */}
        <div style={{
          padding: '0.75rem 1.5rem',
          backgroundColor: '#eff6ff',
          borderBottom: '1px solid #bfdbfe',
          display: 'flex',
          flexWrap: 'wrap',
          gap: '1.5rem',
          fontSize: '0.875rem',
          alignItems: 'center',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
            <Navigation size={16} color="#2563eb" />
            <span style={{ color: '#1e40af' }}>Distance:</span>
            <strong style={{ color: '#1e3a8a' }}>
              {routeLoading ? (
                'Calculating...'
              ) : routeData?.routing_available ? (
                `${routeData.distance_km} km (${Math.round(routeData.distance_meters)} m)`
              ) : routeData?.approximate_straight_line_km ? (
                `Approx. ${routeData.approximate_straight_line_km} km (Straight-line estimate)`
              ) : (
                'Distance unavailable'
              )}
            </strong>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
            <Clock size={16} color="#2563eb" />
            <span style={{ color: '#1e40af' }}>Estimated Duration:</span>
            <strong style={{ color: '#1e3a8a' }}>
              {routeLoading ? (
                'Calculating...'
              ) : routeData?.routing_available ? (
                `${routeData.duration_minutes} min (${Math.round(routeData.duration_seconds)} sec)`
              ) : (
                'ETA unavailable without road routing'
              )}
            </strong>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
            <span style={{
              display: 'inline-block',
              width: 8,
              height: 8,
              borderRadius: '50%',
              backgroundColor: agentLoc ? '#16a34a' : '#f59e0b',
            }} />
            <span style={{ color: '#1e40af' }}>Agent Status:</span>
            <strong style={{ color: '#1e3a8a' }}>
              {agentLoc ? `Live GPS (Updated ${new Date(agentLoc.updated_at).toLocaleTimeString()})` : 'Waiting for GPS ping'}
            </strong>
          </div>
        </div>

        {/* Notification Banner if provider notice */}
        {errorMessage && (
          <div style={{
            backgroundColor: '#fffbeb',
            borderBottom: '1px solid #fef3c7',
            padding: '0.5rem 1.5rem',
            fontSize: '0.8rem',
            color: '#92400e',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}>
            <AlertTriangle size={14} color="#d97706" />
            <span>{errorMessage}</span>
          </div>
        )}

        {/* MapLibre Map Container */}
        <div style={{ position: 'relative', flex: 1, minHeight: '440px', backgroundColor: '#e2e8f0' }}>
          <div ref={mapContainerRef} style={{ width: '100%', height: '100%', minHeight: '440px' }} />

          {/* Map Attribution overlay */}
          <div style={{
            position: 'absolute',
            bottom: '6px',
            right: '6px',
            backgroundColor: 'rgba(255,255,255,0.85)',
            padding: '2px 8px',
            borderRadius: '4px',
            fontSize: '10px',
            color: '#475569',
            pointerEvents: 'none',
            zIndex: 10,
          }}>
            © <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noreferrer" style={{ color: '#2563eb' }}>OpenStreetMap</a> contributors • MapLibre GL • OSRM
          </div>
        </div>

        {/* Footer with Waypoint Info & OSM Link */}
        <div style={{
          padding: '1rem 1.5rem',
          borderTop: '1px solid #e2e8f0',
          display: 'flex',
          flexWrap: 'wrap',
          justifyContent: 'space-between',
          alignItems: 'center',
          backgroundColor: '#f8fafc',
          gap: '0.75rem',
        }}>
          <div style={{ display: 'flex', gap: '1.25rem', fontSize: '0.8rem', color: '#475569' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
              <span style={{ display: 'inline-block', width: 10, height: 10, borderRadius: '50%', backgroundColor: '#ea580c' }} />
              <strong>Pickup:</strong> {assignment.vendor_name || 'Vendor'}
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
              <span style={{ display: 'inline-block', width: 10, height: 10, borderRadius: '50%', backgroundColor: '#dc2626' }} />
              <strong>Destination:</strong> {assignment.delivery_city || 'Customer Destination'}
            </div>
            {agentLoc && (
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                <span style={{ display: 'inline-block', width: 10, height: 10, borderRadius: '50%', backgroundColor: '#2563eb' }} />
                <strong>Agent:</strong> {agentLoc.latitude.toFixed(4)}, {agentLoc.longitude.toFixed(4)}
              </div>
            )}
          </div>

          <div style={{ display: 'flex', gap: '0.75rem' }}>
            <a
              href={osmUrl}
              target="_blank"
              rel="noreferrer"
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.35rem',
                fontSize: '0.8rem',
                color: '#2563eb',
                textDecoration: 'none',
                padding: '0.4rem 0.75rem',
                borderRadius: '0.375rem',
                border: '1px solid #cbd5e1',
                backgroundColor: 'white',
                fontWeight: 600,
              }}
            >
              <MapPin size={14} />
              Open in OpenStreetMap ↗
            </a>
            <button
              onClick={onClose}
              style={{
                padding: '0.4rem 1rem',
                backgroundColor: '#0f172a',
                color: 'white',
                border: 'none',
                borderRadius: '0.375rem',
                fontSize: '0.8rem',
                fontWeight: 600,
                cursor: 'pointer',
              }}
            >
              Close Map
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

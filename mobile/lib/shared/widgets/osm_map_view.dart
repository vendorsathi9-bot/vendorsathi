import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:maplibre_gl/maplibre_gl.dart';
import 'package:url_launcher/url_launcher.dart';
import '../models/models.dart';

/// Embedded Open-Source Map powered by MapLibre GL and OpenStreetMap tiles.
/// Complies with OpenStreetMap tile usage and routing policies.
class OsmMapView extends StatefulWidget {
  final double? initialLat;
  final double? initialLng;
  final double initialZoom;
  final double? vendorLat;
  final double? vendorLng;
  final String? vendorTitle;
  final double? customerLat;
  final double? customerLng;
  final String? customerTitle;
  final double? agentLat;
  final double? agentLng;
  final String? agentTitle;
  final RouteModel? route;
  final List<NearbyVendorModel>? nearbyVendors;
  final void Function(NearbyVendorModel)? onVendorSelected;
  final double height;
  final bool interactive;
  final String? tileUrl;

  const OsmMapView({
    super.key,
    this.initialLat,
    this.initialLng,
    this.initialZoom = 13.0,
    this.vendorLat,
    this.vendorLng,
    this.vendorTitle,
    this.customerLat,
    this.customerLng,
    this.customerTitle,
    this.agentLat,
    this.agentLng,
    this.agentTitle,
    this.route,
    this.nearbyVendors,
    this.onVendorSelected,
    this.height = 280,
    this.interactive = true,
    this.tileUrl,
  });

  @override
  State<OsmMapView> createState() => _OsmMapViewState();
}

class _OsmMapViewState extends State<OsmMapView> {
  MapLibreMapController? _mapController;
  bool _renderError = false;

  // OpenStreetMap raster tile style spec with configurable tile URL
  String get _osmRasterStyle {
    final effectiveTileUrl = widget.tileUrl ?? 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';
    return '''
    {
      "version": 8,
      "sources": {
        "osm-raster-tiles": {
          "type": "raster",
          "tiles": [
            "$effectiveTileUrl"
          ],
          "tileSize": 256,
          "attribution": "© OpenStreetMap contributors"
        }
      },
      "layers": [
        {
          "id": "osm-tiles-layer",
          "type": "raster",
          "source": "osm-raster-tiles",
          "minzoom": 0,
          "maxzoom": 19
        }
      ]
    }
    ''';
  }

  double get _centerLat {
    if (widget.agentLat != null && widget.agentLat != 0) return widget.agentLat!;
    if (widget.vendorLat != null && widget.vendorLat != 0) return widget.vendorLat!;
    if (widget.customerLat != null && widget.customerLat != 0) return widget.customerLat!;
    return widget.initialLat ?? 12.9716; // Bangalore default
  }

  double get _centerLng {
    if (widget.agentLng != null && widget.agentLng != 0) return widget.agentLng!;
    if (widget.vendorLng != null && widget.vendorLng != 0) return widget.vendorLng!;
    if (widget.customerLng != null && widget.customerLng != 0) return widget.customerLng!;
    return widget.initialLng ?? 77.5946;
  }

  void _onMapCreated(MapLibreMapController controller) {
    _mapController = controller;
  }

  Future<void> _onStyleLoaded() async {
    if (!mounted || _mapController == null) return;

    try {
      // 1. Draw OSRM route line if route coordinates exist
      if (widget.route != null && widget.route!.coordinates.isNotEmpty) {
        final lineCoords = widget.route!.coordinates
            .map((c) => LatLng(c[1], c[0])) // c[0]=lng, c[1]=lat -> LatLng(lat, lng)
            .toList();

        // Thick casing layer
        await _mapController!.addLine(LineOptions(
          geometry: lineCoords,
          lineColor: '#1d4ed8',
          lineWidth: 5.0,
          lineOpacity: 0.9,
        ));
      }

      // 2. Add Vendor marker / circle
      if (widget.vendorLat != null && widget.vendorLat != 0 && widget.vendorLng != null && widget.vendorLng != 0) {
        await _mapController!.addCircle(CircleOptions(
          geometry: LatLng(widget.vendorLat!, widget.vendorLng!),
          circleRadius: 9.0,
          circleColor: '#2563eb',
          circleStrokeWidth: 2.0,
          circleStrokeColor: '#ffffff',
        ));
      }

      // 3. Add Customer marker / circle
      if (widget.customerLat != null && widget.customerLat != 0 && widget.customerLng != null && widget.customerLng != 0) {
        await _mapController!.addCircle(CircleOptions(
          geometry: LatLng(widget.customerLat!, widget.customerLng!),
          circleRadius: 9.0,
          circleColor: '#dc2626',
          circleStrokeWidth: 2.0,
          circleStrokeColor: '#ffffff',
        ));
      }

      // 4. Add Live Delivery Agent marker / circle
      if (widget.agentLat != null && widget.agentLat != 0 && widget.agentLng != null && widget.agentLng != 0) {
        await _mapController!.addCircle(CircleOptions(
          geometry: LatLng(widget.agentLat!, widget.agentLng!),
          circleRadius: 10.0,
          circleColor: '#16a34a',
          circleStrokeWidth: 3.0,
          circleStrokeColor: '#ffffff',
        ));
      }

      // 5. Add Nearby Vendors markers
      if (widget.nearbyVendors != null && widget.nearbyVendors!.isNotEmpty) {
        for (final v in widget.nearbyVendors!) {
          if (v.latitude != 0 && v.longitude != 0) {
            await _mapController!.addCircle(CircleOptions(
              geometry: LatLng(v.latitude, v.longitude),
              circleRadius: 8.0,
              circleColor: '#0284c7',
              circleStrokeWidth: 2.0,
              circleStrokeColor: '#ffffff',
            ));
          }
        }
      }
    } catch (e) {
      debugPrint('OsmMapView style decoration warning: $e');
    }
  }

  Future<void> _openExternalOSM() async {
    final fromLat = widget.agentLat ?? widget.vendorLat ?? _centerLat;
    final fromLng = widget.agentLng ?? widget.vendorLng ?? _centerLng;
    final toLat = widget.customerLat ?? _centerLat;
    final toLng = widget.customerLng ?? _centerLng;

    final url = 'https://www.openstreetmap.org/directions?engine=fossgis_osrm_car&route=$fromLat%2C$fromLng%3B$toLat%2C$toLng';
    final uri = Uri.parse(url);
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } else {
        await launchUrl(uri, mode: LaunchMode.platformDefault);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Could not open map: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      height: widget.height,
      decoration: BoxDecoration(
        color: const Color(0xFFF1F5F9),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFCBD5E1)),
      ),
      clipBehavior: Clip.antiAlias,
      child: Stack(
        children: [
          // Map view or Fallback
          if (!_renderError)
            Builder(
              builder: (ctx) {
                try {
                  return MapLibreMap(
                    styleString: _osmRasterStyle,
                    initialCameraPosition: CameraPosition(
                      target: LatLng(_centerLat, _centerLng),
                      zoom: widget.initialZoom,
                    ),
                    onMapCreated: _onMapCreated,
                    onStyleLoadedCallback: _onStyleLoaded,
                    myLocationEnabled: false,
                    trackCameraPosition: true,
                    attributionButtonMargins: const math.Point(-100, -100), // Hide default attribution button to use custom label
                  );
                } catch (e) {
                  WidgetsBinding.instance.addPostFrameCallback((_) {
                    if (mounted) setState(() => _renderError = true);
                  });
                  return _buildFallbackView();
                }
              },
            )
          else
            _buildFallbackView(),

          // Legend Overlay
          Positioned(
            top: 8,
            left: 8,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.92),
                borderRadius: BorderRadius.circular(6),
                boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 4)],
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (widget.vendorLat != null) ...[
                    const Icon(Icons.storefront, size: 14, color: Color(0xFF2563EB)),
                    const SizedBox(width: 3),
                    const Text('Vendor', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold)),
                    const SizedBox(width: 8),
                  ],
                  if (widget.agentLat != null) ...[
                    const Icon(Icons.two_wheeler, size: 14, color: Color(0xFF16A34A)),
                    const SizedBox(width: 3),
                    const Text('Agent', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Color(0xFF16A34A))),
                    const SizedBox(width: 8),
                  ],
                  if (widget.customerLat != null) ...[
                    const Icon(Icons.location_on, size: 14, color: Color(0xFFDC2626)),
                    const SizedBox(width: 3),
                    const Text('Customer', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold)),
                  ],
                  if (widget.nearbyVendors != null && widget.nearbyVendors!.isNotEmpty) ...[
                    const Icon(Icons.storefront, size: 14, color: Color(0xFF0284C7)),
                    const SizedBox(width: 3),
                    Text('${widget.nearbyVendors!.length} Vendors', style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold)),
                  ],
                ],
              ),
            ),
          ),

          // OpenStreetMap External Launch Action
          Positioned(
            top: 8,
            right: 8,
            child: Material(
              color: Colors.white.withOpacity(0.92),
              borderRadius: BorderRadius.circular(6),
              elevation: 2,
              child: InkWell(
                borderRadius: BorderRadius.circular(6),
                onTap: _openExternalOSM,
                child: const Padding(
                  padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.open_in_new, size: 12, color: Color(0xFF2563EB)),
                      SizedBox(width: 4),
                      Text('OSM Directions', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Color(0xFF2563EB))),
                    ],
                  ),
                ),
              ),
            ),
          ),

          // Route Stats Overlay (if route available or fallback straight-line estimate)
          if (widget.route != null)
            Positioned(
              bottom: 24,
              left: 8,
              right: 8,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                decoration: BoxDecoration(
                  color: widget.route!.routingAvailable
                      ? const Color(0xFF1E293B).withOpacity(0.90)
                      : const Color(0xFF78350F).withOpacity(0.92),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        Icon(
                          widget.route!.routingAvailable ? Icons.directions_bike : Icons.info_outline,
                          size: 16,
                          color: widget.route!.routingAvailable ? const Color(0xFF60A5FA) : const Color(0xFFFDE68A),
                        ),
                        const SizedBox(width: 6),
                        Text(
                          widget.route!.routingAvailable
                              ? '${widget.route!.distanceKm.toStringAsFixed(1)} km  •  ${widget.route!.durationMinutes.toStringAsFixed(0)} mins ETA'
                              : (widget.route!.approximateStraightLineKm != null
                                  ? 'Approx. ${widget.route!.approximateStraightLineKm!.toStringAsFixed(1)} km (Straight-line estimate)'
                                  : 'Road routing unavailable'),
                          style: const TextStyle(color: Colors.white, fontSize: 11, fontWeight: FontWeight.w600),
                        ),
                      ],
                    ),
                    Text(
                      widget.route!.routingAvailable ? 'OSRM Road Route' : 'ETA Unavailable',
                      style: TextStyle(
                        color: widget.route!.routingAvailable ? const Color(0xFF93C5FD) : const Color(0xFFFDE68A),
                        fontSize: 10,
                      ),
                    ),
                  ],
                ),
              ),
            ),

          // Mandatory OpenStreetMap Attribution banner
          Positioned(
            bottom: 2,
            right: 6,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 2),
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.85),
                borderRadius: BorderRadius.circular(4),
              ),
              child: const Text(
                '© OpenStreetMap contributors',
                style: TextStyle(fontSize: 9, color: Color(0xFF475569), fontWeight: FontWeight.w500),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFallbackView() {
    return Container(
      width: double.infinity,
      height: double.infinity,
      color: const Color(0xFFF8FAFC),
      padding: const EdgeInsets.all(16),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.map_outlined, size: 36, color: Color(0xFF64748B)),
          const SizedBox(height: 6),
          const Text(
            'OpenStreetMap Location Overview',
            style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Color(0xFF1E293B)),
          ),
          const SizedBox(height: 4),
          Text(
            'Lat: ${_centerLat.toStringAsFixed(4)}, Lng: ${_centerLng.toStringAsFixed(4)}',
            style: const TextStyle(fontSize: 11, color: Color(0xFF64748B)),
          ),
          const SizedBox(height: 10),
          ElevatedButton.icon(
            onPressed: _openExternalOSM,
            icon: const Icon(Icons.navigation, size: 14),
            label: const Text('Open in OpenStreetMap', style: TextStyle(fontSize: 11)),
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF2563EB),
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            ),
          ),
        ],
      ),
    );
  }
}

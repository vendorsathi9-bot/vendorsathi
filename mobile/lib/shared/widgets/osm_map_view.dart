import 'dart:convert';
import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:maplibre_gl/maplibre_gl.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../core/localization/locale_provider.dart';
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
  final String? selectedVendorId;

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
    this.selectedVendorId,
  });

  @override
  State<OsmMapView> createState() => OsmMapViewState();
}

class OsmMapViewState extends State<OsmMapView> with AutomaticKeepAliveClientMixin {
  MapLibreMapController? _mapController;
  bool _renderError = false;
  bool _styleLoaded = false;
  final Map<String, NearbyVendorModel> _circleToVendorMap = {};

  @override
  bool get wantKeepAlive => true;

  Future<void> animateToCoordinates(double lat, double lng, {double zoom = 14.5}) async {
    if (_mapController != null) {
      await _mapController!.animateCamera(
        CameraUpdate.newLatLngZoom(LatLng(lat, lng), zoom),
      );
    }
  }

  // OpenStreetMap raster tile style spec with configurable tile URL
  // Uses jsonEncode to ensure no leading whitespace/newlines exist, which allows
  // maplibre_gl_web's style sanitizer to correctly identify and parse the style object.
  String get _osmRasterStyle {
    final effectiveTileUrl = widget.tileUrl ?? 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';
    return jsonEncode({
      "version": 8,
      "sources": {
        "osm-raster-tiles": {
          "type": "raster",
          "tiles": [effectiveTileUrl],
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
    });
  }

  double get _centerLat {
    if (widget.agentLat != null && widget.agentLat != 0) return widget.agentLat!;
    if (widget.vendorLat != null && widget.vendorLat != 0) return widget.vendorLat!;
    if (widget.customerLat != null && widget.customerLat != 0) return widget.customerLat!;
    if (widget.nearbyVendors != null && widget.nearbyVendors!.isNotEmpty) {
      final valid = widget.nearbyVendors!.where((v) => v.latitude != 0 && v.longitude != 0).toList();
      if (valid.isNotEmpty) {
        return valid.map((v) => v.latitude).reduce((a, b) => a + b) / valid.length;
      }
    }
    return widget.initialLat ?? 12.9767; // Bangalore Majestic default
  }

  double get _centerLng {
    if (widget.agentLng != null && widget.agentLng != 0) return widget.agentLng!;
    if (widget.vendorLng != null && widget.vendorLng != 0) return widget.vendorLng!;
    if (widget.customerLng != null && widget.customerLng != 0) return widget.customerLng!;
    if (widget.nearbyVendors != null && widget.nearbyVendors!.isNotEmpty) {
      final valid = widget.nearbyVendors!.where((v) => v.latitude != 0 && v.longitude != 0).toList();
      if (valid.isNotEmpty) {
        return valid.map((v) => v.longitude).reduce((a, b) => a + b) / valid.length;
      }
    }
    return widget.initialLng ?? 77.5713;
  }

  void _onMapCreated(MapLibreMapController controller) {
    _mapController = controller;
    _mapController!.onCircleTapped.add((Circle circle) {
      final v = _circleToVendorMap[circle.id];
      if (v != null && widget.onVendorSelected != null) {
        widget.onVendorSelected!(v);
      }
    });
  }

  Future<void> _onStyleLoaded() async {
    if (!mounted || _mapController == null) return;
    _styleLoaded = true;
    await _renderMapFeatures();
  }

  @override
  void didUpdateWidget(OsmMapView oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (_mapController != null && _styleLoaded) {
      _renderMapFeatures();
    }
  }

  Future<void> _renderMapFeatures() async {
    if (!mounted || _mapController == null) return;

    try {
      await _mapController!.clearCircles();
      await _mapController!.clearLines();
      _circleToVendorMap.clear();

      final List<LatLng> boundsCoords = [];

      // 1. Draw OSRM route line if route coordinates exist
      if (widget.route != null && widget.route!.coordinates.isNotEmpty) {
        final lineCoords = widget.route!.coordinates
            .map((c) => LatLng(c[1], c[0])) // c[0]=lng, c[1]=lat -> LatLng(lat, lng)
            .toList();

        boundsCoords.addAll(lineCoords);

        await _mapController!.addLine(LineOptions(
          geometry: lineCoords,
          lineColor: '#1d4ed8',
          lineWidth: 5.0,
          lineOpacity: 0.9,
        ));
      }

      // 2. Add Vendor marker
      if (widget.vendorLat != null && widget.vendorLat != 0 && widget.vendorLng != null && widget.vendorLng != 0) {
        final p = LatLng(widget.vendorLat!, widget.vendorLng!);
        boundsCoords.add(p);
        await _mapController!.addCircle(CircleOptions(
          geometry: p,
          circleRadius: 9.0,
          circleColor: '#2563eb',
          circleStrokeWidth: 2.0,
          circleStrokeColor: '#ffffff',
        ));
      }

      // 3. Add Customer marker
      if (widget.customerLat != null && widget.customerLat != 0 && widget.customerLng != null && widget.customerLng != 0) {
        final p = LatLng(widget.customerLat!, widget.customerLng!);
        boundsCoords.add(p);
        await _mapController!.addCircle(CircleOptions(
          geometry: p,
          circleRadius: 9.0,
          circleColor: '#dc2626',
          circleStrokeWidth: 2.0,
          circleStrokeColor: '#ffffff',
        ));
      }

      // 4. Add Live Delivery Agent marker
      if (widget.agentLat != null && widget.agentLat != 0 && widget.agentLng != null && widget.agentLng != 0) {
        final p = LatLng(widget.agentLat!, widget.agentLng!);
        boundsCoords.add(p);
        await _mapController!.addCircle(CircleOptions(
          geometry: p,
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
            final p = LatLng(v.latitude, v.longitude);
            boundsCoords.add(p);
            final isSelected = widget.selectedVendorId != null &&
                v.vendorId == widget.selectedVendorId;
            final circle = await _mapController!.addCircle(CircleOptions(
              geometry: p,
              circleRadius: isSelected ? 12.0 : (v.isOpen ? 8.5 : 7.0),
              circleColor: isSelected
                  ? '#16a34a'
                  : (v.isOpen ? '#0284c7' : '#64748b'),
              circleStrokeWidth: isSelected ? 3.5 : 2.0,
              circleStrokeColor: isSelected ? '#f59e0b' : '#ffffff',
            ));
            _circleToVendorMap[circle.id] = v;
          }
        }
      }

      // Fit map bounds to show available vendors and positions
      if (boundsCoords.length >= 2) {
        double minLat = boundsCoords.first.latitude;
        double maxLat = boundsCoords.first.latitude;
        double minLng = boundsCoords.first.longitude;
        double maxLng = boundsCoords.first.longitude;

        for (final pt in boundsCoords) {
          if (pt.latitude < minLat) minLat = pt.latitude;
          if (pt.latitude > maxLat) maxLat = pt.latitude;
          if (pt.longitude < minLng) minLng = pt.longitude;
          if (pt.longitude > maxLng) maxLng = pt.longitude;
        }

        if ((maxLat - minLat).abs() > 0.001 || (maxLng - minLng).abs() > 0.001) {
          final bounds = LatLngBounds(
            southwest: LatLng(minLat, minLng),
            northeast: LatLng(maxLat, maxLng),
          );
          await _mapController!.animateCamera(
            CameraUpdate.newLatLngBounds(
              bounds,
              left: 40,
              top: 40,
              right: 40,
              bottom: 40,
            ),
          );
        } else {
          await _mapController!.animateCamera(
            CameraUpdate.newLatLngZoom(LatLng(minLat, minLng), widget.initialZoom),
          );
        }
      } else if (boundsCoords.length == 1) {
        await _mapController!.animateCamera(
          CameraUpdate.newLatLngZoom(boundsCoords.first, widget.initialZoom),
        );
      }
    } catch (e) {
      debugPrint('OsmMapView render error: $e');
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
    super.build(context);
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
                    annotationOrder: const [AnnotationType.circle, AnnotationType.line],
                    attributionButtonMargins: const math.Point(-100, -100),
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
    LocaleProvider? loc;
    try {
      loc = LocaleProvider.of(context);
    } catch (_) {}

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
          Text(
            loc?.t('map_load_error') ?? 'OpenStreetMap Location Overview',
            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Color(0xFF1E293B)),
          ),
          const SizedBox(height: 4),
          Text(
            loc?.t('check_connection') ?? 'Lat: ${_centerLat.toStringAsFixed(4)}, Lng: ${_centerLng.toStringAsFixed(4)}',
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 11, color: Color(0xFF64748B)),
          ),
          const SizedBox(height: 10),
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              OutlinedButton.icon(
                onPressed: () {
                  setState(() {
                    _renderError = false;
                  });
                },
                icon: const Icon(Icons.refresh_rounded, size: 14),
                label: Text(loc?.t('refresh') ?? 'Retry', style: const TextStyle(fontSize: 11)),
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                ),
              ),
              const SizedBox(width: 8),
              ElevatedButton.icon(
                onPressed: _openExternalOSM,
                icon: const Icon(Icons.navigation, size: 14),
                label: Text(loc?.t('open_in_osm') ?? 'Open in OpenStreetMap', style: const TextStyle(fontSize: 11)),
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF2563EB),
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

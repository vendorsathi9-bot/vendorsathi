import 'dart:async';
import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:maplibre_gl/maplibre_gl.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../core/api/api_client.dart';
import '../../core/localization/locale_provider.dart';
import '../../core/map/map_style_config.dart';
import '../models/models.dart';

/// Embedded map powered by MapLibre GL.
///
/// The style/provider comes from [MapStyleConfigService] (backed by the
/// backend's `/mobility/config` endpoint). The public OpenStreetMap tile
/// servers are intentionally not used: their usage policy does not permit
/// mobile applications, which is what produced the HTTP 403 error tiles.
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
  final String? selectedVendorId;
  final ApiClient? apiClient;

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
    this.selectedVendorId,
    this.apiClient,
  });

  @override
  State<OsmMapView> createState() => OsmMapViewState();
}

/// Distinct error taxonomy for map tile and provider failures.
enum MapErrorCategory {
  forbidden403,
  rateLimited429,
  timeout,
  noInternet,
  invalidStyleOrSource,
  missingOrInvalidConfig,
  unknown,
}

class OsmMapViewState extends State<OsmMapView>
    with AutomaticKeepAliveClientMixin {
  static const Duration _styleLoadTimeout = Duration(seconds: 15);
  static const Duration _retryCooldown = Duration(seconds: 3);

  MapLibreMapController? _mapController;
  MapStyleConfig? _styleConfig;
  bool _renderError = false;
  bool _styleLoaded = false;
  bool _showLoading = true;
  String? _errorMessage;
  MapErrorCategory? _errorCategory;
  DateTime? _lastRetryTime;
  int _mapAttempt = 0;
  Timer? _loadTimeout;
  final Map<String, NearbyVendorModel> _circleToVendorMap = {};

  @override
  bool get wantKeepAlive => true;

  @override
  void initState() {
    super.initState();
    _loadStyleConfig();
  }

  @override
  void dispose() {
    _loadTimeout?.cancel();
    if (_mapController != null) {
      _mapController!.onCircleTapped.remove(_onCircleTap);
      _mapController = null;
    }
    super.dispose();
  }

  Future<void> _loadStyleConfig({bool forceRefresh = false}) async {
    try {
      final config = await MapStyleConfigService.instance.resolve(
        widget.apiClient,
        forceRefresh: forceRefresh,
      );

      if (!mounted) return;

      if (!config.isValid) {
        setState(() {
          _showLoading = false;
          _renderError = true;
          _errorCategory = MapErrorCategory.missingOrInvalidConfig;
          _errorMessage = 'Missing or invalid map provider configuration.';
        });
        debugPrint('[VendorSaathi Map] Config error: style=${config.styleUrl}, tile=${config.tileUrl}');
        return;
      }

      setState(() {
        _styleConfig = config;
        _styleLoaded = false;
        _showLoading = true;
        _renderError = false;
        _errorCategory = null;
        _errorMessage = null;
        _mapAttempt++;
      });
      _armStyleLoadTimeout();
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _showLoading = false;
        _renderError = true;
        _errorCategory = MapErrorCategory.noInternet;
        _errorMessage = 'Could not load map configuration: $e';
      });
      debugPrint('[VendorSaathi Map] Exception loading config: $e');
    }
  }

  void _retry() {
    final now = DateTime.now();
    if (_lastRetryTime != null && now.difference(_lastRetryTime!) < _retryCooldown) {
      return; // Prevent uncontrolled request loop
    }
    _lastRetryTime = now;
    _loadTimeout?.cancel();
    setState(() {
      _renderError = false;
      _showLoading = true;
      _styleLoaded = false;
      _styleConfig = null;
      _errorCategory = null;
      _errorMessage = null;
    });
    _loadStyleConfig(forceRefresh: true);
  }

  void _armStyleLoadTimeout() {
    _loadTimeout?.cancel();
    _loadTimeout = Timer(_styleLoadTimeout, () {
      if (!mounted || _styleLoaded) return;
      setState(() {
        _showLoading = false;
        _renderError = true;
        _errorCategory = MapErrorCategory.timeout;
        _errorMessage = 'The map provider did not respond in time.';
      });
      debugPrint('[VendorSaathi Map] Timeout loading style: ${_styleConfig?.effectiveStyleString}');
    });
  }

  Future<void> animateToCoordinates(double lat, double lng,
      {double zoom = 14.5}) async {
    if (_mapController != null && _styleLoaded) {
      await _mapController!.animateCamera(
        CameraUpdate.newLatLngZoom(LatLng(lat, lng), zoom),
      );
    }
  }

  double get _centerLat {
    // 1. initialLat has highest priority (screen-specific coordinates)
    if (widget.initialLat != null && widget.initialLat != 0) {
      return widget.initialLat!;
    }
    // 2. customer coordinates
    if (widget.customerLat != null && widget.customerLat != 0) {
      return widget.customerLat!;
    }
    // 3. vendor coordinates
    if (widget.vendorLat != null && widget.vendorLat != 0) {
      return widget.vendorLat!;
    }
    // 4. agent coordinates
    if (widget.agentLat != null && widget.agentLat != 0) {
      return widget.agentLat!;
    }
    // 5. nearby vendors average
    if (widget.nearbyVendors != null && widget.nearbyVendors!.isNotEmpty) {
      final valid = widget.nearbyVendors!
          .where((v) => v.latitude != 0 && v.longitude != 0)
          .toList();
      if (valid.isNotEmpty) {
        final avgLat =
            valid.map((v) => v.latitude).reduce((a, b) => a + b) / valid.length;
        return avgLat;
      }
    }
    // 6. default Bangalore Majestic
    return 12.9767;
  }

  double get _centerLng {
    // 1. initialLng has highest priority (screen-specific coordinates)
    if (widget.initialLng != null && widget.initialLng != 0) {
      return widget.initialLng!;
    }
    // 2. customer coordinates
    if (widget.customerLng != null && widget.customerLng != 0) {
      return widget.customerLng!;
    }
    // 3. vendor coordinates
    if (widget.vendorLng != null && widget.vendorLng != 0) {
      return widget.vendorLng!;
    }
    // 4. agent coordinates
    if (widget.agentLng != null && widget.agentLng != 0) {
      return widget.agentLng!;
    }
    // 5. nearby vendors average
    if (widget.nearbyVendors != null && widget.nearbyVendors!.isNotEmpty) {
      final valid = widget.nearbyVendors!
          .where((v) => v.latitude != 0 && v.longitude != 0)
          .toList();
      if (valid.isNotEmpty) {
        final avgLng = valid.map((v) => v.longitude).reduce((a, b) => a + b) /
            valid.length;
        return avgLng;
      }
    }
    // 6. default Bangalore Majestic
    return 77.5713;
  }

  void _onCircleTap(Circle circle) {
    final v = _circleToVendorMap[circle.id];
    if (v != null && widget.onVendorSelected != null) {
      widget.onVendorSelected!(v);
    }
  }

  void _onMapCreated(MapLibreMapController controller) {
    _mapController = controller;
    _mapController!.onCircleTapped.add(_onCircleTap);
    if (!_styleLoaded) {
      _armStyleLoadTimeout();
    }
  }

  Future<void> _onStyleLoaded() async {
    if (!mounted || _mapController == null) return;
    _loadTimeout?.cancel();
    setState(() {
      _styleLoaded = true;
      _showLoading = false;
      _renderError = false;
      _errorMessage = null;
    });
    await _renderMapFeatures();
  }

  @override
  void didUpdateWidget(OsmMapView oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (_mapController != null && _styleLoaded) {
      if (oldWidget.vendorLat != widget.vendorLat ||
          oldWidget.vendorLng != widget.vendorLng ||
          oldWidget.customerLat != widget.customerLat ||
          oldWidget.customerLng != widget.customerLng ||
          oldWidget.agentLat != widget.agentLat ||
          oldWidget.agentLng != widget.agentLng ||
          oldWidget.selectedVendorId != widget.selectedVendorId ||
          oldWidget.route != widget.route ||
          oldWidget.nearbyVendors?.length != widget.nearbyVendors?.length) {
        _renderMapFeatures();
      }
    }
  }

  Future<void> _renderMapFeatures() async {
    if (!mounted || _mapController == null || !_styleLoaded) return;

    try {
      await _mapController!.clearCircles();
      await _mapController!.clearLines();
      _circleToVendorMap.clear();

      final List<LatLng> boundsCoords = [];

      // 1. Draw OSRM route line if route coordinates exist
      if (widget.route != null && widget.route!.coordinates.isNotEmpty) {
        final lineCoords = widget.route!.coordinates
            .map((c) =>
                LatLng(c[1], c[0])) // c[0]=lng, c[1]=lat -> LatLng(lat, lng)
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
      if (widget.vendorLat != null &&
          widget.vendorLat != 0 &&
          widget.vendorLng != null &&
          widget.vendorLng != 0) {
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
      if (widget.customerLat != null &&
          widget.customerLat != 0 &&
          widget.customerLng != null &&
          widget.customerLng != 0) {
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
      if (widget.agentLat != null &&
          widget.agentLat != 0 &&
          widget.agentLng != null &&
          widget.agentLng != 0) {
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
              circleColor:
                  isSelected ? '#16a34a' : (v.isOpen ? '#0284c7' : '#64748b'),
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

        if ((maxLat - minLat).abs() > 0.001 ||
            (maxLng - minLng).abs() > 0.001) {
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
            CameraUpdate.newLatLngZoom(
                LatLng(minLat, minLng), widget.initialZoom),
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

    final url =
        'https://www.openstreetmap.org/directions?engine=fossgis_osrm_car&route=$fromLat%2C$fromLng%3B$toLat%2C$toLng';
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
        color: const Color(0xFFFFFCEF),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFFEF4CD)),
      ),
      clipBehavior: Clip.antiAlias,
      child: Stack(
        children: [
          // Map view
          if (_styleConfig != null && !_renderError)
            Builder(
              builder: (ctx) {
                try {
                  return MapLibreMap(
                    key: ValueKey('osm-map-$_mapAttempt'),
                    styleString: _styleConfig!.effectiveStyleString,
                    initialCameraPosition: CameraPosition(
                      target: LatLng(_centerLat, _centerLng),
                      zoom: widget.initialZoom,
                    ),
                    onMapCreated: _onMapCreated,
                    onStyleLoadedCallback: _onStyleLoaded,
                    myLocationEnabled: false,
                    trackCameraPosition: true,
                    annotationOrder: const [
                      AnnotationType.circle,
                      AnnotationType.line
                    ],
                    attributionButtonMargins: const math.Point(-100, -100),
                  );
                } catch (e) {
                  WidgetsBinding.instance.addPostFrameCallback((_) {
                    if (mounted) {
                      setState(() {
                        _showLoading = false;
                        _renderError = true;
                        _errorCategory = MapErrorCategory.invalidStyleOrSource;
                        _errorMessage = 'Invalid map style or source: $e';
                      });
                      debugPrint('[VendorSaathi Map] MapLibre renderer exception: $e');
                    }
                  });
                  return const SizedBox.shrink();
                }
              },
            ),

          // Explicit error state (never show raw provider error tiles)
          if (_renderError) _buildErrorView(),

          // Loading state, shown until the style has actually loaded
          if (!_renderError && (_showLoading || !_styleLoaded))
            _buildLoadingView(),

          // Legend Overlay
          Positioned(
            top: 8,
            left: 8,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.92),
                borderRadius: BorderRadius.circular(6),
                boxShadow: const [
                  BoxShadow(color: Colors.black12, blurRadius: 4)
                ],
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (widget.vendorLat != null) ...[
                    const Icon(Icons.storefront,
                        size: 14, color: Color(0xFF2563EB)),
                    const SizedBox(width: 3),
                    const Text('Vendor',
                        style: TextStyle(
                            fontSize: 10, fontWeight: FontWeight.bold)),
                    const SizedBox(width: 8),
                  ],
                  if (widget.agentLat != null) ...[
                    const Icon(Icons.two_wheeler,
                        size: 14, color: Color(0xFF16A34A)),
                    const SizedBox(width: 3),
                    const Text('Agent',
                        style: TextStyle(
                            fontSize: 10,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF16A34A))),
                    const SizedBox(width: 8),
                  ],
                  if (widget.customerLat != null) ...[
                    const Icon(Icons.location_on,
                        size: 14, color: Color(0xFFDC2626)),
                    const SizedBox(width: 3),
                    const Text('Customer',
                        style: TextStyle(
                            fontSize: 10, fontWeight: FontWeight.bold)),
                  ],
                  if (widget.nearbyVendors != null &&
                      widget.nearbyVendors!.isNotEmpty) ...[
                    const Icon(Icons.storefront,
                        size: 14, color: Color(0xFF0284C7)),
                    const SizedBox(width: 3),
                    Text('${widget.nearbyVendors!.length} Vendors',
                        style: const TextStyle(
                            fontSize: 10, fontWeight: FontWeight.bold)),
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
                      Icon(Icons.open_in_new,
                          size: 12, color: Color(0xFF2563EB)),
                      SizedBox(width: 4),
                      Text('OSM Directions',
                          style: TextStyle(
                              fontSize: 10,
                              fontWeight: FontWeight.bold,
                              color: Color(0xFF2563EB))),
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
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
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
                          widget.route!.routingAvailable
                              ? Icons.directions_bike
                              : Icons.info_outline,
                          size: 16,
                          color: widget.route!.routingAvailable
                              ? const Color(0xFF60A5FA)
                              : const Color(0xFFFDE68A),
                        ),
                        const SizedBox(width: 6),
                        Text(
                          widget.route!.routingAvailable
                              ? '${widget.route!.distanceKm.toStringAsFixed(1)} km  •  ${widget.route!.durationMinutes.toStringAsFixed(0)} mins ETA'
                              : (widget.route!.approximateStraightLineKm != null
                                  ? 'Approx. ${widget.route!.approximateStraightLineKm!.toStringAsFixed(1)} km (Straight-line estimate)'
                                  : 'Road routing unavailable'),
                          style: const TextStyle(
                              color: Colors.white,
                              fontSize: 11,
                              fontWeight: FontWeight.w600),
                        ),
                      ],
                    ),
                    Text(
                      widget.route!.routingAvailable
                          ? 'OSRM Road Route'
                          : 'ETA Unavailable',
                      style: TextStyle(
                        color: widget.route!.routingAvailable
                            ? const Color(0xFF93C5FD)
                            : const Color(0xFFFDE68A),
                        fontSize: 10,
                      ),
                    ),
                  ],
                ),
              ),
            ),

          // Provider attribution (required by the map data licences)
          Positioned(
            bottom: 2,
            right: 6,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 2),
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.85),
                borderRadius: BorderRadius.circular(4),
              ),
              child: Text(
                _styleConfig?.attribution ??
                    MapStyleConfig.defaults.attribution,
                style: const TextStyle(
                    fontSize: 9,
                    color: Color(0xFF475569),
                    fontWeight: FontWeight.w500),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLoadingView() {
    LocaleProvider? loc;
    try {
      loc = LocaleProvider.of(context);
    } catch (_) {}

    return Positioned.fill(
      child: ColoredBox(
        color: const Color(0xFFFFFCEF),
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const SizedBox(
                width: 26,
                height: 26,
                child: CircularProgressIndicator(
                  strokeWidth: 2.6,
                  valueColor: AlwaysStoppedAnimation(Color(0xFFF4B400)),
                ),
              ),
              const SizedBox(height: 10),
              Text(
                loc?.t('map_loading') ?? 'Loading map...',
                style: const TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF38200F),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildErrorView() {
    LocaleProvider? loc;
    try {
      loc = LocaleProvider.of(context);
    } catch (_) {}

    IconData errorIcon = Icons.map_outlined;
    String title = loc?.t('map_load_error') ?? 'Unable to load the map';
    String body = loc?.t('map_unavailable_body') ??
        'The map is temporarily unavailable. Check your connection and try again.';

    switch (_errorCategory) {
      case MapErrorCategory.forbidden403:
        errorIcon = Icons.lock_outline;
        title = 'Map Access Denied (HTTP 403)';
        body = 'The tile provider rejected access. Tile usage policy restrictions apply.';
        break;
      case MapErrorCategory.rateLimited429:
        errorIcon = Icons.speed;
        title = 'Rate Limit Exceeded (HTTP 429)';
        body = 'Too many requests sent to the map provider. Please wait a moment.';
        break;
      case MapErrorCategory.timeout:
        errorIcon = Icons.timer_outlined;
        title = 'Network Timeout';
        body = 'The map provider did not respond in time. Please check your network.';
        break;
      case MapErrorCategory.noInternet:
        errorIcon = Icons.wifi_off_rounded;
        title = 'No Internet Connection';
        body = 'Unable to connect to map services. Please check your data connection.';
        break;
      case MapErrorCategory.invalidStyleOrSource:
        errorIcon = Icons.broken_image_outlined;
        title = 'Invalid Map Style';
        body = 'The map style or source configuration could not be parsed.';
        break;
      case MapErrorCategory.missingOrInvalidConfig:
        errorIcon = Icons.settings_suggest_outlined;
        title = 'Configuration Error';
        body = 'Missing or invalid map provider configuration.';
        break;
      case MapErrorCategory.unknown:
      case null:
        break;
    }

    return Positioned.fill(
      child: ColoredBox(
        color: const Color(0xFFFFFCEF),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(errorIcon, size: 36, color: const Color(0xFFF4B400)),
              const SizedBox(height: 8),
              Text(
                title,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontWeight: FontWeight.bold,
                  fontSize: 13,
                  color: Color(0xFF38200F),
                ),
              ),
              const SizedBox(height: 4),
              Text(
                body,
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 11, color: Color(0xFF6B4F3A)),
              ),
              if (_errorMessage != null) ...[
                const SizedBox(height: 3),
                Text(
                  _errorMessage!,
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 9, color: Color(0xFF9A7B62)),
                ),
              ],
              const SizedBox(height: 10),
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  OutlinedButton.icon(
                    onPressed: _retry,
                    icon: const Icon(Icons.refresh_rounded, size: 14),
                    label: Text(loc?.t('refresh') ?? 'Retry',
                        style: const TextStyle(fontSize: 11)),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: const Color(0xFF38200F),
                      side: const BorderSide(color: Color(0xFFF4B400)),
                      padding: const EdgeInsets.symmetric(
                          horizontal: 10, vertical: 6),
                    ),
                  ),
                  const SizedBox(width: 8),
                  ElevatedButton.icon(
                    onPressed: _openExternalOSM,
                    icon: const Icon(Icons.navigation, size: 14),
                    label: Text(
                        loc?.t('open_in_osm') ?? 'Open in OpenStreetMap',
                        style: const TextStyle(fontSize: 11)),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFFF4B400),
                      foregroundColor: const Color(0xFF38200F),
                      padding: const EdgeInsets.symmetric(
                          horizontal: 12, vertical: 6),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

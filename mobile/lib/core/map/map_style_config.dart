import 'dart:convert';
import 'package:flutter/foundation.dart';

import '../api/api_client.dart';

/// Map style/provider configuration shared by every MapLibre surface.
///
/// The backend is the source of truth (`GET /api/v1/mobility/config`) so the
/// provider can be changed with environment variables (MAP_STYLE_URL,
/// MAP_ATTRIBUTION) without shipping a new app build. The compile-time
/// defaults below keep the map usable when the config endpoint is temporarily
/// unreachable.
///
/// The default is OpenFreeMap: a keyless, MapLibre-compatible vector tile
/// provider that explicitly permits commercial website and app usage. The
/// public OpenStreetMap tile servers are intentionally not used because their
/// usage policy forbids mobile applications.
class MapStyleConfig {
  final String provider;
  final String styleUrl;
  final String attribution;
  final String tileUrl;

  const MapStyleConfig({
    required this.provider,
    required this.styleUrl,
    required this.attribution,
    this.tileUrl = '',
  });

  static const MapStyleConfig defaults = MapStyleConfig(
    provider: String.fromEnvironment(
      'MAP_PROVIDER',
      defaultValue: 'openfreemap',
    ),
    styleUrl: String.fromEnvironment(
      'MAP_STYLE_URL',
      defaultValue: 'https://tiles.openfreemap.org/styles/liberty',
    ),
    attribution: String.fromEnvironment(
      'MAP_ATTRIBUTION',
      defaultValue: '© OpenFreeMap © OpenMapTiles © OpenStreetMap contributors',
    ),
  );

  factory MapStyleConfig.fromApi(Map<String, dynamic> json) {
    final styleUrl = (json['style_url'] as String?)?.trim() ?? '';
    final attribution = (json['attribution'] as String?)?.trim() ?? '';
    final provider = (json['provider'] as String?)?.trim() ?? '';

    return MapStyleConfig(
      provider: provider.isNotEmpty ? provider : defaults.provider,
      styleUrl: _isHttpUrl(styleUrl) ? styleUrl : defaults.styleUrl,
      attribution: attribution.isNotEmpty ? attribution : defaults.attribution,
      tileUrl: (json['tile_url'] as String?)?.trim() ?? '',
    );
  }

  bool get isValid => _isHttpUrl(styleUrl) || _isHttpUrl(tileUrl);

  /// Returns the MapLibre style string.
  /// If a raster [tileUrl] is configured, generates a valid MapLibre raster style JSON.
  /// Otherwise, returns [styleUrl] (vector style URL).
  String get effectiveStyleString {
    if (tileUrl.isNotEmpty &&
        tileUrl.contains('{z}') &&
        tileUrl.contains('{x}') &&
        tileUrl.contains('{y}')) {
      return jsonEncode({
        "version": 8,
        "sources": {
          "raster-tiles": {
            "type": "raster",
            "tiles": [tileUrl],
            "tileSize": 256,
            "attribution": attribution,
          }
        },
        "layers": [
          {
            "id": "raster-tiles-layer",
            "type": "raster",
            "source": "raster-tiles",
            "minzoom": 0,
            "maxzoom": 19,
          }
        ]
      });
    }
    return styleUrl.isNotEmpty ? styleUrl : defaults.styleUrl;
  }

  static bool _isHttpUrl(String value) =>
      value.startsWith('https://') || value.startsWith('http://');
}

/// Loads [MapStyleConfig] once per app session and shares it with all maps.
class MapStyleConfigService {
  MapStyleConfigService._();

  static final MapStyleConfigService instance = MapStyleConfigService._();

  MapStyleConfig? _cached;
  Future<MapStyleConfig>? _inFlight;

  Future<MapStyleConfig> resolve(
    ApiClient? apiClient, {
    bool forceRefresh = false,
  }) {
    if (!forceRefresh && _cached != null) {
      return Future.value(_cached);
    }
    if (!forceRefresh && _inFlight != null) {
      return _inFlight!;
    }

    final future = _load(apiClient).then((config) {
      _cached = config;
      return config;
    }).catchError((Object error) {
      debugPrint(
        'MapStyleConfig: could not load /mobility/config, using built-in '
        'defaults: $error',
      );
      const fallback = MapStyleConfig.defaults;
      _cached = fallback;
      return fallback;
    }).whenComplete(() {
      _inFlight = null;
    });

    _inFlight = future;
    return future;
  }

  Future<MapStyleConfig> _load(ApiClient? apiClient) async {
    if (apiClient == null) {
      return MapStyleConfig.defaults;
    }

    final response = await apiClient.get('/mobility/config');
    if (response is! Map) {
      return MapStyleConfig.defaults;
    }

    // Accept both the plain object returned by the API and a defensively
    // wrapped { "data": {...} } shape.
    final data = response['data'];
    final json = data is Map
        ? Map<String, dynamic>.from(data)
        : Map<String, dynamic>.from(response);

    return MapStyleConfig.fromApi(json);
  }

  /// Clears the session cache (used by tests and explicit retries).
  @visibleForTesting
  void reset() {
    _cached = null;
    _inFlight = null;
  }
}

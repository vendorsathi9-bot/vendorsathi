import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:http/http.dart' as http;

class ApiException implements Exception {
  final String code;
  final String message;
  final int statusCode;

  ApiException({required this.code, required this.message, required this.statusCode});

  @override
  String toString() => 'ApiException($statusCode, $code): $message';
}

class ApiClient {
  static const _accessTokenKey = 'vendorsaathi.access_token';
  static const _refreshTokenKey = 'vendorsaathi.refresh_token';

  /// Override with:
  ///   flutter run --dart-define=API_BASE_URL=https://api.example.com/api/v1
  static const _envBaseUrl = String.fromEnvironment('API_BASE_URL');

  /// Deployed backend. Used for release builds when no dart-define is given.
  static const _productionBaseUrl = 'https://vendorsathi.onrender.com/api/v1';

  final String baseUrl;
  final FlutterSecureStorage _storage;
  String? _accessToken;
  String? _refreshToken;

  ApiClient({String? baseUrl, FlutterSecureStorage? storage})
      : baseUrl = baseUrl ?? _resolveBaseUrl(),
        _storage = storage ?? const FlutterSecureStorage();

  static String _resolveBaseUrl() {
    if (_envBaseUrl.isNotEmpty) {
      return _envBaseUrl.replaceFirst(RegExp(r'/+$'), '');
    }
    if (!kDebugMode) {
      return _productionBaseUrl;
    }
    if (!kIsWeb && defaultTargetPlatform == TargetPlatform.android) {
      // Android emulator reaches the host machine through 10.0.2.2.
      return 'http://10.0.2.2:8080/api/v1';
    }
    return 'http://localhost:8080/api/v1';
  }

  bool get hasSession => _accessToken != null || _refreshToken != null;

  /// Loads previously persisted tokens. Call once during app startup.
  Future<void> restoreSession() async {
    try {
      _accessToken = await _storage.read(key: _accessTokenKey);
      _refreshToken = await _storage.read(key: _refreshTokenKey);
    } catch (_) {
      // Secure storage may be unavailable (e.g. unsupported platform); the
      // user simply has to sign in again.
    }
  }

  void setToken(String? token) {
    _accessToken = token;
    if (token == null || token.isEmpty) {
      _refreshToken = null;
      _deleteToken(_accessTokenKey);
      _deleteToken(_refreshTokenKey);
    } else {
      _writeToken(_accessTokenKey, token);
    }
  }

  void setTokens(String? accessToken, String? refreshToken) {
    _accessToken = accessToken;
    _refreshToken = refreshToken;
    if (accessToken != null && accessToken.isNotEmpty) {
      _writeToken(_accessTokenKey, accessToken);
    } else {
      _deleteToken(_accessTokenKey);
    }
    if (refreshToken != null && refreshToken.isNotEmpty) {
      _writeToken(_refreshTokenKey, refreshToken);
    } else {
      _deleteToken(_refreshTokenKey);
    }
  }

  // Persistence is best-effort and must never crash a caller that has already
  // authenticated: a storage failure only means the session is not restored
  // after an app restart.
  void _writeToken(String key, String value) {
    _storage.write(key: key, value: value).catchError((_) {});
  }

  void _deleteToken(String key) {
    _storage.delete(key: key).catchError((_) {});
  }

  Future<void> clearSession() async {
    _accessToken = null;
    _refreshToken = null;
    try {
      await _storage.delete(key: _accessTokenKey);
      await _storage.delete(key: _refreshTokenKey);
    } catch (_) {
      // Ignore storage errors on sign-out.
    }
  }

  String? get token => _accessToken;

  Map<String, String> _headers() {
    final headers = {'Content-Type': 'application/json'};
    if (_accessToken != null && _accessToken!.isNotEmpty) {
      headers['Authorization'] = 'Bearer $_accessToken';
    }
    return headers;
  }

  Uri _uri(String path, {Map<String, String>? queryParams}) {
    var uri = Uri.parse('$baseUrl$path');
    if (queryParams != null && queryParams.isNotEmpty) {
      uri = uri.replace(queryParameters: queryParams);
    }
    return uri;
  }

  /// Exchanges the stored refresh token for a new pair. Returns false when the
  /// session cannot be refreshed; the caller must then sign the user out.
  Future<bool> _refreshAccessToken() async {
    final refresh = _refreshToken;
    if (refresh == null || refresh.isEmpty) return false;
    try {
      final res = await http.post(
        Uri.parse('$baseUrl/auth/refresh'),
        headers: const {'Content-Type': 'application/json'},
        body: jsonEncode({'refresh_token': refresh}),
      );
      if (res.statusCode < 200 || res.statusCode >= 300) {
        await clearSession();
        return false;
      }
      final body = jsonDecode(res.body);
      if (body is! Map<String, dynamic> || body['access_token'] is! String) {
        await clearSession();
        return false;
      }
      setTokens(
        body['access_token'] as String,
        body['refresh_token'] as String?,
      );
      return true;
    } catch (_) {
      return false;
    }
  }

  Future<http.Response> _send(
    String method,
    String path, {
    Map<String, String>? queryParams,
    dynamic body,
    bool allowRefresh = true,
  }) async {
    final uri = _uri(path, queryParams: queryParams);
    late http.Response response;

    switch (method) {
      case 'GET':
        response = await http.get(uri, headers: _headers());
        break;
      case 'POST':
        response = await http.post(uri, headers: _headers(), body: body != null ? jsonEncode(body) : null);
        break;
      case 'PATCH':
        response = await http.patch(uri, headers: _headers(), body: body != null ? jsonEncode(body) : null);
        break;
      case 'DELETE':
        response = await http.delete(uri, headers: _headers());
        break;
      default:
        throw ArgumentError('Unsupported HTTP method: $method');
    }

    // Refresh once and retry. Auth endpoints are excluded to avoid loops.
    if (response.statusCode == 401 &&
        allowRefresh &&
        !path.startsWith('/auth/login') &&
        !path.startsWith('/auth/register') &&
        !path.startsWith('/auth/refresh')) {
      if (await _refreshAccessToken()) {
        return _send(method, path, queryParams: queryParams, body: body, allowRefresh: false);
      }
    }

    return response;
  }

  Future<dynamic> get(String path, {Map<String, String>? queryParams}) async {
    return _handleResponse(await _send('GET', path, queryParams: queryParams));
  }

  Future<dynamic> post(String path, {dynamic body}) async {
    return _handleResponse(await _send('POST', path, body: body));
  }

  Future<dynamic> patch(String path, {dynamic body}) async {
    return _handleResponse(await _send('PATCH', path, body: body));
  }

  Future<dynamic> delete(String path) async {
    return _handleResponse(await _send('DELETE', path));
  }

  dynamic _handleResponse(http.Response res) {
    if (res.statusCode == 204) return null;

    dynamic jsonBody;
    try {
      jsonBody = jsonDecode(res.body);
    } catch (_) {
      jsonBody = null;
    }

    if (res.statusCode >= 200 && res.statusCode < 300) {
      return jsonBody;
    }

    String code = 'UNKNOWN_ERROR';
    String message = 'Request failed (${res.statusCode})';

    if (jsonBody is Map<String, dynamic> && jsonBody.containsKey('error')) {
      final errObj = jsonBody['error'];
      if (errObj is Map<String, dynamic>) {
        code = errObj['code'] ?? code;
        message = errObj['message'] ?? message;
      } else if (errObj is String) {
        message = errObj;
      }
    }

    throw ApiException(code: code, message: message, statusCode: res.statusCode);
  }

  static Map<String, dynamic>? _cachedMapConfig;

  /// Fetches the dynamic map tile configuration from the backend mobility service.
  /// Resolves relative tile URLs to the absolute API server address.
  Future<Map<String, dynamic>?> getMapConfig({bool forceRefresh = false}) async {
    if (!forceRefresh && _cachedMapConfig != null) {
      return _cachedMapConfig;
    }
    try {
      final res = await get('/mobility/config');
      if (res is Map<String, dynamic>) {
        final config = Map<String, dynamic>.from(res);
        final rawTileUrl = config['tile_url'] as String?;
        if (rawTileUrl != null && rawTileUrl.startsWith('/')) {
          final uri = Uri.parse(baseUrl);
          final origin = '${uri.scheme}://${uri.host}:${uri.port}';
          config['tile_url'] = '$origin$rawTileUrl';
        }
        _cachedMapConfig = config;
        return config;
      }
    } catch (_) {
      // Return cached config or fallback if network unavailable
    }
    return _cachedMapConfig;
  }
}

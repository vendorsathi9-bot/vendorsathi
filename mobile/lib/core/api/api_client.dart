import 'dart:convert';
import 'package:flutter/foundation.dart';
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
  final String baseUrl;
  String? _accessToken;

  ApiClient({String? baseUrl})
      : baseUrl = baseUrl ??
            (!kIsWeb && defaultTargetPlatform == TargetPlatform.android
                ? 'http://10.0.2.2:8080/api/v1'
                : 'http://localhost:8080/api/v1');

  void setToken(String? token) {
    _accessToken = token;
  }

  String? get token => _accessToken;

  Map<String, String> _headers() {
    final headers = {'Content-Type': 'application/json'};
    if (_accessToken != null && _accessToken!.isNotEmpty) {
      headers['Authorization'] = 'Bearer $_accessToken';
    }
    return headers;
  }

  Future<dynamic> get(String path, {Map<String, String>? queryParams}) async {
    var uri = Uri.parse('$baseUrl$path');
    if (queryParams != null && queryParams.isNotEmpty) {
      uri = uri.replace(queryParameters: queryParams);
    }
    final response = await http.get(uri, headers: _headers());
    return _handleResponse(response);
  }

  Future<dynamic> post(String path, {dynamic body}) async {
    final uri = Uri.parse('$baseUrl$path');
    final response = await http.post(
      uri,
      headers: _headers(),
      body: body != null ? jsonEncode(body) : null,
    );
    return _handleResponse(response);
  }

  Future<dynamic> patch(String path, {dynamic body}) async {
    final uri = Uri.parse('$baseUrl$path');
    final response = await http.patch(
      uri,
      headers: _headers(),
      body: body != null ? jsonEncode(body) : null,
    );
    return _handleResponse(response);
  }

  Future<dynamic> delete(String path) async {
    final uri = Uri.parse('$baseUrl$path');
    final response = await http.delete(uri, headers: _headers());
    return _handleResponse(response);
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

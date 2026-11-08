import 'package:flutter/material.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../api/api_client.dart';
import 'app_strings.dart';

class LocaleProvider extends InheritedWidget {
  final String currentLanguage;
  final Function(String langCode) setLanguage;
  final ApiClient? apiClient;

  const LocaleProvider({
    super.key,
    required this.currentLanguage,
    required this.setLanguage,
    this.apiClient,
    required super.child,
  });

  static LocaleProvider of(BuildContext context) {
    final provider = context.dependOnInheritedWidgetOfExactType<LocaleProvider>();
    assert(provider != null, 'No LocaleProvider found in context');
    return provider!;
  }

  String t(String key) => AppStrings.tr(key, currentLanguage);

  @override
  bool updateShouldNotify(LocaleProvider oldWidget) {
    return oldWidget.currentLanguage != currentLanguage;
  }
}

class LanguageController {
  static const _storage = FlutterSecureStorage();
  static const _langKey = 'preferred_language';

  static Future<String?> getSavedLanguage() async {
    try {
      return await _storage.read(key: _langKey);
    } catch (_) {
      return null;
    }
  }

  static Future<void> saveLanguage(String langCode, {ApiClient? apiClient}) async {
    try {
      await _storage.write(key: _langKey, value: langCode);
    } catch (_) {}

    // If authenticated, sync with user profile
    if (apiClient != null && apiClient.token != null) {
      try {
        await apiClient.patch('/users/me/profile', body: {
          'preferred_language': langCode,
        });
      } catch (_) {}
    }
  }
}

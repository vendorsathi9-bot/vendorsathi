import 'package:flutter/material.dart';
import 'app/theme.dart';
import 'core/api/api_client.dart';
import 'features/auth/login_screen.dart';
import 'features/customer_home/customer_home_screen.dart';
import 'features/vendor_home/vendor_home_screen.dart';
import 'shared/models/models.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const VendorSaathiApp());
}

class VendorSaathiApp extends StatefulWidget {
  const VendorSaathiApp({Key? key}) : super(key: key);

  @override
  State<VendorSaathiApp> createState() => _VendorSaathiAppState();
}

class _VendorSaathiAppState extends State<VendorSaathiApp> {
  final ApiClient _apiClient = ApiClient();
  UserModel? _currentUser;
  bool _preferVendorExperience = true;

  void _onLoginSuccess(UserModel user) {
    setState(() {
      _currentUser = user;
      _preferVendorExperience = user.isVendor;
    });
  }

  void _onLogout() {
    _apiClient.setToken(null);
    setState(() {
      _currentUser = null;
    });
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'VendorSaathi',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      home: _currentUser == null
          ? LoginScreen(
              apiClient: _apiClient,
              onLoginSuccess: _onLoginSuccess,
            )
          : (_preferVendorExperience && _currentUser!.isVendor)
              ? VendorHomeScreen(
                  apiClient: _apiClient,
                  onLogout: _onLogout,
                )
              : CustomerHomeScreen(
                  apiClient: _apiClient,
                  onLogout: _onLogout,
                ),
    );
  }
}

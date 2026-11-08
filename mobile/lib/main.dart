import 'package:flutter/material.dart';
import 'app/theme.dart';
import 'core/api/api_client.dart';
import 'core/localization/locale_provider.dart';
import 'features/auth/login_screen.dart';
import 'features/auth/register_screen.dart';
import 'features/customer_home/customer_home_screen.dart';
import 'features/delivery_agent/delivery_agent_home_screen.dart';
import 'features/language/language_selection_screen.dart';
import 'features/vendor_home/vendor_home_screen.dart';
import 'shared/models/models.dart';

enum AppExperience { vendor, customer, deliveryAgent }

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final savedLang = await LanguageController.getSavedLanguage();
  runApp(VendorSaathiApp(initialLanguage: savedLang));
}

class VendorSaathiApp extends StatefulWidget {
  final String? initialLanguage;

  const VendorSaathiApp({super.key, this.initialLanguage});

  @override
  State<VendorSaathiApp> createState() => _VendorSaathiAppState();
}

class _VendorSaathiAppState extends State<VendorSaathiApp> {
  final ApiClient _apiClient = ApiClient();
  late String _currentLanguage;
  bool _hasChosenLanguage = false;
  bool _showRegister = false;
  UserModel? _currentUser;
  AppExperience _currentExperience = AppExperience.customer;

  @override
  void initState() {
    super.initState();
    if (widget.initialLanguage != null && widget.initialLanguage!.isNotEmpty) {
      _currentLanguage = widget.initialLanguage!;
      _hasChosenLanguage = true;
    } else {
      _currentLanguage = 'en';
      _hasChosenLanguage = false;
    }
  }

  void _onLanguageSelected(String langCode) {
    setState(() {
      _currentLanguage = langCode;
      _hasChosenLanguage = true;
    });
  }

  void _setLanguage(String langCode) {
    setState(() {
      _currentLanguage = langCode;
    });
  }

  void _onLoginSuccess(UserModel user) {
    setState(() {
      _currentUser = user;
      if (user.isDeliveryAgent && !user.isVendor) {
        _currentExperience = AppExperience.deliveryAgent;
      } else if (user.isVendor) {
        _currentExperience = AppExperience.vendor;
      } else {
        _currentExperience = AppExperience.customer;
      }

      if (user.preferredLanguage.isNotEmpty) {
        _currentLanguage = user.preferredLanguage;
      }
    });
  }

  void _onLogout() {
    _apiClient.setToken(null);
    setState(() {
      _currentUser = null;
      _showRegister = false;
      _currentExperience = AppExperience.customer;
    });
  }

  void _switchToCustomerExperience() {
    setState(() {
      _currentExperience = AppExperience.customer;
    });
  }

  void _switchToVendorExperience() {
    setState(() {
      _currentExperience = AppExperience.vendor;
    });
  }

  void _switchToDeliveryExperience() {
    setState(() {
      _currentExperience = AppExperience.deliveryAgent;
    });
  }

  @override
  Widget build(BuildContext context) {
    return LocaleProvider(
      currentLanguage: _currentLanguage,
      setLanguage: _setLanguage,
      apiClient: _apiClient,
      child: MaterialApp(
        title: 'VendorSaathi',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.lightTheme,
        home: Builder(
          builder: (ctx) {
            // 1. First-launch Language Selection
            if (!_hasChosenLanguage) {
              return LanguageSelectionScreen(
                initialLanguage: _currentLanguage,
                onLanguageSelected: _onLanguageSelected,
              );
            }

            // 2. Unauthenticated: Login or Register
            if (_currentUser == null) {
              if (_showRegister) {
                return RegisterScreen(
                  apiClient: _apiClient,
                  onRegisterSuccess: _onLoginSuccess,
                  onBackToLogin: () => setState(() => _showRegister = false),
                );
              }
              return LoginScreen(
                apiClient: _apiClient,
                onLoginSuccess: _onLoginSuccess,
                onNavigateToRegister: () => setState(() => _showRegister = true),
              );
            }

            // 3. Authenticated: Delivery Agent Experience
            if (_currentExperience == AppExperience.deliveryAgent) {
              return DeliveryAgentHomeScreen(
                apiClient: _apiClient,
                onLogout: _onLogout,
                canSwitchExperience: _currentUser!.isCustomer || _currentUser!.isVendor,
                onSwitchExperience: () {
                  if (_currentUser!.isVendor) {
                    _switchToVendorExperience();
                  } else {
                    _switchToCustomerExperience();
                  }
                },
              );
            }

            // 4. Authenticated: Vendor Experience
            if (_currentExperience == AppExperience.vendor && _currentUser!.isVendor) {
              return VendorHomeScreen(
                apiClient: _apiClient,
                onLogout: _onLogout,
                onSwitchToCustomer: _switchToCustomerExperience,
              );
            }

            // 5. Authenticated: Customer Experience
            return CustomerHomeScreen(
              apiClient: _apiClient,
              onLogout: _onLogout,
              onSwitchToVendor: _currentUser!.isDeliveryAgent
                  ? _switchToDeliveryExperience
                  : _switchToVendorExperience,
              canSwitchToVendor: _currentUser!.isVendor || _currentUser!.isDeliveryAgent,
            );
          },
        ),
      ),
    );
  }
}

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import '../../core/api/api_client.dart';
import '../../core/design_system/app_colors.dart';
import '../../core/design_system/app_decorations.dart';
import '../../core/design_system/components/app_buttons.dart';
import '../../core/design_system/components/app_inputs.dart';
import '../../core/localization/locale_provider.dart';
import '../../shared/models/models.dart';

class LoginScreen extends StatefulWidget {
  final ApiClient apiClient;
  final Function(UserModel user) onLoginSuccess;
  final VoidCallback onNavigateToRegister;

  const LoginScreen({
    super.key,
    required this.apiClient,
    required this.onLoginSuccess,
    required this.onNavigateToRegister,
  });

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _phoneController = TextEditingController(text: '+919876500001');
  final _passwordController = TextEditingController(text: 'VendorPassword123!');
  bool _isLoading = false;
  bool _obscurePassword = true;
  String? _errorMessage;

  @override
  void dispose() {
    _phoneController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  String _localizeError(dynamic error, LocaleProvider loc) {
    final str = error.toString().toLowerCase();
    if (str.contains('invalid credentials') || str.contains('invalid_credentials')) {
      return loc.t('invalid_credentials');
    }
    if (str.contains('account is not active') || str.contains('not active') || str.contains('account_inactive')) {
      return loc.t('account_not_active');
    }
    if (str.contains('phone and password')) {
      return loc.t('phone_password_required');
    }
    return error.toString().replaceAll('ApiException', '').trim();
  }

  Future<void> _handleLogin() async {
    if (_isLoading) return;
    final loc = LocaleProvider.of(context);
    final phone = _phoneController.text.trim();
    final password = _passwordController.text.trim();

    if (phone.isEmpty || password.isEmpty) {
      setState(() => _errorMessage = loc.t('phone_password_required'));
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final res = await widget.apiClient.post('/auth/login', body: {
        'phone': phone,
        'password': password,
      });

      final token = res['access_token'] as String;
      widget.apiClient.setToken(token);

      final user = UserModel.fromJson(res['user']);

      if (user.preferredLanguage.isNotEmpty && mounted) {
        LocaleProvider.of(context).setLanguage(user.preferredLanguage);
      }

      widget.onLoginSuccess(user);
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = _localizeError(e, loc);
        });
      }
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  void _fillCredentials(String phone, String password) {
    setState(() {
      _phoneController.text = phone;
      _passwordController.text = password;
      _errorMessage = null;
    });
  }

  Widget _buildLanguagePill(String code, String label, LocaleProvider loc) {
    final isSelected = loc.currentLanguage == code;
    return GestureDetector(
      onTap: () {
        if (!isSelected) {
          loc.setLanguage(code);
        }
      },
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 180),
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
        decoration: BoxDecoration(
          color: isSelected ? AppColors.primary : Colors.transparent,
          borderRadius: AppRadii.roundedPill,
          boxShadow: isSelected ? AppShadows.card : null,
        ),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 12,
            fontWeight: isSelected ? FontWeight.bold : FontWeight.w500,
            color: isSelected ? Colors.white : AppColors.textSecondary,
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        actions: [
          Container(
            margin: const EdgeInsets.symmetric(vertical: 10, horizontal: 16),
            padding: const EdgeInsets.all(3),
            decoration: BoxDecoration(
              color: AppColors.mintSoft,
              borderRadius: AppRadii.roundedPill,
              border: Border.all(color: AppColors.mint, width: 1),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                _buildLanguagePill('en', 'English', loc),
                _buildLanguagePill('kn', 'ಕನ್ನಡ', loc),
              ],
            ),
          ),
        ],
      ),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 12.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Brand Avatar with forest green accent
                Center(
                  child: Container(
                    width: 76,
                    height: 76,
                    decoration: BoxDecoration(
                      color: AppColors.mintSoft,
                      shape: BoxShape.circle,
                      border: Border.all(color: AppColors.mint, width: 1.5),
                    ),
                    child: const Icon(
                      Icons.storefront_rounded,
                      size: 40,
                      color: AppColors.primary,
                    ),
                  ),
                ),
                const SizedBox(height: 18),
                Text(
                  loc.t('app_name'),
                  textAlign: TextAlign.center,
                  style: const TextStyle(
                    fontSize: 26,
                    fontWeight: FontWeight.w800,
                    color: AppColors.textPrimary,
                    letterSpacing: -0.4,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  loc.t('tagline'),
                  textAlign: TextAlign.center,
                  style: const TextStyle(
                    fontSize: 13,
                    color: AppColors.textSecondary,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 28),

                // Error alert banner if any
                if (_errorMessage != null)
                  Container(
                    padding: const EdgeInsets.all(12),
                    margin: const EdgeInsets.only(bottom: 16),
                    decoration: BoxDecoration(
                      color: AppColors.errorBg,
                      borderRadius: AppRadii.roundedMd,
                      border: Border.all(color: AppColors.error.withValues(alpha: 0.3), width: 1),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.error_outline_rounded, color: AppColors.error, size: 18),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            _errorMessage!,
                            style: const TextStyle(color: AppColors.errorText, fontSize: 13, fontWeight: FontWeight.w500),
                          ),
                        ),
                      ],
                    ),
                  ),

                // Phone Input Field
                AppTextField(
                  label: loc.t('phone'),
                  controller: _phoneController,
                  keyboardType: TextInputType.phone,
                  prefixIcon: Icons.phone_outlined,
                  hint: '+919876500001',
                ),
                const SizedBox(height: 16),

                // Password Input Field
                AppTextField(
                  label: loc.t('password'),
                  controller: _passwordController,
                  obscureText: _obscurePassword,
                  prefixIcon: Icons.lock_outline_rounded,
                  hint: '••••••••••••',
                  suffix: IconButton(
                    icon: Icon(
                      _obscurePassword ? Icons.visibility_off_outlined : Icons.visibility_outlined,
                      color: AppColors.textTertiary,
                      size: 20,
                    ),
                    onPressed: () => setState(() => _obscurePassword = !_obscurePassword),
                  ),
                ),
                const SizedBox(height: 24),

                // Sign In Button
                AppPrimaryButton(
                  label: loc.t('sign_in'),
                  icon: Icons.login_rounded,
                  isLoading: _isLoading,
                  onPressed: _isLoading ? null : _handleLogin,
                ),
                const SizedBox(height: 14),

                // Register Navigation Link
                Center(
                  child: TextButton(
                    onPressed: widget.onNavigateToRegister,
                    style: TextButton.styleFrom(
                      foregroundColor: AppColors.primary,
                      textStyle: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
                    ),
                    child: Text(loc.t('dont_have_account')),
                  ),
                ),
                const SizedBox(height: 20),

                // Demo Test Account Quick-Login Pills (Development only)
                if (kDebugMode || const bool.fromEnvironment('ALLOW_DEMO_ACCOUNTS', defaultValue: true)) ...[
                  const Divider(),
                  const SizedBox(height: 14),
                  Text(
                    loc.t('demo_accounts'),
                    textAlign: TextAlign.center,
                    style: const TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                      color: AppColors.textTertiary,
                      letterSpacing: 0.5,
                    ),
                  ),
                  const SizedBox(height: 10),
                  SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        AppSecondaryButton(
                          label: loc.currentLanguage == 'kn' ? 'ವ್ಯಾಪಾರಿ' : 'Vendor',
                          icon: Icons.store_rounded,
                          isFullWidth: false,
                          padding: const EdgeInsets.symmetric(vertical: 9, horizontal: 14),
                          onPressed: () => _fillCredentials('+919876500001', 'VendorPassword123!'),
                        ),
                        const SizedBox(width: 8),
                        AppSecondaryButton(
                          label: loc.currentLanguage == 'kn' ? 'ಗ್ರಾಹಕ' : 'Customer',
                          icon: Icons.shopping_basket_rounded,
                          isFullWidth: false,
                          padding: const EdgeInsets.symmetric(vertical: 9, horizontal: 14),
                          onPressed: () => _fillCredentials('+919888877777', 'CustomerPassword123!'),
                        ),
                        const SizedBox(width: 8),
                        AppSecondaryButton(
                          label: loc.currentLanguage == 'kn' ? 'ಡೆಲಿವರಿ ಸಿಬ್ಬಂದಿ' : 'Delivery Agent',
                          icon: Icons.two_wheeler_rounded,
                          isFullWidth: false,
                          padding: const EdgeInsets.symmetric(vertical: 9, horizontal: 14),
                          onPressed: () => _fillCredentials('+919888822222', 'CustomerPassword123!'),
                        ),
                      ],
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}

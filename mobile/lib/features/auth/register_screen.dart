import 'package:flutter/material.dart';
import '../../core/api/api_client.dart';
import '../../core/design_system/app_colors.dart';
import '../../core/design_system/app_decorations.dart';
import '../../core/design_system/components/app_buttons.dart';
import '../../core/design_system/components/app_inputs.dart';
import '../../core/localization/locale_provider.dart';
import '../../shared/models/models.dart';

class RegisterScreen extends StatefulWidget {
  final ApiClient apiClient;
  final Function(UserModel user) onRegisterSuccess;
  final VoidCallback onBackToLogin;

  const RegisterScreen({
    super.key,
    required this.apiClient,
    required this.onRegisterSuccess,
    required this.onBackToLogin,
  });

  @override
  State<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends State<RegisterScreen> {
  final _phoneController = TextEditingController(text: '+91');
  final _passwordController = TextEditingController();
  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  bool _isLoading = false;
  String? _errorMessage;

  @override
  void dispose() {
    _phoneController.dispose();
    _passwordController.dispose();
    _nameController.dispose();
    _emailController.dispose();
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

  Future<void> _handleRegister() async {
    final loc = LocaleProvider.of(context);
    final phone = _phoneController.text.trim();
    final password = _passwordController.text.trim();
    final name = _nameController.text.trim();
    final email = _emailController.text.trim();

    if (phone.isEmpty || password.isEmpty) {
      setState(() => _errorMessage = loc.t('phone_password_required'));
      return;
    }
    if (password.length < 8) {
      setState(() => _errorMessage = loc.t('password_min_length'));
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final res = await widget.apiClient.post('/auth/register', body: {
        'phone': phone,
        'password': password,
        'full_name': name.isNotEmpty ? name : null,
        'email': email.isNotEmpty ? email : null,
        'preferred_language': loc.currentLanguage,
      });

      final token = res['access_token'] as String;
      widget.apiClient.setToken(token);

      final user = UserModel.fromJson(res['user']);
      widget.onRegisterSuccess(user);
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

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_rounded, color: AppColors.textPrimary),
          onPressed: widget.onBackToLogin,
        ),
      ),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 8.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Center(
                  child: Container(
                    width: 68,
                    height: 68,
                    decoration: BoxDecoration(
                      color: AppColors.mintSoft,
                      shape: BoxShape.circle,
                      border: Border.all(color: AppColors.mint, width: 1.5),
                    ),
                    child: const Icon(
                      Icons.person_add_alt_1_rounded,
                      size: 34,
                      color: AppColors.primary,
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                Text(
                  loc.t('create_account'),
                  textAlign: TextAlign.center,
                  style: const TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.w800,
                    color: AppColors.textPrimary,
                    letterSpacing: -0.3,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  loc.t('tagline'),
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 13, color: AppColors.textSecondary),
                ),
                const SizedBox(height: 24),

                if (_errorMessage != null)
                  Container(
                    padding: const EdgeInsets.all(12),
                    margin: const EdgeInsets.only(bottom: 16),
                    decoration: BoxDecoration(
                      color: AppColors.errorBg,
                      borderRadius: AppRadii.roundedMd,
                      border: Border.all(color: AppColors.error.withValues(alpha: 0.3)),
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

                AppTextField(
                  label: loc.t('phone'),
                  controller: _phoneController,
                  keyboardType: TextInputType.phone,
                  prefixIcon: Icons.phone_outlined,
                  hint: '+919876500000',
                ),
                const SizedBox(height: 14),

                AppTextField(
                  label: loc.t('password'),
                  controller: _passwordController,
                  obscureText: true,
                  prefixIcon: Icons.lock_outline_rounded,
                  hint: 'Min 8 characters',
                ),
                const SizedBox(height: 14),

                AppTextField(
                  label: loc.t('full_name'),
                  controller: _nameController,
                  prefixIcon: Icons.badge_outlined,
                  hint: 'Ramesh Gowda',
                ),
                const SizedBox(height: 14),

                AppTextField(
                  label: loc.t('email'),
                  controller: _emailController,
                  keyboardType: TextInputType.emailAddress,
                  prefixIcon: Icons.mail_outline_rounded,
                  hint: 'name@example.com',
                ),
                const SizedBox(height: 24),

                AppPrimaryButton(
                  label: loc.t('register'),
                  icon: Icons.check_circle_outline_rounded,
                  isLoading: _isLoading,
                  onPressed: _handleRegister,
                ),
                const SizedBox(height: 14),

                Center(
                  child: TextButton(
                    onPressed: widget.onBackToLogin,
                    style: TextButton.styleFrom(
                      foregroundColor: AppColors.primary,
                      textStyle: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
                    ),
                    child: Text(loc.t('already_have_account')),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

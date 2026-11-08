import 'package:flutter/material.dart';
import '../../core/api/api_client.dart';
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

  Future<void> _handleRegister() async {
    final phone = _phoneController.text.trim();
    final password = _passwordController.text.trim();
    final name = _nameController.text.trim();
    final email = _emailController.text.trim();

    if (phone.isEmpty || password.isEmpty) {
      setState(() => _errorMessage = 'Please enter phone and password');
      return;
    }
    if (password.length < 8) {
      setState(() => _errorMessage = 'Password must be at least 8 characters');
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final body = <String, dynamic>{
        'phone': phone,
        'password': password,
      };
      if (email.isNotEmpty) {
        body['email'] = email;
      }

      final res = await widget.apiClient.post('/auth/register', body: body);
      final token = res['access_token'] as String;
      widget.apiClient.setToken(token);

      // If name was provided, update profile
      if (name.isNotEmpty) {
        try {
          await widget.apiClient.patch('/customers/profile/me', body: {
            'display_name': name,
          });
        } catch (_) {}
      }

      final user = UserModel.fromJson(res['user']);
      widget.onRegisterSuccess(user);
    } catch (e) {
      setState(() {
        _errorMessage = e.toString().replaceAll('ApiException', '');
      });
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(loc.t('create_account')),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: widget.onBackToLogin,
        ),
        actions: [
          TextButton.icon(
            onPressed: () {
              final newLang = loc.currentLanguage == 'en' ? 'kn' : 'en';
              loc.setLanguage(newLang);
              LanguageController.saveLanguage(newLang, apiClient: widget.apiClient);
            },
            icon: const Icon(Icons.language, size: 18),
            label: Text(loc.currentLanguage == 'en' ? 'ಕನ್ನಡ' : 'English'),
          ),
        ],
      ),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Icon(Icons.person_add_rounded, size: 64, color: Color(0xFF1E40AF)),
                const SizedBox(height: 12),
                Text(
                  loc.t('create_account'),
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: Color(0xFF0F172A)),
                ),
                const SizedBox(height: 4),
                Text(
                  loc.t('tagline'),
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 13, color: Color(0xFF64748B)),
                ),
                const SizedBox(height: 24),
                if (_errorMessage != null)
                  Container(
                    padding: const EdgeInsets.all(12),
                    margin: const EdgeInsets.only(bottom: 16),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFEE2E2),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: const Color(0xFFFECACA), width: 1),
                    ),
                    child: Text(
                      _errorMessage!,
                      style: const TextStyle(color: Color(0xFF991B1B), fontSize: 13),
                    ),
                  ),
                TextField(
                  controller: _nameController,
                  decoration: InputDecoration(
                    labelText: loc.t('full_name'),
                    prefixIcon: const Icon(Icons.badge_outlined),
                    border: const OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _phoneController,
                  keyboardType: TextInputType.phone,
                  decoration: InputDecoration(
                    labelText: loc.t('phone'),
                    prefixIcon: const Icon(Icons.phone_outlined),
                    border: const OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _emailController,
                  keyboardType: TextInputType.emailAddress,
                  decoration: InputDecoration(
                    labelText: loc.t('email'),
                    prefixIcon: const Icon(Icons.email_outlined),
                    border: const OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _passwordController,
                  obscureText: true,
                  decoration: InputDecoration(
                    labelText: loc.t('password'),
                    prefixIcon: const Icon(Icons.lock_outline),
                    border: const OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 24),
                ElevatedButton(
                  onPressed: _isLoading ? null : _handleRegister,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF1E40AF),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                  ),
                  child: _isLoading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2),
                        )
                      : Text(
                          loc.t('register'),
                          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                        ),
                ),
                const SizedBox(height: 16),
                TextButton(
                  onPressed: widget.onBackToLogin,
                  child: Text(loc.t('already_have_account')),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

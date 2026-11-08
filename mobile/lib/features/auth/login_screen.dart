import 'package:flutter/material.dart';
import '../../core/api/api_client.dart';
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
  String? _errorMessage;

  Future<void> _handleLogin() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final res = await widget.apiClient.post('/auth/login', body: {
        'phone': _phoneController.text.trim(),
        'password': _passwordController.text.trim(),
      });

      final token = res['access_token'] as String;
      widget.apiClient.setToken(token);

      final user = UserModel.fromJson(res['user']);

      // If user has a preferred language saved on profile, apply it
      if (user.preferredLanguage.isNotEmpty && mounted) {
        LocaleProvider.of(context).setLanguage(user.preferredLanguage);
      }

      widget.onLoginSuccess(user);
    } catch (e) {
      setState(() {
        _errorMessage = e.toString().replaceAll('ApiException', '');
      });
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
    });
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
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
          const SizedBox(width: 8),
        ],
      ),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 8.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Icon(Icons.storefront_rounded, size: 72, color: Color(0xFF1E40AF)),
                const SizedBox(height: 16),
                Text(
                  loc.t('app_name'),
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 28, fontWeight: FontWeight.bold, color: Color(0xFF0F172A)),
                ),
                const SizedBox(height: 4),
                Text(
                  loc.t('tagline'),
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 13, color: Color(0xFF64748B)),
                ),
                const SizedBox(height: 28),
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
                  onPressed: _isLoading ? null : _handleLogin,
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
                          loc.t('sign_in'),
                          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                        ),
                ),
                const SizedBox(height: 16),
                TextButton(
                  onPressed: widget.onNavigateToRegister,
                  child: Text(loc.t('dont_have_account')),
                ),
                const SizedBox(height: 20),
                const Divider(),
                const SizedBox(height: 12),
                Text(
                  loc.currentLanguage == 'kn' ? 'ಪರೀಕ್ಷಾ ಖಾತೆಗಳು (Quick Fill):' : 'Demo Test Accounts:',
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: Color(0xFF64748B)),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: OutlinedButton(
                        onPressed: () => _fillCredentials('+919876500001', 'VendorPassword123!'),
                        style: OutlinedButton.styleFrom(
                          padding: const EdgeInsets.symmetric(vertical: 8),
                        ),
                        child: const Text('Vendor Suresh', style: TextStyle(fontSize: 11)),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: OutlinedButton(
                        onPressed: () => _fillCredentials('+919888877777', 'CustomerPassword123!'),
                        style: OutlinedButton.styleFrom(
                          padding: const EdgeInsets.symmetric(vertical: 8),
                        ),
                        child: const Text('Customer Ramesh', style: TextStyle(fontSize: 11)),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

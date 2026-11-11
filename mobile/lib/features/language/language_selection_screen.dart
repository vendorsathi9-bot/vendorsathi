import 'package:flutter/material.dart';
import '../../core/design_system/app_colors.dart';
import '../../core/design_system/app_decorations.dart';
import '../../core/design_system/components/app_buttons.dart';
import '../../core/localization/locale_provider.dart';

class LanguageSelectionScreen extends StatefulWidget {
  final Function(String langCode) onLanguageSelected;
  final String initialLanguage;

  const LanguageSelectionScreen({
    super.key,
    required this.onLanguageSelected,
    this.initialLanguage = 'en',
  });

  @override
  State<LanguageSelectionScreen> createState() => _LanguageSelectionScreenState();
}

class _LanguageSelectionScreenState extends State<LanguageSelectionScreen> {
  late String _selectedLanguage;

  @override
  void initState() {
    super.initState();
    _selectedLanguage = widget.initialLanguage;
  }

  void _handleContinue() async {
    await LanguageController.saveLanguage(_selectedLanguage);
    widget.onLanguageSelected(_selectedLanguage);
  }

  Widget _buildOptionCard({
    required String title,
    required String subtitle,
    required String scriptLetter,
    required String code,
    required bool isSelected,
  }) {
    return PressableScale(
      onPressed: () => setState(() => _selectedLanguage = code),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 180),
        padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 16),
        decoration: BoxDecoration(
          color: isSelected ? AppColors.mintSoft : AppColors.surface,
          borderRadius: AppRadii.roundedLg,
          border: Border.all(
            color: isSelected ? AppColors.primary : AppColors.border,
            width: isSelected ? 2.0 : 1.0,
          ),
          boxShadow: isSelected ? AppShadows.card : null,
        ),
        child: Row(
          children: [
            // Letter avatar circle
            Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                color: isSelected ? AppColors.primary : AppColors.surfaceMuted,
                borderRadius: AppRadii.roundedMd,
              ),
              child: Center(
                child: Text(
                  scriptLetter,
                  style: TextStyle(
                    fontSize: 22,
                    fontWeight: FontWeight.bold,
                    color: isSelected ? Colors.white : AppColors.textSecondary,
                  ),
                ),
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: TextStyle(
                      fontSize: 17,
                      fontWeight: FontWeight.w700,
                      color: isSelected ? AppColors.primaryDark : AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 3),
                  Text(
                    subtitle,
                    style: TextStyle(
                      fontSize: 13,
                      color: isSelected ? AppColors.primary : AppColors.textSecondary,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ],
              ),
            ),
            // Custom Radio Circle
            Container(
              width: 22,
              height: 22,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                border: Border.all(
                  color: isSelected ? AppColors.primary : AppColors.border,
                  width: 2,
                ),
                color: isSelected ? AppColors.primary : Colors.transparent,
              ),
              child: isSelected
                  ? const Center(
                      child: Icon(Icons.check, size: 14, color: Colors.white),
                    )
                  : null,
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final isKn = _selectedLanguage == 'kn';

    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 28.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Spacer(),
              // Top hero icon
              Center(
                child: Container(
                  width: 84,
                  height: 84,
                  decoration: BoxDecoration(
                    color: AppColors.mintSoft,
                    shape: BoxShape.circle,
                    border: Border.all(color: AppColors.mint, width: 1.5),
                  ),
                  child: const Icon(
                    Icons.translate_rounded,
                    size: 42,
                    color: AppColors.primary,
                  ),
                ),
              ),
              const SizedBox(height: 24),
              const Text(
                'Choose your preferred language',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 22,
                  fontWeight: FontWeight.w800,
                  color: AppColors.textPrimary,
                  letterSpacing: -0.4,
                ),
              ),
              const SizedBox(height: 6),
              const Text(
                'ನಿಮ್ಮ ಆದ್ಯತೆಯ ಭಾಷೆಯನ್ನು ಆಯ್ಕೆಮಾಡಿ',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 15,
                  color: AppColors.textSecondary,
                  fontWeight: FontWeight.w500,
                ),
              ),
              const SizedBox(height: 36),

              // English Option Card
              _buildOptionCard(
                title: 'English',
                subtitle: 'Continue in English',
                scriptLetter: 'A',
                code: 'en',
                isSelected: _selectedLanguage == 'en',
              ),
              const SizedBox(height: 14),

              // Kannada Option Card
              _buildOptionCard(
                title: 'ಕನ್ನಡ',
                subtitle: 'ಕನ್ನಡದಲ್ಲಿ ಮುಂದುವರಿಯಿರಿ',
                scriptLetter: 'ಕ',
                code: 'kn',
                isSelected: _selectedLanguage == 'kn',
              ),

              const Spacer(),

              // Continue button
              AppPrimaryButton(
                label: isKn ? 'ಮುಂದುವರಿಯಿರಿ' : 'Continue',
                icon: Icons.arrow_forward_rounded,
                onPressed: _handleContinue,
              ),
              const SizedBox(height: 8),
            ],
          ),
        ),
      ),
    );
  }
}

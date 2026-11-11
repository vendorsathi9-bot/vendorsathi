import 'package:flutter/material.dart';

/// VendorSaathi Design System — Color Palette
/// Visual Direction: Deep forest green brand, warm off-white canvas, soft mint accents, dark charcoal text.
class AppColors {
  // Brand Primary — Deep Forest Green
  static const Color primary = Color(0xFF1B4332);
  static const Color primaryDark = Color(0xFF133024);
  static const Color primaryLight = Color(0xFF2D6A4F);
  static const Color primaryHover = Color(0xFF225740);

  // Brand Secondary & Accents — Mint & Emerald
  static const Color accent = Color(0xFF40916C);
  static const Color accentLight = Color(0xFF52B788);
  static const Color mint = Color(0xFFD8F3DC);
  static const Color mintSoft = Color(0xFFE8F5EE);
  static const Color mintLight = Color(0xFFF2F9F5);

  // Backgrounds & Canvas — Warm Off-White
  static const Color background = Color(0xFFFAF9F6);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color surfaceMuted = Color(0xFFF4F6F4);
  static const Color surfaceElevated = Color(0xFFFFFFFF);

  // Typography — Charcoal
  static const Color textPrimary = Color(0xFF19241E);
  static const Color textSecondary = Color(0xFF5A6B62);
  static const Color textTertiary = Color(0xFF8A9A92);
  static const Color textWhite = Color(0xFFFFFFFF);

  // Borders & Dividers
  static const Color border = Color(0xFFE2E8E4);
  static const Color borderLight = Color(0xFFEDF2EE);
  static const Color borderFocused = Color(0xFF1B4332);
  static const Color divider = Color(0xFFECEFE7);

  // Status & Feedback
  static const Color success = Color(0xFF16A34A);
  static const Color successBg = Color(0xFFDCFCE7);
  static const Color successText = Color(0xFF14532D);

  static const Color warning = Color(0xFFD97706);
  static const Color warningBg = Color(0xFFFEF3C7);
  static const Color warningText = Color(0xFF78350F);

  static const Color error = Color(0xFFDC2626);
  static const Color errorBg = Color(0xFFFEE2E2);
  static const Color errorText = Color(0xFF7F1D1D);

  static const Color info = Color(0xFF2563EB);
  static const Color infoBg = Color(0xFFDBEAFE);
  static const Color infoText = Color(0xFF1E3A8A);

  // Neutral Grays for disabled & skeleton states
  static const Color disabled = Color(0xFFCBD5E1);
  static const Color skeletonBase = Color(0xFFE6EBE8);
  static const Color skeletonHighlight = Color(0xFFF4F8F6);
}

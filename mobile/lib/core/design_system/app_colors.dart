import 'package:flutter/material.dart';

/// VendorSaathi Design System — Color Palette
/// Source of truth: Landing page (tailwind.config.ts)
/// Brand: Warm cream/amber/brown — NOT green/mint
class AppColors {
  // Brand Primary — Golden Amber
  static const Color primary = Color(0xFFF4B400);        // #F4B400
  static const Color primaryDark = Color(0xFFDCA000);    // #DCA000
  static const Color primaryLight = Color(0xFFF7C933);   // Lighter amber for hover/active
  static const Color primaryHover = Color(0xFFE6A800);   // Between primary and primaryDark

  // Brand Accent — Soft Cream
  static const Color accent = Color(0xFFFEF4CD);         // #FEF4CD - soft cream accent
  static const Color accentLight = Color(0xFFFEF9E7);    // Lighter cream
  static const Color accentSoft = Color(0xFFFDF6D8);     // Subtle cream background

  // Backgrounds & Canvas — Warm Cream
  static const Color background = Color(0xFFFFFCEF);     // #FFFCEF - main page background
  static const Color surface = Color(0xFFFFFFFF);        // #FFFFFF - cards and surfaces
  static const Color surfaceMuted = Color(0xFFFFF8E8);   // Slightly warmer than surface
  static const Color surfaceElevated = Color(0xFFFFFFFF); // Elevated surfaces

  // Typography — Espresso & Brown
  static const Color textPrimary = Color(0xFF38200F);    // #38200F - espresso, primary text
  static const Color textSecondary = Color(0xFF806B55);  // #806B55 - brown, secondary text
  static const Color textTertiary = Color(0xFFA6937D);   // Muted brown for tertiary
  static const Color textWhite = Color(0xFFFFFFFF);      // White text on dark backgrounds

  // Borders & Dividers
  static const Color border = Color(0xFFEAE1D1);         // #EAE1D1 - warm beige borders
  static const Color borderLight = Color(0xFFF3EBD8);    // Lighter border
  static const Color borderFocused = Color(0xFFF4B400);  // Primary amber for focus
  static const Color divider = Color(0xFFEAE1D1);        // Same as border

  // Status & Feedback — Preserve semantic meaning, align with brand
  static const Color success = Color(0xFF2F7D4F);        // #2F7D4F - success green (from landing)
  static const Color successBg = Color(0xFFEAF6EE);      // #EAF6EE - success background
  static const Color successText = Color(0xFF1E5A35);    // Darker success text

  static const Color warning = Color(0xFFF4B400);        // Use primary amber for warning
  static const Color warningBg = Color(0xFFFEF4CD);      // Soft cream background
  static const Color warningText = Color(0xFF8B6D00);    // Dark amber text

  static const Color error = Color(0xFFDC2626);          // Standard error red
  static const Color errorBg = Color(0xFFFEE2E2);        // Light error background
  static const Color errorText = Color(0xFF7F1D1D);      // Dark error text

  static const Color info = Color(0xFF2563EB);           // Standard info blue
  static const Color infoBg = Color(0xFFDBEAFE);         // Light info background
  static const Color infoText = Color(0xFF1E3A8A);       // Dark info text

  // Neutral for disabled & skeleton states
  static const Color disabled = Color(0xFFD4C4A8);       // Warm disabled
  static const Color skeletonBase = Color(0xFFF3EBD8);   // Warm skeleton base
  static const Color skeletonHighlight = Color(0xFFF9F2E7); // Warm skeleton highlight

  // Brand Accent Compatibility Aliases (maps old mint tokens to warm soft cream)
  static const Color mint = accent;                      // #FEF4CD
  static const Color mintSoft = accentSoft;              // #FDF6D8
  static const Color mintLight = accentLight;            // #FEF9E7

  // Espresso-based shadow colors (matching landing page shadows)
  static const Color shadowEspresso = Color(0xFF38200F); // Espresso for shadows
  static const Color shadowEspresso05 = Color(0x0D38200F); // ~5% opacity
  static const Color shadowEspresso10 = Color(0x1A38200F); // ~10% opacity
  static const Color shadowEspresso20 = Color(0x3338200F); // ~20% opacity
  static const Color shadowEspresso30 = Color(0x4D38200F); // ~30% opacity
  static const Color shadowEspresso40 = Color(0x6638200F); // ~40% opacity
  static const Color shadowEspresso50 = Color(0x8038200F); // ~50% opacity
}
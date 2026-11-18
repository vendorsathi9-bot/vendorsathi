import 'package:flutter/material.dart';
import 'app_colors.dart';

/// VendorSaathi Design System — Spacing, Radii, and Shadows
class AppSpacing {
  static const double xxs = 4.0;
  static const double xs = 8.0;
  static const double sm = 12.0;
  static const double md = 16.0;
  static const double lg = 20.0;
  static const double xl = 24.0;
  static const double xxl = 32.0;
}

class AppRadii {
  static const double sm = 8.0;
  static const double md = 12.0;
  static const double lg = 16.0;
  static const double xl = 20.0;
  static const double pill = 999.0;

  static const BorderRadius roundedSm = BorderRadius.all(Radius.circular(sm));
  static const BorderRadius roundedMd = BorderRadius.all(Radius.circular(md));
  static const BorderRadius roundedLg = BorderRadius.all(Radius.circular(lg));
  static const BorderRadius roundedXl = BorderRadius.all(Radius.circular(xl));
  static const BorderRadius roundedPill = BorderRadius.all(Radius.circular(pill));
}

class AppShadows {
  /// Subtle, elegant card elevation — espresso-based
  static const List<BoxShadow> card = [
    BoxShadow(
      color: AppColors.shadowEspresso10,
      blurRadius: 8,
      offset: Offset(0, 2),
    ),
    BoxShadow(
      color: AppColors.shadowEspresso05,
      blurRadius: 1,
      offset: Offset(0, 1),
    ),
  ];

  /// Elevated component (bottom navigation, modals, popovers) — espresso-based
  static const List<BoxShadow> elevated = [
    BoxShadow(
      color: AppColors.shadowEspresso20,
      blurRadius: 16,
      offset: Offset(0, 4),
    ),
    BoxShadow(
      color: AppColors.shadowEspresso10,
      blurRadius: 4,
      offset: Offset(0, 2),
    ),
  ];

  /// Floating button or sticky bar shadow — espresso-based
  static const List<BoxShadow> floating = [
    BoxShadow(
      color: AppColors.shadowEspresso30,
      blurRadius: 16,
      offset: Offset(0, 6),
    ),
  ];

  /// Prominent shadow for dialogs, bottom sheets
  static const List<BoxShadow> modal = [
    BoxShadow(
      color: AppColors.shadowEspresso20,
      blurRadius: 24,
      offset: Offset(0, 8),
    ),
    BoxShadow(
      color: AppColors.shadowEspresso10,
      blurRadius: 8,
      offset: Offset(0, 4),
    ),
  ];

  /// Inner shadow for pressed states
  static const List<BoxShadow> inner = [
    BoxShadow(
      color: AppColors.shadowEspresso20,
      blurRadius: 4,
      offset: Offset(0, 2),
    ),
  ];
}
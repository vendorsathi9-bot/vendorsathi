import 'package:flutter/material.dart';

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
  /// Subtle, elegant card elevation
  static const List<BoxShadow> card = [
    BoxShadow(
      color: Color(0x08000000),
      blurRadius: 8,
      offset: Offset(0, 2),
    ),
    BoxShadow(
      color: Color(0x05000000),
      blurRadius: 1,
      offset: Offset(0, 1),
    ),
  ];

  /// Elevated component (bottom navigation, modals, popovers)
  static const List<BoxShadow> elevated = [
    BoxShadow(
      color: Color(0x10000000),
      blurRadius: 16,
      offset: Offset(0, 4),
    ),
    BoxShadow(
      color: Color(0x06000000),
      blurRadius: 4,
      offset: Offset(0, 2),
    ),
  ];

  /// Floating button or sticky bar shadow
  static const List<BoxShadow> floating = [
    BoxShadow(
      color: Color(0x181B4332), // primary tint shadow
      blurRadius: 16,
      offset: Offset(0, 6),
    ),
  ];
}

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:vendorsaathi_mobile/core/localization/locale_provider.dart';
import 'package:vendorsaathi_mobile/core/map/map_style_config.dart';
import 'package:vendorsaathi_mobile/main.dart';
import 'package:vendorsaathi_mobile/shared/models/models.dart';
import 'package:vendorsaathi_mobile/shared/widgets/osm_map_view.dart';
import 'package:vendorsaathi_mobile/shared/widgets/sales_trend_chart.dart';

void main() {
  testWidgets('App renders branded splash screen on startup and transitions to language selection', (WidgetTester tester) async {
    await tester.pumpWidget(const VendorSaathiApp(initialLanguage: null, skipSplash: false));
    await tester.pump();

    // Verify splash screen branding elements are visible
    expect(find.text('VendorSaathi'), findsOneWidget);
    expect(find.byIcon(Icons.storefront_rounded), findsOneWidget);

    // Fast-forward past the splash timer (1200ms)
    await tester.pump(const Duration(milliseconds: 1300));
    await tester.pumpAndSettle();

    // Transitions to first-launch language selection
    expect(find.text('Choose your preferred language'), findsOneWidget);
    expect(find.text('English'), findsOneWidget);
    expect(find.text('ಕನ್ನಡ'), findsOneWidget);
  });

  testWidgets('App renders language selection on first launch when skipping splash', (WidgetTester tester) async {
    // When initialLanguage is null (first launch), language selection screen is rendered
    await tester.pumpWidget(const VendorSaathiApp(initialLanguage: null, skipSplash: true));
    await tester.pump();

    expect(find.text('Choose your preferred language'), findsOneWidget);
    expect(find.text('English'), findsOneWidget);
    expect(find.text('ಕನ್ನಡ'), findsOneWidget);
  });

  testWidgets('App renders login screen when language is pre-selected', (WidgetTester tester) async {
    await tester.pumpWidget(const VendorSaathiApp(initialLanguage: 'en', skipSplash: true));
    await tester.pump();

    expect(find.text('VendorSaathi'), findsOneWidget);
    expect(find.text('Sign In'), findsWidgets);
  });

  test('RouteModel parses OSRM GeoJSON response and coordinates accurately', () {
    final osrmJson = {
      'routing_available': true,
      'distance_meters': 5420.5,
      'duration_seconds': 780.0,
      'distance_km': 5.42,
      'duration_minutes': 13.0,
      'geometry': {
        'type': 'LineString',
        'coordinates': [
          [77.5727, 12.9778],
          [77.5946, 12.9716],
          [77.6245, 12.9352],
        ],
      },
      'attribution': '© OpenStreetMap contributors',
      'directions_url': 'https://www.openstreetmap.org/directions?engine=fossgis_osrm_car&route=12.9778%2C77.5727%3B12.9352%2C77.6245',
      'provider': 'osrm',
    };

    final route = RouteModel.fromJson(osrmJson);

    expect(route.routingAvailable, true);
    expect(route.distanceMeters, 5420.5);
    expect(route.durationSeconds, 780.0);
    expect(route.distanceKm, 5.42);
    expect(route.durationMinutes, 13.0);
    expect(route.coordinates.length, 3);
    expect(route.coordinates[0], [77.5727, 12.9778]); // [lng, lat]
    expect(route.attribution, '© OpenStreetMap contributors');
    expect(route.provider, 'osrm');
    expect(route.directionsUrl, contains('openstreetmap.org'));
  });

  test('RouteModel parses OSRM fallback with straight-line distance correctly', () {
    final fallbackJson = {
      'routing_available': false,
      'provider': 'osrm',
      'distance_type': 'straight_line_estimate',
      'approximate_straight_line_km': 4.85,
      'distance_km': 0.0,
      'duration_minutes': 0.0,
      'distance_meters': 0.0,
      'duration_seconds': 0.0,
      'attribution': '© OpenStreetMap contributors, OSRM',
      'message': 'Road routing currently unavailable from provider. Approximate straight-line distance is 4.85 km (road travel ETA unavailable).',
      'directions_url': 'https://www.openstreetmap.org/directions?engine=fossgis_osrm_car&route=12.9778%2C77.5727%3B12.9352%2C77.6245',
    };

    final route = RouteModel.fromJson(fallbackJson);
    expect(route.routingAvailable, false);
    expect(route.distanceType, 'straight_line_estimate');
    expect(route.approximateStraightLineKm, 4.85);
    expect(route.distanceKm, 0.0);
    expect(route.durationMinutes, 0.0);
    expect(route.coordinates.isEmpty, true);
    expect(route.message, contains('Approximate straight-line distance'));
  });

  test('NearbyVendorModel parses latitude, longitude and distance accurately', () {
    final vendorJson = {
      'vendor_id': '550e8400-e29b-41d4-a716-446655440000',
      'business_name': 'Ramesh Fresh Fruits',
      'phone': '9876543210',
      'status': 'ACTIVE',
      'availability_status': 'OPEN',
      'distance_km': 1.45,
      'latitude': 12.9716,
      'longitude': 77.5946,
      'address_line1': 'Gandhi Bazaar Main Rd',
      'city': 'Bangalore',
    };

    final vendor = NearbyVendorModel.fromJson(vendorJson);
    expect(vendor.vendorId, '550e8400-e29b-41d4-a716-446655440000');
    expect(vendor.businessName, 'Ramesh Fresh Fruits');
    expect(vendor.latitude, 12.9716);
    expect(vendor.longitude, 77.5946);
    expect(vendor.distanceKm, 1.45);
  });

  testWidgets('OsmMapView displays mandatory attribution and navigation action', (WidgetTester tester) async {
    final route = RouteModel(
      distanceKm: 3.5,
      durationMinutes: 10,
      routingAvailable: true,
      mapsApiAvailable: true,
      directionsUrl: 'https://www.openstreetmap.org',
      attribution: '© OpenStreetMap contributors',
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: OsmMapView(
            vendorLat: 12.9778,
            vendorLng: 77.5727,
            customerLat: 12.9352,
            customerLng: 77.6245,
            route: route,
          ),
        ),
      ),
    );
    await tester.pump();

    // Verify OpenStreetMap attribution is visible
    expect(find.textContaining('OpenStreetMap contributors'), findsOneWidget);
    expect(find.text('OSM Directions'), findsOneWidget);
    expect(find.text('Vendor'), findsOneWidget);
    expect(find.text('Customer'), findsOneWidget);
  });

  testWidgets('SalesTrendChart displays empty state when no sales exist', (WidgetTester tester) async {
    final emptyTrends = [
      SalesTrendModel(date: '2026-11-03', orderCount: 0, salesPaise: 0, profitPaise: 0),
      SalesTrendModel(date: '2026-11-04', orderCount: 0, salesPaise: 0, profitPaise: 0),
    ];

    await tester.pumpWidget(
      LocaleProvider(
        currentLanguage: 'en',
        setLanguage: (_) {},
        child: MaterialApp(
          home: Scaffold(
            body: SalesTrendChart(trends: emptyTrends),
          ),
        ),
      ),
    );
    await tester.pump();

    expect(find.text('No sales recorded in the past 7 days'), findsOneWidget);
  });

  testWidgets('SalesTrendChart renders sales bars and totals when data exists', (WidgetTester tester) async {
    final trends = [
      SalesTrendModel(date: '2026-11-06', orderCount: 0, salesPaise: 0, profitPaise: 0),
      SalesTrendModel(date: '2026-11-07', orderCount: 2, salesPaise: 60000, profitPaise: 9000),
    ];

    await tester.pumpWidget(
      LocaleProvider(
        currentLanguage: 'en',
        setLanguage: (_) {},
        child: MaterialApp(
          home: Scaffold(
            body: SalesTrendChart(trends: trends),
          ),
        ),
      ),
    );
    await tester.pump();

    expect(find.text('Sales Trends'), findsOneWidget);
    expect(find.text('₹600.00'), findsWidgets);
    expect(find.text('2'), findsWidgets);
  });

  testWidgets('Kannada language selection renders Kannada text immediately', (WidgetTester tester) async {
    await tester.pumpWidget(const VendorSaathiApp(initialLanguage: 'kn', skipSplash: true));
    await tester.pump();

    expect(find.text('ವೆಂಡರ್‌ಸಾಥಿ'), findsOneWidget);
    expect(find.text('ಸೈನ್ ಇನ್'), findsWidgets);
    expect(find.text('ದೂರವಾಣಿ ಸಂಖ್ಯೆ'), findsOneWidget);
    expect(find.text('ಪಾಸ್‌ವರ್ಡ್'), findsOneWidget);
  });

  test('MapStyleConfig generates raster style JSON when tile_url is configured', () {
    final cfg = MapStyleConfig.fromApi({
      'provider': 'carto',
      'tile_url': 'https://a.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}.png',
      'attribution': '© OpenStreetMap contributors, © CARTO',
    });

    expect(cfg.provider, 'carto');
    expect(cfg.tileUrl, contains('cartocdn.com'));
    expect(cfg.effectiveStyleString, contains('"type":"raster"'));
    expect(cfg.effectiveStyleString, contains('cartocdn.com'));
    expect(cfg.effectiveStyleString, contains('© OpenStreetMap contributors, © CARTO'));
  });

  test('MapStyleConfig uses vector style URL when style_url is configured', () {
    final cfg = MapStyleConfig.fromApi({
      'provider': 'openfreemap',
      'style_url': 'https://tiles.openfreemap.org/styles/liberty',
      'attribution': '© OpenFreeMap',
    });

    expect(cfg.provider, 'openfreemap');
    expect(cfg.effectiveStyleString, 'https://tiles.openfreemap.org/styles/liberty');
    expect(cfg.isValid, true);
  });
}

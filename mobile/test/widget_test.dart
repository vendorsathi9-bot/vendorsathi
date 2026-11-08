import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:vendorsaathi_mobile/main.dart';
import 'package:vendorsaathi_mobile/shared/models/models.dart';
import 'package:vendorsaathi_mobile/shared/widgets/osm_map_view.dart';

void main() {
  testWidgets('App renders language selection on first launch', (WidgetTester tester) async {
    // When initialLanguage is null (first launch), language selection screen is rendered
    await tester.pumpWidget(const VendorSaathiApp(initialLanguage: null));
    await tester.pump();

    expect(find.text('Choose your preferred language'), findsOneWidget);
    expect(find.text('English'), findsOneWidget);
    expect(find.text('ಕನ್ನಡ'), findsOneWidget);
  });

  testWidgets('App renders login screen when language is pre-selected', (WidgetTester tester) async {
    await tester.pumpWidget(const VendorSaathiApp(initialLanguage: 'en'));
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
    expect(find.text('© OpenStreetMap contributors'), findsOneWidget);
    expect(find.text('OSM Directions'), findsOneWidget);
    expect(find.text('Vendor'), findsOneWidget);
    expect(find.text('Customer'), findsOneWidget);
  });
}

import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../core/api/api_client.dart';
import '../../core/design_system/app_colors.dart';
import '../../core/design_system/app_decorations.dart';
import '../../core/design_system/components/app_badges.dart';
import '../../core/design_system/components/app_buttons.dart';
import '../../core/localization/locale_provider.dart';
import '../../shared/models/models.dart';
import '../../shared/widgets/osm_map_view.dart';

class DeliveryRouteScreen extends StatefulWidget {
  final ApiClient apiClient;
  final DeliveryAssignmentModel assignment;
  final Function(String orderId, String newStatus, String successMsg) onUpdateStatus;

  const DeliveryRouteScreen({
    super.key,
    required this.apiClient,
    required this.assignment,
    required this.onUpdateStatus,
  });

  @override
  State<DeliveryRouteScreen> createState() => _DeliveryRouteScreenState();
}

class _DeliveryRouteScreenState extends State<DeliveryRouteScreen> {
  final GlobalKey<OsmMapViewState> _mapKey = GlobalKey<OsmMapViewState>();
  final DraggableScrollableController _sheetController = DraggableScrollableController();

  RouteModel? _route;
  bool _isLoadingRoute = true;
  String? _routeError;

  double? _agentLat;
  double? _agentLng;

  @override
  void initState() {
    super.initState();
    _fetchAgentLocationAndRoute();
  }

  @override
  void dispose() {
    _sheetController.dispose();
    super.dispose();
  }

  Future<void> _fetchAgentLocationAndRoute() async {
    setState(() {
      _isLoadingRoute = true;
      _routeError = null;
    });

    // 1. Try fetching current GPS position
    try {
      final perm = await Geolocator.checkPermission();
      if (perm == LocationPermission.always || perm == LocationPermission.whileInUse) {
        final pos = await Geolocator.getCurrentPosition(
          desiredAccuracy: LocationAccuracy.high,
          timeLimit: const Duration(seconds: 5),
        );
        _agentLat = pos.latitude;
        _agentLng = pos.longitude;
      }
    } catch (_) {}

    // 2. Determine route origin and destination
    // If agent is picked up, route is agent (or vendor) -> customer
    // If agent is assigned, route is agent -> vendor (or vendor -> customer)
    final fromLat = _agentLat ?? (widget.assignment.vendorLat != 0 ? widget.assignment.vendorLat : 12.9767);
    final fromLng = _agentLng ?? (widget.assignment.vendorLng != 0 ? widget.assignment.vendorLng : 77.5713);
    final toLat = widget.assignment.deliveryLat != 0 ? widget.assignment.deliveryLat : 12.9716;
    final toLng = widget.assignment.deliveryLng != 0 ? widget.assignment.deliveryLng : 77.5946;

    // 3. Fetch OSRM road route from backend
    try {
      final res = await widget.apiClient.get(
        '/mobility/route',
        queryParams: {
          'from_lat': fromLat.toString(),
          'from_lng': fromLng.toString(),
          'to_lat': toLat.toString(),
          'to_lng': toLng.toString(),
        },
      );

      if (res != null && mounted) {
        setState(() {
          _route = RouteModel.fromJson(res);
          _isLoadingRoute = false;
        });
      } else {
        if (mounted) setState(() => _isLoadingRoute = false);
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _routeError = 'Road routing service temporarily unavailable';
          _isLoadingRoute = false;
        });
      }
    }
  }

  void _recenter() {
    final lat = _agentLat ?? (widget.assignment.vendorLat != 0 ? widget.assignment.vendorLat : 12.9767);
    final lng = _agentLng ?? (widget.assignment.vendorLng != 0 ? widget.assignment.vendorLng : 77.5713);
    _mapKey.currentState?.animateToCoordinates(lat, lng, zoom: 14.5);
  }

  void _callCustomer() async {
    final phone = widget.assignment.customerPhone.isNotEmpty
        ? widget.assignment.customerPhone
        : '+919888877777';
    final uri = Uri.parse('tel:$phone');
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri);
    }
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);
    final assignment = widget.assignment;
    final shortId = assignment.orderId.length > 8 ? assignment.orderId.substring(0, 8) : assignment.orderId;
    final isAssigned = assignment.status == 'ASSIGNED';
    final isPickedUp = assignment.status == 'PICKED_UP' || assignment.status == 'OUT_FOR_DELIVERY';

    return Scaffold(
      backgroundColor: AppColors.background,
      body: Stack(
        children: [
          // 1. Full-screen Interactive Map
          Positioned.fill(
            child: OsmMapView(
              key: _mapKey,
              height: double.infinity,
              initialLat: _agentLat ?? (assignment.vendorLat != 0 ? assignment.vendorLat : 12.9767),
              initialLng: _agentLng ?? (assignment.vendorLng != 0 ? assignment.vendorLng : 77.5713),
              initialZoom: 13.5,
              vendorLat: assignment.vendorLat != 0 ? assignment.vendorLat : null,
              vendorLng: assignment.vendorLng != 0 ? assignment.vendorLng : null,
              vendorTitle: assignment.vendorName.isNotEmpty ? assignment.vendorName : 'Pickup Vendor',
              customerLat: assignment.deliveryLat != 0 ? assignment.deliveryLat : null,
              customerLng: assignment.deliveryLng != 0 ? assignment.deliveryLng : null,
              customerTitle: assignment.customerName.isNotEmpty ? assignment.customerName : 'Delivery Location',
              agentLat: _agentLat,
              agentLng: _agentLng,
              agentTitle: 'Your Location',
              route: _route,
            ),
          ),

          // 2. Top Header Bar Overlay
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            child: Container(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: [
                    Colors.black.withValues(alpha: 0.65),
                    Colors.black.withValues(alpha: 0.25),
                    Colors.transparent,
                  ],
                  stops: const [0.0, 0.7, 1.0],
                ),
              ),
              child: SafeArea(
                bottom: false,
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 14),
                  child: Row(
                    children: [
                      // Back Button
                      Material(
                        color: Colors.white,
                        shape: const CircleBorder(),
                        elevation: 3,
                        child: InkWell(
                          customBorder: const CircleBorder(),
                          onTap: () => Navigator.pop(context),
                          child: const Padding(
                            padding: EdgeInsets.all(8),
                            child: Icon(Icons.arrow_back_rounded, size: 20, color: AppColors.textPrimary),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(
                              loc.currentLanguage == 'kn' ? 'ವಿತರಣಾ ಮಾರ್ಗ' : 'Delivery Route',
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 18,
                                fontWeight: FontWeight.w800,
                                shadows: [
                                  Shadow(blurRadius: 4, color: Colors.black54, offset: Offset(0, 1)),
                                ],
                              ),
                            ),
                            Text(
                              'Order #$shortId • ${assignment.status}',
                              style: TextStyle(
                                color: Colors.white.withValues(alpha: 0.9),
                                fontSize: 12,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ],
                        ),
                      ),
                      AppStatusBadge.fromOrderStatus(assignment.status),
                    ],
                  ),
                ),
              ),
            ),
          ),

          // 3. Floating Recenter / Map Controls (Right Side)
          Positioned(
            right: 16,
            bottom: 270,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                FloatingActionButton.small(
                  heroTag: 'recenter_delivery_route_btn',
                  backgroundColor: Colors.white,
                  foregroundColor: AppColors.primary,
                  elevation: 3,
                  onPressed: _recenter,
                  tooltip: 'Recenter',
                  child: const Icon(Icons.my_location_rounded, size: 20),
                ),
              ],
            ),
          ),

          // 4. Draggable Collapsible Delivery Details Bottom Sheet
          DraggableScrollableSheet(
            controller: _sheetController,
            initialChildSize: 0.32,
            minChildSize: 0.12,
            maxChildSize: 0.70,
            builder: (ctx, scrollController) {
              return Container(
                decoration: BoxDecoration(
                  color: AppColors.surface,
                  borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.18),
                      blurRadius: 14,
                      offset: const Offset(0, -3),
                    ),
                  ],
                ),
                child: ListView(
                  controller: scrollController,
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
                  children: [
                    // Drag Handle
                    Center(
                      child: Container(
                        width: 40,
                        height: 4,
                        margin: const EdgeInsets.only(bottom: 12),
                        decoration: BoxDecoration(
                          color: AppColors.border,
                          borderRadius: BorderRadius.circular(2),
                        ),
                      ),
                    ),

                    // Route ETA & Distance Pill (if route loaded)
                    if (_route != null && _route!.routingAvailable) ...[
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        margin: const EdgeInsets.only(bottom: 12),
                        decoration: BoxDecoration(
                          color: AppColors.mintSoft,
                          borderRadius: AppRadii.roundedMd,
                          border: Border.all(color: AppColors.mint),
                        ),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Row(
                              children: [
                                const Icon(Icons.two_wheeler_rounded, size: 18, color: AppColors.primary),
                                const SizedBox(width: 8),
                                Text(
                                  '${_route!.distanceKm.toStringAsFixed(1)} km road distance',
                                  style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: AppColors.primaryDark),
                                ),
                              ],
                            ),
                            Text(
                              '~${_route!.durationMinutes.toStringAsFixed(0)} mins ETA',
                              style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 13, color: AppColors.primary),
                            ),
                          ],
                        ),
                      ),
                    ] else if (_isLoadingRoute) ...[
                      const Padding(
                        padding: EdgeInsets.symmetric(vertical: 4),
                        child: Center(
                          child: SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.primary),
                          ),
                        ),
                      ),
                      const SizedBox(height: 8),
                    ] else if (_routeError != null) ...[
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                        margin: const EdgeInsets.only(bottom: 10),
                        decoration: const BoxDecoration(
                          color: AppColors.warningBg,
                          borderRadius: AppRadii.roundedSm,
                        ),
                        child: Row(
                          children: [
                            const Icon(Icons.info_outline, size: 14, color: AppColors.warning),
                            const SizedBox(width: 6),
                            Expanded(
                              child: Text(
                                _routeError!,
                                style: const TextStyle(fontSize: 11, color: AppColors.warningText, fontWeight: FontWeight.w500),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],

                    // Pickup Location Section
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          width: 28,
                          height: 28,
                          decoration: const BoxDecoration(color: AppColors.mintSoft, shape: BoxShape.circle),
                          child: const Icon(Icons.storefront_rounded, size: 16, color: AppColors.primary),
                        ),
                        const SizedBox(width: 10),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(loc.t('pickup_location'), style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: AppColors.textTertiary)),
                              const SizedBox(height: 2),
                              Text(
                                assignment.vendorName.isNotEmpty ? assignment.vendorName : 'Vendor Mandi Store, Bangalore',
                                style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),

                    // Delivery Location Section
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          width: 28,
                          height: 28,
                          decoration: const BoxDecoration(color: AppColors.infoBg, shape: BoxShape.circle),
                          child: const Icon(Icons.place_rounded, size: 16, color: AppColors.info),
                        ),
                        const SizedBox(width: 10),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(loc.t('delivery_location'), style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: AppColors.textTertiary)),
                              const SizedBox(height: 2),
                              Text(
                                assignment.deliveryAddressLine1.isNotEmpty ? assignment.deliveryAddressLine1 : 'Customer Delivery Address, Bangalore',
                                style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
                              ),
                              if (assignment.customerName.isNotEmpty) ...[
                                const SizedBox(height: 2),
                                Text(
                                  'Recipient: ${assignment.customerName}',
                                  style: const TextStyle(fontSize: 12, color: AppColors.textSecondary),
                                ),
                              ],
                            ],
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 18),

                    // Contact & Delivery Actions
                    Row(
                      children: [
                        Expanded(
                          child: AppSecondaryButton(
                            label: 'Call Customer',
                            icon: Icons.phone_rounded,
                            onPressed: _callCustomer,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 10),

                    if (isAssigned) ...[
                      AppPrimaryButton(
                        label: loc.t('confirm_pickup'),
                        icon: Icons.inventory_rounded,
                        onPressed: () {
                          widget.onUpdateStatus(assignment.orderId, 'PICKED_UP', 'Pickup confirmed! Starting transit to customer.');
                          Navigator.pop(context);
                        },
                      ),
                    ] else if (isPickedUp) ...[
                      AppPrimaryButton(
                        label: loc.t('mark_delivered'),
                        icon: Icons.check_circle_rounded,
                        onPressed: () {
                          widget.onUpdateStatus(assignment.orderId, 'DELIVERED', 'Order safely delivered!');
                          Navigator.pop(context);
                        },
                      ),
                    ],
                  ],
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}

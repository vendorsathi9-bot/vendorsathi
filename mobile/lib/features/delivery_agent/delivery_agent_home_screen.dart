import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../core/api/api_client.dart';
import '../../core/localization/locale_provider.dart';
import '../../shared/models/models.dart';
import '../../shared/widgets/osm_map_view.dart';
import 'location_update_service.dart';

class DeliveryAgentHomeScreen extends StatefulWidget {
  final ApiClient apiClient;
  final VoidCallback onLogout;
  final VoidCallback? onSwitchExperience;
  final bool canSwitchExperience;

  const DeliveryAgentHomeScreen({
    super.key,
    required this.apiClient,
    required this.onLogout,
    this.onSwitchExperience,
    this.canSwitchExperience = false,
  });

  @override
  State<DeliveryAgentHomeScreen> createState() => _DeliveryAgentHomeScreenState();
}

class _DeliveryAgentHomeScreenState extends State<DeliveryAgentHomeScreen> {
  late final LocationUpdateService _locationService;
  List<DeliveryAssignmentModel> _assignments = [];
  bool _isLoading = true;
  String? _errorMessage;
  int _selectedTabIndex = 0; // 0: Active, 1: History

  @override
  void initState() {
    super.initState();
    _locationService = LocationUpdateService(apiClient: widget.apiClient);
    _locationService.onLocationUpdated = (pos, time) {
      if (mounted) setState(() {});
    };
    _locationService.onError = (err) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(err), backgroundColor: Colors.red),
        );
      }
    };
    _fetchAssignments();
  }

  @override
  void dispose() {
    _locationService.dispose();
    super.dispose();
  }

  Future<void> _fetchAssignments() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final res = await widget.apiClient.get('/delivery/agent/assignments');
      if (res != null && res['data'] is List) {
        final list = (res['data'] as List)
            .map((item) => DeliveryAssignmentModel.fromJson(item))
            .toList();
        setState(() {
          _assignments = list;
          _isLoading = false;
        });
      } else {
        setState(() {
          _assignments = [];
          _isLoading = false;
        });
      }
    } catch (e) {
      setState(() {
        _errorMessage = e.toString();
        _isLoading = false;
      });
    }
  }

  Future<void> _updateStatus(String orderId, String newStatus) async {
    try {
      await widget.apiClient.patch(
        '/delivery/orders/$orderId/status',
        body: {'status': newStatus},
      );

      if (newStatus == 'DELIVERED') {
        if (_locationService.activeOrderId == orderId) {
          _locationService.stopBroadcasting();
        }
      }

      await _fetchAssignments();

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Delivery updated to $newStatus'), backgroundColor: Colors.green),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to update status: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _openNavigation(DeliveryAssignmentModel assignment) async {
    final destLat = assignment.deliveryLat;
    final destLng = assignment.deliveryLng;
    final originLat = assignment.vendorLat;
    final originLng = assignment.vendorLng;

    String url;
    if (destLat != 0 && destLng != 0) {
      if (originLat != 0 && originLng != 0) {
        url = 'https://www.openstreetmap.org/directions?engine=fossgis_osrm_car&route=$originLat%2C$originLng%3B$destLat%2C$destLng';
      } else {
        url = 'https://www.openstreetmap.org/?mlat=$destLat&mlon=$destLng#map=16/$destLat/$destLng';
      }
    } else {
      final address = Uri.encodeComponent(
          '${assignment.deliveryAddressLine1}, ${assignment.deliveryCity}');
      url = 'https://nominatim.openstreetmap.org/search?q=$address';
    }

    final uri = Uri.parse(url);
    try {
      final canLaunch = await canLaunchUrl(uri);
      if (canLaunch) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } else {
        await launchUrl(uri, mode: LaunchMode.platformDefault);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Could not open map: $e')),
        );
      }
    }
  }

  Future<void> _showRoutePreview(DeliveryAssignmentModel assignment) async {
    RouteModel? routeInfo;
    try {
      final vendorLat = assignment.vendorLat != 0 ? assignment.vendorLat : 12.9778;
      final vendorLng = assignment.vendorLng != 0 ? assignment.vendorLng : 77.5727;
      final custLat = assignment.deliveryLat != 0 ? assignment.deliveryLat : 12.9352;
      final custLng = assignment.deliveryLng != 0 ? assignment.deliveryLng : 77.6245;

      final res = await widget.apiClient.get('/mobility/route', queryParams: {
        'origin_lat': vendorLat.toString(),
        'origin_lng': vendorLng.toString(),
        'dest_lat': custLat.toString(),
        'dest_lng': custLng.toString(),
      });
      if (res != null) {
        routeInfo = RouteModel.fromJson(res);
      }
    } catch (_) {}

    if (!mounted) return;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
      builder: (ctx) {
        return Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      const Icon(Icons.map_rounded, color: Color(0xFF2563EB)),
                      const SizedBox(width: 8),
                      Text(
                        'Route Preview: Order #${assignment.orderId.substring(0, 8)}',
                        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                      ),
                    ],
                  ),
                  IconButton(
                    icon: const Icon(Icons.close),
                    onPressed: () => Navigator.pop(ctx),
                  ),
                ],
              ),
              const Divider(),
              OsmMapView(
                vendorLat: assignment.vendorLat != 0 ? assignment.vendorLat : 12.9778,
                vendorLng: assignment.vendorLng != 0 ? assignment.vendorLng : 77.5727,
                customerLat: assignment.deliveryLat != 0 ? assignment.deliveryLat : 12.9352,
                customerLng: assignment.deliveryLng != 0 ? assignment.deliveryLng : 77.6245,
                route: routeInfo,
                height: 240,
              ),
              const SizedBox(height: 12),
              if (routeInfo != null) ...[
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      routeInfo.routingAvailable
                          ? 'Road Distance: ${routeInfo.distanceKm.toStringAsFixed(1)} km'
                          : (routeInfo.approximateStraightLineKm != null
                              ? 'Approx. Straight-Line: ${routeInfo.approximateStraightLineKm!.toStringAsFixed(1)} km'
                              : 'Distance: Unavailable'),
                      style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                    ),
                    Text(
                      routeInfo.routingAvailable
                          ? 'Est. Duration: ${routeInfo.durationMinutes.toStringAsFixed(0)} mins'
                          : 'ETA: Unavailable',
                      style: TextStyle(
                        fontWeight: FontWeight.w600,
                        fontSize: 13,
                        color: routeInfo.routingAvailable ? const Color(0xFF2563EB) : const Color(0xFFD97706),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
              ],
              SizedBox(
                width: double.infinity,
                child: ElevatedButton.icon(
                  onPressed: () {
                    Navigator.pop(ctx);
                    _openNavigation(assignment);
                  },
                  icon: const Icon(Icons.navigation, size: 16),
                  label: const Text('Start Navigation in OpenStreetMap'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF2563EB),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Future<void> _toggleLocationBroadcasting(String orderId) async {
    if (_locationService.isBroadcasting && _locationService.activeOrderId == orderId) {
      _locationService.stopBroadcasting();
      setState(() {});
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Location broadcast stopped')),
      );
    } else {
      final success = await _locationService.startBroadcasting(orderId);
      setState(() {});
      if (success && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Broadcasting live GPS to customer & vendor'), backgroundColor: Colors.teal),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final t = LocaleProvider.of(context);
    final activeAssignments =
        _assignments.where((a) => a.status != 'DELIVERED').toList();
    final completedAssignments =
        _assignments.where((a) => a.status == 'DELIVERED').toList();

    final displayedList =
        _selectedTabIndex == 0 ? activeAssignments : completedAssignments;

    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            const Icon(Icons.delivery_dining, size: 28),
            const SizedBox(width: 8),
            Text(t.t('delivery_portal')),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.language),
            tooltip: t.t('settings_language'),
            onPressed: () {
              final newLang = t.currentLanguage == 'en' ? 'kn' : 'en';
              t.setLanguage(newLang);
            },
          ),
          if (widget.canSwitchExperience && widget.onSwitchExperience != null)
            IconButton(
              icon: const Icon(Icons.swap_horiz),
              tooltip: 'Switch Portal',
              onPressed: widget.onSwitchExperience,
            ),
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _fetchAssignments,
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            tooltip: t.t('logout'),
            onPressed: widget.onLogout,
          ),
        ],
      ),
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _selectedTabIndex,
        onTap: (index) => setState(() => _selectedTabIndex = index),
        items: [
          BottomNavigationBarItem(
            icon: const Icon(Icons.two_wheeler),
            label: 'Active (${activeAssignments.length})',
          ),
          BottomNavigationBarItem(
            icon: const Icon(Icons.history),
            label: 'History (${completedAssignments.length})',
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _errorMessage != null
              ? Center(
                  child: Padding(
                    padding: const EdgeInsets.all(24),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(Icons.error_outline, size: 48, color: Colors.red),
                        const SizedBox(height: 12),
                        Text(_errorMessage!, textAlign: TextAlign.center),
                        const SizedBox(height: 16),
                        ElevatedButton(
                          onPressed: _fetchAssignments,
                          child: Text(t.t('refresh')),
                        )
                      ],
                    ),
                  ),
                )
              : displayedList.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            _selectedTabIndex == 0
                                ? Icons.task_alt
                                : Icons.history_toggle_off,
                            size: 64,
                            color: Colors.grey[400],
                          ),
                          const SizedBox(height: 16),
                          Text(
                            _selectedTabIndex == 0
                                ? 'No pending deliveries assigned'
                                : 'No completed deliveries yet',
                            style: TextStyle(
                              fontSize: 16,
                              color: Colors.grey[600],
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            _selectedTabIndex == 0
                                ? 'New assignments will appear here automatically'
                                : 'Past fulfilled orders will be listed here',
                            style: TextStyle(fontSize: 13, color: Colors.grey[400]),
                          ),
                        ],
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: _fetchAssignments,
                      child: ListView.builder(
                        padding: const EdgeInsets.all(12),
                        itemCount: displayedList.length,
                        itemBuilder: (context, index) {
                          return _buildAssignmentCard(displayedList[index], t);
                        },
                      ),
                    ),
    );
  }

  Widget _buildAssignmentCard(DeliveryAssignmentModel assignment, LocaleProvider t) {
    final isBroadcastingThis =
        _locationService.isBroadcasting && _locationService.activeOrderId == assignment.orderId;

    return Card(
      margin: const EdgeInsets.only(bottom: 14),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header: Order ID & Status
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: Colors.blue.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: const Icon(Icons.receipt_long, color: Colors.blue, size: 20),
                    ),
                    const SizedBox(width: 10),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Order #${assignment.orderId.length > 8 ? assignment.orderId.substring(0, 8) : assignment.orderId}',
                          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
                        ),
                        Text(
                          '₹${(assignment.orderTotalPaise / 100).toStringAsFixed(2)}',
                          style: const TextStyle(
                            color: Colors.green,
                            fontWeight: FontWeight.bold,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                _buildStatusChip(assignment.status),
              ],
            ),
            const Divider(height: 24),

            // Pickup Info (Vendor)
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(Icons.store, color: Colors.deepOrange, size: 20),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Pickup: ${assignment.vendorName.isNotEmpty ? assignment.vendorName : "Vendor"}',
                        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
                      ),
                      if (assignment.vendorPhone.isNotEmpty)
                        Text(
                          assignment.vendorPhone,
                          style: TextStyle(color: Colors.grey[600], fontSize: 13),
                        ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),

            // Delivery Info (Customer)
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(Icons.location_on, color: Colors.red, size: 20),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Deliver to: ${assignment.customerName.isNotEmpty ? assignment.customerName : "Customer"}',
                        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
                      ),
                      if (assignment.customerPhone.isNotEmpty)
                        Text(
                          assignment.customerPhone,
                          style: TextStyle(color: Colors.grey[600], fontSize: 13),
                        ),
                      if (assignment.deliveryAddressLine1.isNotEmpty)
                        Text(
                          '${assignment.deliveryAddressLine1}, ${assignment.deliveryCity}',
                          style: TextStyle(color: Colors.grey[700], fontSize: 13),
                        ),
                    ],
                  ),
                ),
              ],
            ),

            // Live broadcasting pulse banner if active
            if (isBroadcastingThis) ...[
              const SizedBox(height: 14),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: Colors.teal.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.teal.withOpacity(0.3)),
                ),
                child: Row(
                  children: [
                    const SizedBox(
                      width: 12,
                      height: 12,
                      child: CircularProgressIndicator(strokeWidth: 2, color: Colors.teal),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        _locationService.lastPosition != null
                            ? 'Sharing GPS: ${_locationService.lastPosition!.latitude.toStringAsFixed(4)}, ${_locationService.lastPosition!.longitude.toStringAsFixed(4)}'
                            : t.t('sharing_location'),
                        style: const TextStyle(
                          color: Colors.teal,
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],

            const SizedBox(height: 16),

            // Action Buttons
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                // Route Preview Map
                OutlinedButton.icon(
                  onPressed: () => _showRoutePreview(assignment),
                  icon: const Icon(Icons.map_outlined, size: 18),
                  label: const Text('Route Preview'),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: const Color(0xFF1E40AF),
                    side: const BorderSide(color: Color(0xFF93C5FD)),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                  ),
                ),

                // Navigation / Directions
                OutlinedButton.icon(
                  onPressed: () => _openNavigation(assignment),
                  icon: const Icon(Icons.directions, size: 18),
                  label: Text(t.t('open_navigation')),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: Colors.blue[700],
                    side: BorderSide(color: Colors.blue[300]!),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                  ),
                ),

                // GPS Live Broadcast toggle (for active orders)
                if (assignment.status != 'DELIVERED')
                  OutlinedButton.icon(
                    onPressed: () => _toggleLocationBroadcasting(assignment.orderId),
                    icon: Icon(
                      isBroadcastingThis ? Icons.gps_off : Icons.gps_fixed,
                      size: 18,
                      color: isBroadcastingThis ? Colors.red : Colors.teal,
                    ),
                    label: Text(
                      isBroadcastingThis ? t.t('stop_sharing') : t.t('share_location'),
                      style: TextStyle(
                        color: isBroadcastingThis ? Colors.red : Colors.teal,
                      ),
                    ),
                    style: OutlinedButton.styleFrom(
                      side: BorderSide(
                        color: isBroadcastingThis ? Colors.red : Colors.teal,
                      ),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    ),
                  ),

                // Progression Action button
                if (assignment.status == 'ASSIGNED')
                  ElevatedButton.icon(
                    onPressed: () => _updateStatus(assignment.orderId, 'ACCEPTED'),
                    icon: const Icon(Icons.check, size: 18),
                    label: Text(t.t('accept_delivery')),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.blue,
                      foregroundColor: Colors.white,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    ),
                  )
                else if (assignment.status == 'ACCEPTED')
                  ElevatedButton.icon(
                    onPressed: () => _updateStatus(assignment.orderId, 'PICKED_UP'),
                    icon: const Icon(Icons.inventory, size: 18),
                    label: Text(t.t('mark_picked_up')),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.indigo,
                      foregroundColor: Colors.white,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    ),
                  )
                else if (assignment.status == 'PICKED_UP')
                  ElevatedButton.icon(
                    onPressed: () => _updateStatus(assignment.orderId, 'OUT_FOR_DELIVERY'),
                    icon: const Icon(Icons.delivery_dining, size: 18),
                    label: Text(t.t('mark_out_for_delivery')),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.amber[800],
                      foregroundColor: Colors.white,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    ),
                  )
                else if (assignment.status == 'OUT_FOR_DELIVERY')
                  ElevatedButton.icon(
                    onPressed: () => _updateStatus(assignment.orderId, 'DELIVERED'),
                    icon: const Icon(Icons.task_alt, size: 18),
                    label: Text(t.t('mark_delivered')),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.green,
                      foregroundColor: Colors.white,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    ),
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatusChip(String status) {
    Color bg;
    Color fg;
    String label;

    switch (status) {
      case 'DELIVERED':
        bg = Colors.green.shade100;
        fg = Colors.green.shade800;
        label = 'Delivered';
        break;
      case 'OUT_FOR_DELIVERY':
        bg = Colors.amber.shade100;
        fg = Colors.amber.shade900;
        label = 'Out for Delivery';
        break;
      case 'PICKED_UP':
        bg = Colors.indigo.shade100;
        fg = Colors.indigo.shade800;
        label = 'Picked Up';
        break;
      case 'ACCEPTED':
        bg = Colors.purple.shade100;
        fg = Colors.purple.shade800;
        label = 'Accepted';
        break;
      default:
        bg = Colors.blueGrey.shade100;
        fg = Colors.blueGrey.shade800;
        label = 'Assigned';
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Text(
        label,
        style: TextStyle(color: fg, fontSize: 12, fontWeight: FontWeight.bold),
      ),
    );
  }
}

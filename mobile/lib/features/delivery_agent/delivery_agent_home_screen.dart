import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../core/api/api_client.dart';
import '../../core/design_system/app_colors.dart';
import '../../core/design_system/app_decorations.dart';
import '../../core/design_system/components/app_badges.dart';
import '../../core/design_system/components/app_buttons.dart';
import '../../core/design_system/components/app_cards.dart';
import '../../core/design_system/components/app_sheets.dart';
import '../../core/design_system/components/app_states.dart';
import '../../core/localization/locale_provider.dart';
import '../../shared/models/models.dart';
import 'delivery_route_screen.dart';
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
  int _selectedTabIndex = 0; // 0: Active Deliveries, 1: History

  @override
  void initState() {
    super.initState();
    _locationService = LocationUpdateService(apiClient: widget.apiClient);
    _locationService.onLocationUpdated = (pos, time) {
      if (mounted) setState(() {});
    };
    _locationService.onError = (err) {
      if (mounted) AppFeedback.showError(context, err);
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
      final raw = res is List ? res : (res != null ? (res['data'] as List?) : null);
      if (raw != null && mounted) {
        setState(() {
          _assignments = raw.map((e) => DeliveryAssignmentModel.fromJson(e)).toList();
          _isLoading = false;
        });

        // If an active in-transit assignment exists, start broadcasting GPS
        DeliveryAssignmentModel? inTransit;
        try {
          inTransit = _assignments.firstWhere(
            (a) => a.status == 'PICKED_UP' || a.status == 'OUT_FOR_DELIVERY',
          );
        } catch (_) {}
        if (inTransit != null && !_locationService.isBroadcasting) {
          _locationService.startBroadcasting(inTransit.orderId);
        }
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = e.toString().replaceAll('ApiException', '').trim();
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _updateStatus(String orderId, String newStatus, String successMsg) async {
    try {
      await widget.apiClient.patch('/delivery/orders/$orderId/status', body: {'status': newStatus});

      if (newStatus == 'PICKED_UP') {
        _locationService.startBroadcasting(orderId);
      } else if (newStatus == 'DELIVERED') {
        _locationService.stopBroadcasting();
      }

      if (mounted) {
        AppFeedback.showSuccess(context, successMsg);
        _fetchAssignments();
      }
    } catch (e) {
      if (mounted) AppFeedback.showError(context, 'Update failed: $e');
    }
  }

  void _launchPhoneCall(String phone) async {
    final uri = Uri.parse('tel:$phone');
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri);
    }
  }

  void _openRouteScreen(DeliveryAssignmentModel job) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (ctx) => DeliveryRouteScreen(
          apiClient: widget.apiClient,
          assignment: job,
          onUpdateStatus: (orderId, newStatus, msg) => _updateStatus(orderId, newStatus, msg),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);

    final activeJobs = _assignments.where((a) => a.status != 'DELIVERED' && a.status != 'CANCELLED').toList();
    final pastJobs = _assignments.where((a) => a.status == 'DELIVERED' || a.status == 'CANCELLED').toList();
    final displayJobs = _selectedTabIndex == 0 ? activeJobs : pastJobs;

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              loc.t('delivery_agent_portal'),
              style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 17, color: AppColors.textPrimary),
            ),
            Row(
              children: [
                Container(
                  width: 7,
                  height: 7,
                  decoration: BoxDecoration(
                    color: _locationService.isBroadcasting ? AppColors.success : AppColors.textTertiary,
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(width: 5),
                Text(
                  _locationService.isBroadcasting ? loc.t('gps_broadcasting') : loc.t('gps_paused'),
                  style: TextStyle(
                    fontSize: 11,
                    color: _locationService.isBroadcasting ? AppColors.successText : AppColors.textSecondary,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
          ],
        ),
        actions: [
          IconButton(
            tooltip: loc.t('refresh'),
            icon: const Icon(Icons.refresh_rounded, size: 20, color: AppColors.textPrimary),
            onPressed: _fetchAssignments,
          ),
          PopupMenuButton<String>(
            icon: const Icon(Icons.more_vert_rounded, size: 20, color: AppColors.textPrimary),
            onSelected: (val) {
              if (val == 'switch') {
                widget.onSwitchExperience?.call();
              } else if (val == 'lang') {
                loc.setLanguage(loc.currentLanguage == 'en' ? 'kn' : 'en');
              } else if (val == 'logout') {
                widget.onLogout();
              }
            },
            itemBuilder: (ctx) => [
              if (widget.canSwitchExperience)
                const PopupMenuItem(
                  value: 'switch',
                  child: Row(
                    children: [
                      Icon(Icons.swap_horiz_rounded, size: 18, color: AppColors.primary),
                      SizedBox(width: 8),
                      Text('Switch Role', style: TextStyle(fontSize: 13)),
                    ],
                  ),
                ),
              PopupMenuItem(
                value: 'lang',
                child: Row(
                  children: [
                    const Icon(Icons.translate_rounded, size: 18, color: AppColors.textSecondary),
                    const SizedBox(width: 8),
                    Text(
                      loc.currentLanguage == 'en' ? 'ಕನ್ನಡಕ್ಕೆ ಬದಲಾಯಿಸಿ' : 'Switch to English',
                      style: const TextStyle(fontSize: 13),
                    ),
                  ],
                ),
              ),
              const PopupMenuDivider(),
              PopupMenuItem(
                value: 'logout',
                child: Row(
                  children: [
                    const Icon(Icons.logout_rounded, size: 18, color: AppColors.error),
                    const SizedBox(width: 8),
                    Text(loc.t('logout'), style: const TextStyle(fontSize: 13, color: AppColors.error)),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _fetchAssignments,
        color: AppColors.primary,
        child: Column(
          children: [
            // Outdoor High-Contrast Tab Selector
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
              child: Row(
                children: [
                  Expanded(
                    child: AppFilterChip(
                      label: '${loc.t('active_deliveries')} (${activeJobs.length})',
                      icon: Icons.two_wheeler_rounded,
                      isSelected: _selectedTabIndex == 0,
                      onTap: () => setState(() => _selectedTabIndex = 0),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: AppFilterChip(
                      label: '${loc.t('delivery_history')} (${pastJobs.length})',
                      icon: Icons.history_rounded,
                      isSelected: _selectedTabIndex == 1,
                      onTap: () => setState(() => _selectedTabIndex = 1),
                    ),
                  ),
                ],
              ),
            ),

            // Deliveries List
            Expanded(
              child: _isLoading
                  ? ListView.builder(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      itemCount: 3,
                      itemBuilder: (_, __) => const AppCardSkeleton(),
                    )
                  : _errorMessage != null
                      ? AppErrorState(message: _errorMessage!, onRetry: _fetchAssignments)
                      : displayJobs.isEmpty
                          ? AppEmptyState(
                              icon: Icons.two_wheeler_outlined,
                              title: _selectedTabIndex == 0 ? 'No active deliveries assigned' : 'No past deliveries',
                              subtitle: _selectedTabIndex == 0
                                  ? 'New orders assigned to you will appear here'
                                  : 'Completed deliveries will appear here',
                              actionLabel: loc.t('refresh'),
                              onAction: _fetchAssignments,
                            )
                          : ListView.builder(
                              padding: const EdgeInsets.fromLTRB(16, 6, 16, 32),
                              itemCount: displayJobs.length,
                              itemBuilder: (ctx, index) {
                                final job = displayJobs[index];
                                final shortId = job.orderId.length > 8 ? job.orderId.substring(0, 8) : job.orderId;
                                final isAssigned = job.status == 'ASSIGNED';
                                final isPickedUp = job.status == 'PICKED_UP' || job.status == 'OUT_FOR_DELIVERY';
                                final isDelivered = job.status == 'DELIVERED';

                                return AppCard(
                                  padding: const EdgeInsets.all(16),
                                  margin: const EdgeInsets.only(bottom: 14),
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      // Header
                                      Row(
                                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                        children: [
                                          Text(
                                            'Delivery Order #$shortId',
                                            style: const TextStyle(fontSize: 15, fontWeight: FontWeight.bold, color: AppColors.textPrimary),
                                          ),
                                          AppStatusBadge.fromOrderStatus(job.status),
                                        ],
                                      ),
                                      const Divider(height: 18),

                                      // Pickup Info
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
                                                  job.pickupAddress.isNotEmpty ? job.pickupAddress : 'Vendor Mandi Store, Bangalore',
                                                  style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
                                                ),
                                              ],
                                            ),
                                          ),
                                        ],
                                      ),
                                      const SizedBox(height: 12),

                                      // Dropoff Info
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
                                                  job.deliveryAddress.isNotEmpty ? job.deliveryAddress : 'Customer Address, Bangalore',
                                                  style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
                                                ),
                                                if (job.customerName.isNotEmpty) ...[
                                                  const SizedBox(height: 2),
                                                  Text(
                                                    'Recipient: ${job.customerName}',
                                                    style: const TextStyle(fontSize: 12, color: AppColors.textSecondary),
                                                  ),
                                                ],
                                              ],
                                            ),
                                          ),
                                        ],
                                      ),
                                      const SizedBox(height: 16),

                                      // Large Outdoor Action Buttons
                                      if (isAssigned) ...[
                                        Row(
                                          children: [
                                            Expanded(
                                              child: AppSecondaryButton(
                                                label: 'View Route',
                                                icon: Icons.navigation_rounded,
                                                onPressed: () => _openRouteScreen(job),
                                              ),
                                            ),
                                            const SizedBox(width: 8),
                                            Expanded(
                                              child: AppPrimaryButton(
                                                label: loc.t('confirm_pickup'),
                                                icon: Icons.inventory_rounded,
                                                onPressed: () => _updateStatus(job.orderId, 'PICKED_UP', 'Pickup confirmed! Heading to customer.'),
                                              ),
                                            ),
                                          ],
                                        ),
                                      ] else if (isPickedUp) ...[
                                        Row(
                                          children: [
                                            Expanded(
                                              child: AppSecondaryButton(
                                                label: 'Call Customer',
                                                icon: Icons.phone_rounded,
                                                onPressed: () => _launchPhoneCall(job.customerPhone.isNotEmpty ? job.customerPhone : '+919888877777'),
                                              ),
                                            ),
                                            const SizedBox(width: 8),
                                            Expanded(
                                              child: AppSecondaryButton(
                                                label: 'View Route',
                                                icon: Icons.navigation_rounded,
                                                onPressed: () => _openRouteScreen(job),
                                              ),
                                            ),
                                          ],
                                        ),
                                        const SizedBox(height: 10),
                                        AppPrimaryButton(
                                          label: loc.t('mark_delivered'),
                                          icon: Icons.check_circle_rounded,
                                          onPressed: () => _updateStatus(job.orderId, 'DELIVERED', 'Order safely delivered!'),
                                        ),
                                      ] else if (isDelivered) ...[
                                        Container(
                                          padding: const EdgeInsets.all(8),
                                          decoration: const BoxDecoration(
                                            color: AppColors.successBg,
                                            borderRadius: AppRadii.roundedSm,
                                          ),
                                          child: const Row(
                                            mainAxisAlignment: MainAxisAlignment.center,
                                            children: [
                                              Icon(Icons.check_circle_rounded, size: 16, color: AppColors.success),
                                              SizedBox(width: 6),
                                              Text('Delivered Successfully', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: AppColors.successText)),
                                            ],
                                          ),
                                        ),
                                      ],
                                    ],
                                  ),
                                );
                              },
                            ),
            ),
          ],
        ),
      ),
    );
  }
}

import 'dart:async';
import 'package:flutter/material.dart';
import '../../../core/api/api_client.dart';
import '../../../core/design_system/app_colors.dart';
import '../../../core/design_system/app_decorations.dart';
import '../../../core/design_system/components/app_badges.dart';
import '../../../core/design_system/components/app_buttons.dart';
import '../../../core/design_system/components/app_cards.dart';
import '../../../core/design_system/components/app_sheets.dart';
import '../../../core/design_system/components/app_states.dart';
import '../../../core/localization/locale_provider.dart';
import '../../../shared/models/models.dart';
import '../../../shared/widgets/osm_map_view.dart';

class CustomerOrdersTab extends StatefulWidget {
  final ApiClient apiClient;
  final List<OrderModel> orders;
  final bool isLoading;
  final VoidCallback onRefresh;

  const CustomerOrdersTab({
    super.key,
    required this.apiClient,
    required this.orders,
    required this.isLoading,
    required this.onRefresh,
  });

  @override
  State<CustomerOrdersTab> createState() => _CustomerOrdersTabState();
}

class _CustomerOrdersTabState extends State<CustomerOrdersTab> {
  int _selectedFilter = 0; // 0: Active, 1: Past

  void _showTrackingModal(OrderModel order) {
    AppBottomSheet.show(
      context: context,
      title: 'Delivery Tracking #${order.id.length > 8 ? order.id.substring(0, 8) : order.id}',
      child: _OrderTrackingView(apiClient: widget.apiClient, order: order),
    );
  }

  void _showPickupInfoModal(OrderModel order) {
    final shortId = order.id.length > 8 ? order.id.substring(0, 8) : order.id;
    AppBottomSheet.show(
      context: context,
      title: 'Store Pickup Info #$shortId',
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: AppColors.mintSoft,
                borderRadius: AppRadii.roundedMd,
                border: Border.all(color: AppColors.mint),
              ),
              child: const Row(
                children: [
                  Icon(Icons.storefront_rounded, color: AppColors.primary, size: 28),
                  SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Ready for Counter Pickup',
                          style: TextStyle(fontWeight: FontWeight.bold, fontSize: 15, color: AppColors.primaryDark),
                        ),
                        SizedBox(height: 2),
                        Text(
                          'Your package has been prepared and is waiting at the storefront counter.',
                          style: TextStyle(fontSize: 12, color: AppColors.textSecondary),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: AppColors.surface,
                borderRadius: AppRadii.roundedMd,
                border: Border.all(color: AppColors.border),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Pickup Reference', style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: AppColors.textTertiary)),
                  const SizedBox(height: 2),
                  Text('Order #$shortId', style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w800, color: AppColors.textPrimary)),
                  const Divider(height: 16),
                  const Text('Location', style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: AppColors.textTertiary)),
                  const SizedBox(height: 2),
                  const Text('Vendor Store Counter, Bangalore', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.textPrimary)),
                ],
              ),
            ),
            const SizedBox(height: 16),
            AppPrimaryButton(
              label: 'Done',
              icon: Icons.check_circle_outline_rounded,
              onPressed: () => Navigator.pop(context),
            ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
  }

  void _showRatingModal(OrderModel order) {
    int rating = 5;
    final comment = TextEditingController();

    showDialog(
      context: context,
      builder: (dlgCtx) => StatefulBuilder(
        builder: (ctx, setDlgState) => AlertDialog(
          backgroundColor: AppColors.surface,
          shape: const RoundedRectangleBorder(borderRadius: AppRadii.roundedLg),
          title: const Text('Rate Your Experience', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: List.generate(5, (index) {
                  final star = index + 1;
                  return IconButton(
                    icon: Icon(
                      star <= rating ? Icons.star_rounded : Icons.star_outline_rounded,
                      color: Colors.amber,
                      size: 32,
                    ),
                    onPressed: () => setDlgState(() => rating = star),
                  );
                }),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: comment,
                decoration: const InputDecoration(
                  hintText: 'Add a comment (optional)...',
                  border: OutlineInputBorder(),
                ),
                maxLines: 2,
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(dlgCtx),
              child: const Text('Cancel', style: TextStyle(color: AppColors.textSecondary)),
            ),
            ElevatedButton(
              onPressed: () async {
                try {
                  await widget.apiClient.post('/reviews', body: {
                    'order_id': order.id,
                    'vendor_id': order.vendorId,
                    'rating': rating,
                    'comment': comment.text.trim(),
                  });
                  if (mounted) {
                    Navigator.pop(dlgCtx);
                    AppFeedback.showSuccess(context, 'Thank you for your rating!');
                  }
                } catch (e) {
                  // Review error
                }
              },
              style: ElevatedButton.styleFrom(backgroundColor: AppColors.primary),
              child: const Text('Submit'),
            ),
          ],
        ),
      ),
    );
  }

  void _requestCancellation(OrderModel order) {
    showDialog(
      context: context,
      builder: (dlgCtx) => AlertDialog(
        backgroundColor: AppColors.surface,
        shape: const RoundedRectangleBorder(borderRadius: AppRadii.roundedLg),
        title: const Text('Cancel Order?', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
        content: const Text('Are you sure you want to request cancellation for this order?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dlgCtx),
            child: const Text('No, Keep Order'),
          ),
          ElevatedButton(
            onPressed: () async {
              try {
                await widget.apiClient.post('/orders/${order.id}/request-cancellation', body: {
                  'reason': 'Requested by customer',
                });
                if (mounted) {
                  Navigator.pop(dlgCtx);
                  AppFeedback.showSuccess(context, 'Cancellation requested');
                  widget.onRefresh();
                }
              } catch (e) {
                // Cancel error
              }
            },
            style: ElevatedButton.styleFrom(backgroundColor: AppColors.error),
            child: const Text('Request Cancel'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);

    final activeOrders = widget.orders.where((o) {
      final s = o.status.toUpperCase();
      return s != 'COMPLETED' && s != 'CANCELLED' && s != 'REJECTED';
    }).toList();

    final pastOrders = widget.orders.where((o) {
      final s = o.status.toUpperCase();
      return s == 'COMPLETED' || s == 'CANCELLED' || s == 'REJECTED';
    }).toList();

    final displayOrders = _selectedFilter == 0 ? activeOrders : pastOrders;

    return RefreshIndicator(
      onRefresh: () async => widget.onRefresh(),
      color: AppColors.primary,
      child: Column(
        children: [
          // Filter Tabs (Active vs Completed)
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
            child: Row(
              children: [
                Expanded(
                  child: AppFilterChip(
                    label: '${loc.t('active')} (${activeOrders.length})',
                    icon: Icons.timelapse_rounded,
                    isSelected: _selectedFilter == 0,
                    onTap: () => setState(() => _selectedFilter = 0),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: AppFilterChip(
                    label: '${loc.t('completed')} (${pastOrders.length})',
                    icon: Icons.history_rounded,
                    isSelected: _selectedFilter == 1,
                    onTap: () => setState(() => _selectedFilter = 1),
                  ),
                ),
              ],
            ),
          ),

          // Orders List
          Expanded(
            child: widget.isLoading
                ? ListView.builder(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    itemCount: 3,
                    itemBuilder: (_, __) => const AppCardSkeleton(),
                  )
                : displayOrders.isEmpty
                    ? AppEmptyState(
                        icon: Icons.receipt_long_outlined,
                        title: loc.t('empty_orders_title'),
                        subtitle: loc.t('empty_orders_subtitle'),
                        actionLabel: loc.t('refresh'),
                        onAction: widget.onRefresh,
                      )
                    : ListView.builder(
                        padding: const EdgeInsets.fromLTRB(16, 4, 16, 24),
                        itemCount: displayOrders.length,
                        itemBuilder: (ctx, index) {
                          final order = displayOrders[index];
                          final totalRupees = (order.totalPaise / 100).toStringAsFixed(2);
                          final shortId = order.id.length > 8 ? order.id.substring(0, 8) : order.id;
                          final isPending = order.status.toUpperCase() == 'PENDING';
                          final isOutForDelivery = order.status.toUpperCase() == 'OUT_FOR_DELIVERY';
                          final isCompleted = order.status.toUpperCase() == 'COMPLETED';

                          return AppCard(
                            padding: const EdgeInsets.all(14),
                            margin: const EdgeInsets.only(bottom: 12),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                // Order ID and status badge
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                  children: [
                                    Text(
                                      'Order #$shortId',
                                      style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: AppColors.textPrimary),
                                    ),
                                    AppStatusBadge.fromOrderStatus(order.status),
                                  ],
                                ),
                                const SizedBox(height: 8),
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                  children: [
                                    Row(
                                      children: [
                                        Icon(
                                          order.fulfilmentType == 'DELIVERY'
                                              ? Icons.delivery_dining_rounded
                                              : Icons.storefront_rounded,
                                          size: 15,
                                          color: AppColors.textSecondary,
                                        ),
                                        const SizedBox(width: 4),
                                        Text(
                                          order.fulfilmentType,
                                          style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.textSecondary),
                                        ),
                                      ],
                                    ),
                                    Text(
                                      '₹$totalRupees',
                                      style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w800, color: AppColors.primary),
                                    ),
                                  ],
                                ),
                                const Divider(height: 18),

                                // Timeline step indicator
                                _OrderStepProgress(
                                  status: order.status,
                                  fulfilmentType: order.fulfilmentType,
                                ),

                                const SizedBox(height: 12),
                                // Actions
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.end,
                                  children: [
                                    if (isPending) ...[
                                      AppOutlinedButton(
                                        label: 'Cancel',
                                        color: AppColors.error,
                                        onPressed: () => _requestCancellation(order),
                                      ),
                                    ] else if (order.fulfilmentType == 'DELIVERY' && isOutForDelivery) ...[
                                      AppSecondaryButton(
                                        label: loc.t('order_tracking'),
                                        icon: Icons.navigation_rounded,
                                        isFullWidth: false,
                                        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                                        onPressed: () => _showTrackingModal(order),
                                      ),
                                    ] else if (order.fulfilmentType == 'PICKUP' && order.status == 'READY_FOR_PICKUP') ...[
                                      AppSecondaryButton(
                                        label: loc.t('status_ready_for_pickup'),
                                        icon: Icons.storefront_rounded,
                                        isFullWidth: false,
                                        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                                        onPressed: () => _showPickupInfoModal(order),
                                      ),
                                    ] else if (isCompleted) ...[
                                      AppSecondaryButton(
                                        label: 'Rate Order',
                                        icon: Icons.star_rate_rounded,
                                        isFullWidth: false,
                                        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                                        onPressed: () => _showRatingModal(order),
                                      ),
                                    ],
                                  ],
                                ),
                              ],
                            ),
                          );
                        },
                      ),
          ),
        ],
      ),
    );
  }
}

/// Step Progress Indicator for Order Lifecycle
class _OrderStepProgress extends StatelessWidget {
  final String status;
  final String fulfilmentType;

  const _OrderStepProgress({
    required this.status,
    required this.fulfilmentType,
  });

  bool get _isPickup => fulfilmentType.toUpperCase() == 'PICKUP';

  List<String> get _steps {
    if (_isPickup) {
      return ['Placed', 'Accepted', 'Preparing', 'Ready', 'Picked Up'];
    }
    return ['Placed', 'Accepted', 'Preparing', 'On Way', 'Delivered'];
  }

  int get _stepIndex {
    switch (status.toUpperCase()) {
      case 'PENDING':
        return 0;
      case 'ACCEPTED':
        return 1;
      case 'PREPARING':
        return 2;
      case 'READY_FOR_PICKUP':
        return _isPickup ? 3 : 2;
      case 'OUT_FOR_DELIVERY':
        return 3;
      case 'COMPLETED':
      case 'DELIVERED':
        return 4;
      default:
        return 0;
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = status.toUpperCase();

    if (s == 'CANCELLATION_REQUESTED') {
      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        decoration: BoxDecoration(
          color: AppColors.warningBg,
          borderRadius: AppRadii.roundedSm,
          border: Border.all(color: AppColors.warning.withValues(alpha: 0.3)),
        ),
        child: const Row(
          children: [
            Icon(Icons.pending_actions_rounded, size: 16, color: AppColors.warning),
            SizedBox(width: 8),
            Expanded(
              child: Text(
                'Cancellation requested • Awaiting vendor review',
                style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.warningText),
              ),
            ),
          ],
        ),
      );
    }

    if (s == 'CANCELLED') {
      return Container(
        padding: const EdgeInsets.all(8),
        decoration: const BoxDecoration(
          color: AppColors.errorBg,
          borderRadius: AppRadii.roundedSm,
        ),
        child: const Row(
          children: [
            Icon(Icons.cancel_outlined, size: 16, color: AppColors.error),
            SizedBox(width: 6),
            Text('This order has been cancelled', style: TextStyle(fontSize: 12, color: AppColors.errorText)),
          ],
        ),
      );
    }

    if (s == 'REJECTED') {
      return Container(
        padding: const EdgeInsets.all(8),
        decoration: const BoxDecoration(
          color: AppColors.errorBg,
          borderRadius: AppRadii.roundedSm,
        ),
        child: const Row(
          children: [
            Icon(Icons.block_rounded, size: 16, color: AppColors.error),
            SizedBox(width: 6),
            Text('This order was rejected by the vendor', style: TextStyle(fontSize: 12, color: AppColors.errorText)),
          ],
        ),
      );
    }

    final steps = _steps;
    final activeIndex = _stepIndex;

    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: List.generate(steps.length, (idx) {
        final isDone = idx <= activeIndex;
        final isCurrent = idx == activeIndex;

        return Expanded(
          child: Column(
            children: [
              Row(
                children: [
                  Expanded(
                    child: Container(
                      height: 3,
                      color: idx == 0
                          ? Colors.transparent
                          : (idx <= activeIndex ? AppColors.primary : AppColors.border),
                    ),
                  ),
                  Container(
                    width: 14,
                    height: 14,
                    decoration: BoxDecoration(
                      color: isDone ? AppColors.primary : AppColors.surface,
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: isDone ? AppColors.primary : AppColors.border,
                        width: 2,
                      ),
                    ),
                    child: isDone
                        ? const Icon(Icons.check, size: 8, color: Colors.white)
                        : null,
                  ),
                  Expanded(
                    child: Container(
                      height: 3,
                      color: idx == steps.length - 1
                          ? Colors.transparent
                          : (idx < activeIndex ? AppColors.primary : AppColors.border),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 4),
              Text(
                steps[idx],
                style: TextStyle(
                  fontSize: 9,
                  fontWeight: isCurrent ? FontWeight.bold : FontWeight.normal,
                  color: isCurrent ? AppColors.primaryDark : AppColors.textTertiary,
                ),
                maxLines: 1,
              ),
            ],
          ),
        );
      }),
    );
  }
}

/// Live Tracking Modal View
class _OrderTrackingView extends StatefulWidget {
  final ApiClient apiClient;
  final OrderModel order;

  const _OrderTrackingView({required this.apiClient, required this.order});

  @override
  State<_OrderTrackingView> createState() => _OrderTrackingViewState();
}

class _OrderTrackingViewState extends State<_OrderTrackingView> {
  Timer? _pollingTimer;
  double? _agentLat;
  double? _agentLng;
  DateTime? _lastPing;

  @override
  void initState() {
    super.initState();
    _fetchLocation();
    _pollingTimer = Timer.periodic(const Duration(seconds: 10), (_) => _fetchLocation());
  }

  @override
  void dispose() {
    _pollingTimer?.cancel();
    super.dispose();
  }

  Future<void> _fetchLocation() async {
    try {
      final res = await widget.apiClient.get('/delivery/orders/${widget.order.id}/location');
      if (res != null && mounted) {
        setState(() {
          _agentLat = (res['latitude'] as num?)?.toDouble();
          _agentLng = (res['longitude'] as num?)?.toDouble();
          _lastPing = DateTime.now();
        });
      }
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: AppColors.mintSoft,
              borderRadius: AppRadii.roundedMd,
              border: Border.all(color: AppColors.mint),
            ),
            child: Row(
              children: [
                const Icon(Icons.near_me_rounded, color: AppColors.primary, size: 20),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    _lastPing != null
                        ? 'Live GPS • Updated just now'
                        : 'Awaiting live GPS coordinates from delivery agent...',
                    style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.primaryDark),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          if (_agentLat != null && _agentLat != 0 && _agentLng != null && _agentLng != 0) ...[
            ClipRRect(
              borderRadius: AppRadii.roundedLg,
              child: SizedBox(
                height: 240,
                child: OsmMapView(
                  apiClient: widget.apiClient,
                  agentLat: _agentLat!,
                  agentLng: _agentLng!,
                  agentTitle: 'Delivery Agent',
                  height: 240,
                ),
              ),
            ),
          ] else ...[
            Container(
              height: 180,
              decoration: BoxDecoration(
                color: AppColors.surfaceMuted,
                borderRadius: AppRadii.roundedLg,
                border: Border.all(color: AppColors.border),
              ),
              child: const Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.two_wheeler_outlined, size: 40, color: AppColors.textTertiary),
                  SizedBox(height: 10),
                  Text(
                    'GPS Tracking Pending',
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: AppColors.textPrimary),
                  ),
                  SizedBox(height: 4),
                  Padding(
                    padding: EdgeInsets.symmetric(horizontal: 24),
                    child: Text(
                      'Live GPS broadcast will start automatically once the delivery partner begins transit to your address.',
                      textAlign: TextAlign.center,
                      style: TextStyle(fontSize: 12, color: AppColors.textSecondary, height: 1.35),
                    ),
                  ),
                ],
              ),
            ),
          ],
          const SizedBox(height: 16),
        ],
      ),
    );
  }
}

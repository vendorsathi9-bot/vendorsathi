import 'package:flutter/material.dart';
import '../../../core/api/api_client.dart';
import '../../../core/design_system/app_colors.dart';
import '../../../core/design_system/components/app_badges.dart';
import '../../../core/design_system/components/app_buttons.dart';
import '../../../core/design_system/components/app_cards.dart';
import '../../../core/design_system/components/app_sheets.dart';
import '../../../core/design_system/components/app_states.dart';
import '../../../core/localization/locale_provider.dart';
import '../../../shared/models/models.dart';

class VendorOrdersTab extends StatefulWidget {
  final ApiClient apiClient;
  final List<OrderModel> orders;
  final bool isLoading;
  final VoidCallback onRefresh;

  const VendorOrdersTab({
    super.key,
    required this.apiClient,
    required this.orders,
    required this.isLoading,
    required this.onRefresh,
  });

  @override
  State<VendorOrdersTab> createState() => _VendorOrdersTabState();
}

class _VendorOrdersTabState extends State<VendorOrdersTab> {
  String _statusFilter = 'ALL';

  Future<void> _updateStatus(String orderId, String actionEndpoint, String successMsg) async {
    try {
      await widget.apiClient.post('/orders/$orderId/$actionEndpoint');
      if (mounted) {
        AppFeedback.showSuccess(context, successMsg);
        widget.onRefresh();
      }
    } catch (e) {
      if (mounted) {
        AppFeedback.showError(context, 'Status update failed: $e');
      }
    }
  }

  Future<void> _decideCancellation(String orderId, String decision) async {
    try {
      await widget.apiClient.post('/orders/$orderId/decide-cancellation', body: {
        'decision': decision,
      });
      if (mounted) {
        AppFeedback.showSuccess(
          context,
          decision == 'APPROVED' ? 'Cancellation approved' : 'Cancellation declined',
        );
        widget.onRefresh();
      }
    } catch (e) {
      if (mounted) {
        AppFeedback.showError(context, 'Decision failed: $e');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);

    final filteredOrders = widget.orders.where((o) {
      if (_statusFilter == 'ALL') return true;
      return o.status.toUpperCase() == _statusFilter;
    }).toList();

    final statusFilters = ['ALL', 'PENDING', 'ACCEPTED', 'PREPARING', 'READY_FOR_PICKUP', 'OUT_FOR_DELIVERY', 'COMPLETED'];

    return RefreshIndicator(
      onRefresh: () async => widget.onRefresh(),
      color: AppColors.primary,
      child: Column(
        children: [
          // Horizontal status filter chips
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
            child: SizedBox(
              height: 36,
              child: ListView.separated(
                scrollDirection: Axis.horizontal,
                itemCount: statusFilters.length,
                separatorBuilder: (_, __) => const SizedBox(width: 8),
                itemBuilder: (ctx, idx) {
                  final status = statusFilters[idx];
                  String label = status == 'ALL' ? loc.t('all') : status.replaceAll('_', ' ');
                  if (status == 'PENDING') label = loc.t('status_pending');
                  if (status == 'ACCEPTED') label = loc.t('status_accepted');
                  if (status == 'PREPARING') label = loc.t('status_preparing');
                  if (status == 'READY_FOR_PICKUP') label = loc.t('status_ready_for_pickup');
                  if (status == 'OUT_FOR_DELIVERY') label = loc.t('status_out_delivery');
                  if (status == 'COMPLETED') label = loc.t('status_completed');

                  return AppFilterChip(
                    label: label,
                    isSelected: _statusFilter == status,
                    onTap: () => setState(() => _statusFilter = status),
                  );
                },
              ),
            ),
          ),

          // Orders List
          Expanded(
            child: widget.isLoading
                ? ListView.builder(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    itemCount: 4,
                    itemBuilder: (_, __) => const AppCardSkeleton(),
                  )
                : filteredOrders.isEmpty
                    ? AppEmptyState(
                        icon: Icons.receipt_long_outlined,
                        title: loc.t('empty_orders_title'),
                        subtitle: loc.t('empty_orders_subtitle'),
                        actionLabel: loc.t('refresh'),
                        onAction: widget.onRefresh,
                      )
                    : ListView.builder(
                        padding: const EdgeInsets.fromLTRB(16, 4, 16, 100),
                        itemCount: filteredOrders.length,
                        itemBuilder: (ctx, index) {
                          final order = filteredOrders[index];
                          final totalRupees = (order.totalPaise / 100).toStringAsFixed(2);
                          final shortId = order.id.length > 8 ? order.id.substring(0, 8) : order.id;
                          final status = order.status.toUpperCase();

                          return AppCard(
                            padding: const EdgeInsets.all(14),
                            margin: const EdgeInsets.only(bottom: 12),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
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
                                const SizedBox(height: 6),
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

                                // State Transition Action Buttons
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.end,
                                  children: [
                                    if (status == 'PENDING') ...[
                                      AppOutlinedButton(
                                        label: 'Reject',
                                        color: AppColors.error,
                                        onPressed: () => _updateStatus(order.id, 'reject', 'Order rejected'),
                                      ),
                                      const SizedBox(width: 8),
                                      AppPrimaryButton(
                                        label: 'Accept Order',
                                        icon: Icons.check_circle_outline_rounded,
                                        isFullWidth: false,
                                        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                                        onPressed: () => _updateStatus(order.id, 'accept', 'Order accepted'),
                                      ),
                                    ] else if (status == 'CANCELLATION_REQUESTED') ...[
                                      AppOutlinedButton(
                                        label: 'Decline',
                                        color: AppColors.textSecondary,
                                        onPressed: () => _decideCancellation(order.id, 'DECLINED'),
                                      ),
                                      const SizedBox(width: 8),
                                      AppPrimaryButton(
                                        label: 'Approve Cancel',
                                        icon: Icons.cancel_outlined,
                                        isFullWidth: false,
                                        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                                        onPressed: () => _decideCancellation(order.id, 'APPROVED'),
                                      ),
                                    ] else if (status == 'ACCEPTED') ...[
                                      AppPrimaryButton(
                                        label: 'Start Preparing',
                                        icon: Icons.soup_kitchen_rounded,
                                        isFullWidth: false,
                                        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                                        onPressed: () => _updateStatus(order.id, 'prepare', 'Preparing order'),
                                      ),
                                    ] else if (status == 'PREPARING') ...[
                                      if (order.fulfilmentType == 'PICKUP')
                                        AppPrimaryButton(
                                          label: 'Ready for Pickup',
                                          icon: Icons.storefront_rounded,
                                          isFullWidth: false,
                                          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                                          onPressed: () => _updateStatus(order.id, 'ready', 'Marked ready for pickup'),
                                        )
                                      else
                                        AppPrimaryButton(
                                          label: 'Dispatch Delivery',
                                          icon: Icons.moped_rounded,
                                          isFullWidth: false,
                                          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                                          onPressed: () => _updateStatus(order.id, 'out-for-delivery', 'Order dispatched for delivery'),
                                        ),
                                    ] else if (status == 'READY_FOR_PICKUP' || status == 'OUT_FOR_DELIVERY') ...[
                                      AppPrimaryButton(
                                        label: 'Complete Order',
                                        icon: Icons.done_all_rounded,
                                        isFullWidth: false,
                                        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                                        onPressed: () => _updateStatus(order.id, 'complete', 'Order marked completed!'),
                                      ),
                                    ] else if (status == 'COMPLETED') ...[
                                      const Row(
                                        children: [
                                          Icon(Icons.check_circle_rounded, color: AppColors.success, size: 16),
                                          SizedBox(width: 4),
                                          Text('Completed & Settled', style: TextStyle(fontSize: 12, color: AppColors.success, fontWeight: FontWeight.bold)),
                                        ],
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

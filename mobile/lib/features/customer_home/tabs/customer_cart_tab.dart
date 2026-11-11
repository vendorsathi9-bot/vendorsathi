import 'package:flutter/material.dart';
import '../../../core/api/api_client.dart';
import '../../../core/design_system/app_colors.dart';
import '../../../core/design_system/app_decorations.dart';
import '../../../core/design_system/components/app_buttons.dart';
import '../../../core/design_system/components/app_cards.dart';
import '../../../core/design_system/components/app_sheets.dart';
import '../../../core/design_system/components/app_states.dart';
import '../../../core/localization/locale_provider.dart';
import '../../../shared/models/models.dart';
import '../widgets/customer_checkout_sheet.dart';

class CustomerCartTab extends StatelessWidget {
  final ApiClient apiClient;
  final CartModel? cart;
  final List<CustomerAddressModel> addresses;
  final LoyaltySummaryModel? loyalty;
  final bool isLoading;
  final Function(String itemId, int newQty) onUpdateItemQty;
  final VoidCallback onClearCart;
  final VoidCallback onRefresh;
  final VoidCallback onOrderSuccess;
  final VoidCallback onExploreVendors;
  final VoidCallback onAddressAdded;

  const CustomerCartTab({
    super.key,
    required this.apiClient,
    this.cart,
    required this.addresses,
    this.loyalty,
    required this.isLoading,
    required this.onUpdateItemQty,
    required this.onClearCart,
    required this.onRefresh,
    required this.onOrderSuccess,
    required this.onExploreVendors,
    required this.onAddressAdded,
  });

  void _openCheckout(BuildContext context) {
    if (cart == null || cart!.items.isEmpty) return;

    AppBottomSheet.show(
      context: context,
      title: LocaleProvider.of(context).t('checkout'),
      child: CustomerCheckoutSheet(
        apiClient: apiClient,
        cart: cart!,
        addresses: addresses,
        loyalty: loyalty,
        onOrderSuccess: onOrderSuccess,
        onAddressAdded: onAddressAdded,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);
    final hasItems = cart != null && cart!.items.isNotEmpty;

    if (isLoading) {
      return Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: List.generate(3, (_) => const AppCardSkeleton()),
        ),
      );
    }

    if (!hasItems) {
      return AppEmptyState(
        icon: Icons.shopping_basket_outlined,
        title: loc.t('empty_cart_title'),
        subtitle: loc.t('empty_cart_subtitle'),
        actionLabel: loc.t('explore_vendors'),
        onAction: onExploreVendors,
      );
    }

    final totalPaise = cart!.totalPaise;
    final totalRupees = (totalPaise / 100).toStringAsFixed(2);

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: RefreshIndicator(
        onRefresh: () async => onRefresh(),
        color: AppColors.primary,
        child: ListView(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 140),
          children: [
            // Vendor banner
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              decoration: BoxDecoration(
                color: AppColors.mintSoft,
                borderRadius: AppRadii.roundedMd,
                border: Border.all(color: AppColors.mint),
              ),
              child: Row(
                children: [
                  const Icon(Icons.storefront_rounded, size: 20, color: AppColors.primary),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '${loc.t('single_vendor_cart')}: ${cart!.vendorName ?? 'Vendor'}',
                      style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: AppColors.primaryDark),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.delete_sweep_outlined, size: 20, color: AppColors.error),
                    tooltip: loc.t('clear_cart'),
                    onPressed: onClearCart,
                    constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                    padding: EdgeInsets.zero,
                  ),
                ],
              ),
            ),
            const SizedBox(height: 14),

            // Cart Items
            ...cart!.items.map((item) {
              final linePrice = ((item.pricePaise * item.quantity) / 100).toStringAsFixed(2);

              return AppCard(
                padding: const EdgeInsets.all(12),
                margin: const EdgeInsets.only(bottom: 10),
                child: Row(
                  children: [
                    Container(
                      width: 44,
                      height: 44,
                      decoration: const BoxDecoration(
                        color: AppColors.surfaceMuted,
                        borderRadius: AppRadii.roundedMd,
                      ),
                      child: const Icon(Icons.shopping_bag_outlined, color: AppColors.primary, size: 22),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            item.productName,
                            style: const TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: AppColors.textPrimary,
                              height: 1.25,
                            ),
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                          ),
                          const SizedBox(height: 2),
                          Text(
                            '₹${(item.pricePaise / 100).toStringAsFixed(2)} × ${item.quantity}',
                            style: const TextStyle(fontSize: 12, color: AppColors.textTertiary),
                          ),
                        ],
                      ),
                    ),
                    Text(
                      '₹$linePrice',
                      style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w700, color: AppColors.textPrimary),
                    ),
                    const SizedBox(width: 12),
                    // Quantity stepper
                    Container(
                      decoration: const BoxDecoration(
                        color: AppColors.mintSoft,
                        borderRadius: AppRadii.roundedPill,
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          IconButton(
                            icon: const Icon(Icons.remove, size: 16, color: AppColors.primary),
                            onPressed: () => onUpdateItemQty(item.id, item.quantity - 1),
                            constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                            padding: EdgeInsets.zero,
                          ),
                          Padding(
                            padding: const EdgeInsets.symmetric(horizontal: 4),
                            child: Text(
                              '${item.quantity}',
                              style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: AppColors.primary),
                            ),
                          ),
                          IconButton(
                            icon: const Icon(Icons.add, size: 16, color: AppColors.primary),
                            onPressed: () => onUpdateItemQty(item.id, item.quantity + 1),
                            constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                            padding: EdgeInsets.zero,
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              );
            }),
            const SizedBox(height: 14),

            // Price Breakdown Card
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: AppColors.surface,
                borderRadius: AppRadii.roundedLg,
                border: Border.all(color: AppColors.border),
                boxShadow: AppShadows.card,
              ),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(loc.t('items_total'), style: const TextStyle(fontSize: 13, color: AppColors.textSecondary)),
                      Text('₹$totalRupees', style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(loc.t('delivery_charge'), style: const TextStyle(fontSize: 13, color: AppColors.textSecondary)),
                      const Text('FREE', style: TextStyle(fontSize: 13, color: AppColors.success, fontWeight: FontWeight.bold)),
                    ],
                  ),
                  const Divider(height: 20),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(loc.t('grand_total'), style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
                      Text(
                        '₹$totalRupees',
                        style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w800, color: AppColors.primary),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
      bottomSheet: Container(
        padding: const EdgeInsets.all(16),
        decoration: const BoxDecoration(
          color: AppColors.surface,
          boxShadow: AppShadows.elevated,
          borderRadius: BorderRadius.only(
            topLeft: Radius.circular(20),
            topRight: Radius.circular(20),
          ),
        ),
        child: SafeArea(
          child: AppPrimaryButton(
            label: '${loc.t('proceed_checkout')} • ₹$totalRupees',
            icon: Icons.lock_outline_rounded,
            onPressed: () => _openCheckout(context),
          ),
        ),
      ),
    );
  }
}

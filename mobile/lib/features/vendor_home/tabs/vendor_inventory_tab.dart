import 'package:flutter/material.dart';
import '../../../core/api/api_client.dart';
import '../../../core/design_system/app_colors.dart';
import '../../../core/design_system/app_decorations.dart';
import '../../../core/design_system/components/app_buttons.dart';
import '../../../core/design_system/components/app_cards.dart';
import '../../../core/design_system/components/app_inputs.dart';
import '../../../core/design_system/components/app_sheets.dart';
import '../../../core/design_system/components/app_states.dart';
import '../../../core/localization/locale_provider.dart';
import '../../../shared/models/models.dart';

class VendorInventoryTab extends StatelessWidget {
  final ApiClient apiClient;
  final List<ProductModel> products;
  final Map<String, InventoryModel> inventoryMap;
  final bool isLoading;
  final VoidCallback onRefresh;

  const VendorInventoryTab({
    super.key,
    required this.apiClient,
    required this.products,
    required this.inventoryMap,
    required this.isLoading,
    required this.onRefresh,
  });

  void _showAdjustStockDialog(BuildContext context, ProductModel product, InventoryModel? inv) {
    final loc = LocaleProvider.of(context);
    final deltaCtrl = TextEditingController(text: '10');
    final reasonCtrl = TextEditingController(text: 'Mandi Restock');
    bool isSubmitting = false;

    showDialog(
      context: context,
      builder: (dlgCtx) => StatefulBuilder(
        builder: (ctx, setDlgState) => AlertDialog(
          backgroundColor: AppColors.surface,
          shape: const RoundedRectangleBorder(borderRadius: AppRadii.roundedLg),
          title: Text('Adjust Stock: ${product.name}', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15)),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                'Current Available: ${inv?.availableQty ?? 0}',
                style: const TextStyle(fontSize: 13, color: AppColors.textSecondary, fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 14),
              AppTextField(label: 'Quantity Delta (+/-)', controller: deltaCtrl, keyboardType: const TextInputType.numberWithOptions(signed: true)),
              const SizedBox(height: 12),
              AppTextField(label: 'Reason', controller: reasonCtrl),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(dlgCtx),
              child: Text(loc.t('cancel'), style: const TextStyle(color: AppColors.textSecondary)),
            ),
                    ElevatedButton(
                      onPressed: isSubmitting
                          ? null
                          : () async {
                              final delta = int.tryParse(deltaCtrl.text.trim()) ?? 0;
                              if (delta == 0) return;

                              final currentOnHand = inv?.onHandQty ?? 0;
                              if (currentOnHand + delta < 0) {
                                AppFeedback.showError(
                                  context,
                                  'Resulting on-hand quantity cannot be negative ($currentOnHand + $delta = ${currentOnHand + delta})',
                                );
                                return;
                              }

                              setDlgState(() => isSubmitting = true);
                              try {
                                await apiClient.post('/inventory/adjust', body: {
                                  'product_id': product.id,
                                  'qty_delta': delta,
                                  'reason': reasonCtrl.text.trim(),
                                });
                                if (context.mounted) {
                                  Navigator.pop(dlgCtx);
                                  AppFeedback.showSuccess(context, 'Stock adjusted successfully');
                                  onRefresh();
                                }
                              } catch (e) {
                                setDlgState(() => isSubmitting = false);
                                if (context.mounted) {
                                  AppFeedback.showError(context, 'Adjustment error: $e');
                                }
                              }
                            },
                      style: ElevatedButton.styleFrom(backgroundColor: AppColors.primary),
                      child: isSubmitting
                          ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                          : Text(loc.t('save')),
                    ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (isLoading) {
      return Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: List.generate(4, (_) => const AppCardSkeleton()),
        ),
      );
    }

    if (products.isEmpty) {
      return const AppEmptyState(
        icon: Icons.inventory_2_outlined,
        title: 'No inventory items',
        subtitle: 'Products in your catalog will appear here with real-time stock levels',
      );
    }

    return RefreshIndicator(
      onRefresh: () async => onRefresh(),
      color: AppColors.primary,
      child: ListView.builder(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 100),
        itemCount: products.length,
        itemBuilder: (ctx, index) {
          final product = products[index];
          final inv = inventoryMap[product.id];
          final available = inv?.availableQty ?? 0;
          final onHand = inv?.onHandQty ?? 0;
          final reserved = inv?.reservedQty ?? 0;

          return AppCard(
            padding: const EdgeInsets.all(14),
            margin: const EdgeInsets.only(bottom: 12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      child: Text(
                        product.name,
                        style: const TextStyle(fontSize: 15, fontWeight: FontWeight.bold, color: AppColors.textPrimary),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                      decoration: BoxDecoration(
                        color: available > 5 ? AppColors.mintSoft : (available > 0 ? AppColors.warningBg : AppColors.errorBg),
                        borderRadius: AppRadii.roundedPill,
                      ),
                      child: Text(
                        available > 5 ? 'In Stock' : (available > 0 ? 'Low Stock' : 'Out of Stock'),
                        style: TextStyle(
                          fontSize: 10,
                          fontWeight: FontWeight.bold,
                          color: available > 5 ? AppColors.primaryDark : (available > 0 ? AppColors.warningText : AppColors.errorText),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('Available: $available', style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.primary)),
                    Text('On Hand: $onHand • Reserved: $reserved', style: const TextStyle(fontSize: 11, color: AppColors.textTertiary)),
                  ],
                ),
                const Divider(height: 16),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    AppSecondaryButton(
                      label: 'Adjust Stock',
                      icon: Icons.tune_rounded,
                      isFullWidth: false,
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                      onPressed: () => _showAdjustStockDialog(context, product, inv),
                    ),
                  ],
                ),
              ],
            ),
          );
        },
      ),
    );
  }
}

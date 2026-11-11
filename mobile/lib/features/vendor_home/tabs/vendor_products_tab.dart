import 'package:flutter/material.dart';
import '../../../core/api/api_client.dart';
import '../../../core/design_system/app_colors.dart';
import '../../../core/design_system/app_decorations.dart';
import '../../../core/design_system/components/app_cards.dart';
import '../../../core/design_system/components/app_inputs.dart';
import '../../../core/design_system/components/app_sheets.dart';
import '../../../core/design_system/components/app_states.dart';
import '../../../core/localization/locale_provider.dart';
import '../../../shared/models/models.dart';

class VendorProductsTab extends StatefulWidget {
  final ApiClient apiClient;
  final List<ProductModel> products;
  final List<CategoryModel> categories;
  final bool isLoading;
  final VoidCallback onRefresh;

  const VendorProductsTab({
    super.key,
    required this.apiClient,
    required this.products,
    required this.categories,
    required this.isLoading,
    required this.onRefresh,
  });

  @override
  State<VendorProductsTab> createState() => _VendorProductsTabState();
}

class _VendorProductsTabState extends State<VendorProductsTab> {
  String _searchQuery = '';
  final TextEditingController _searchController = TextEditingController();

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _showAddProductDialog() {
    final loc = LocaleProvider.of(context);
    final nameCtrl = TextEditingController();
    final priceCtrl = TextEditingController();
    final descCtrl = TextEditingController();
    String? selectedCatId = widget.categories.isNotEmpty ? widget.categories.first.id : null;
    bool isSubmitting = false;

    showDialog(
      context: context,
      builder: (dlgCtx) => StatefulBuilder(
        builder: (ctx, setDlgState) => AlertDialog(
          backgroundColor: AppColors.surface,
          shape: const RoundedRectangleBorder(borderRadius: AppRadii.roundedLg),
          title: Text(loc.t('add_product'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                AppTextField(label: loc.t('product_name'), controller: nameCtrl, hint: 'e.g. Robusta Banana 1kg'),
                const SizedBox(height: 12),
                AppTextField(label: 'Price in Rupees (₹)', controller: priceCtrl, keyboardType: TextInputType.number, hint: '45.00'),
                const SizedBox(height: 12),
                AppTextField(label: loc.t('product_desc'), controller: descCtrl, maxLines: 2),
                const SizedBox(height: 12),
                if (widget.categories.isNotEmpty) ...[
                  const Align(
                    alignment: Alignment.centerLeft,
                    child: Text('Category', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.textPrimary)),
                  ),
                  const SizedBox(height: 6),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12),
                    decoration: BoxDecoration(
                      color: AppColors.surface,
                      borderRadius: AppRadii.roundedMd,
                      border: Border.all(color: AppColors.border),
                    ),
                    child: DropdownButtonHideUnderline(
                      child: DropdownButton<String>(
                        value: selectedCatId,
                        isExpanded: true,
                        items: widget.categories.map((c) {
                          return DropdownMenuItem(value: c.id, child: Text(c.name, style: const TextStyle(fontSize: 13)));
                        }).toList(),
                        onChanged: (val) => setDlgState(() => selectedCatId = val),
                      ),
                    ),
                  ),
                ],
              ],
            ),
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
                      final name = nameCtrl.text.trim();
                      final priceText = priceCtrl.text.trim();
                      if (name.isEmpty || priceText.isEmpty) return;

                      final priceDouble = double.tryParse(priceText) ?? 0.0;
                      final pricePaise = (priceDouble * 100).toInt();

                      setDlgState(() => isSubmitting = true);
                      try {
                        await widget.apiClient.post('/products', body: {
                          'name': name,
                          'category_id': selectedCatId,
                          'price_paise': pricePaise,
                          'description': descCtrl.text.trim(),
                          'active': true,
                        });
                        if (mounted) {
                          Navigator.pop(dlgCtx);
                          AppFeedback.showSuccess(context, 'Product added successfully');
                          widget.onRefresh();
                        }
                      } catch (e) {
                        setDlgState(() => isSubmitting = false);
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

  Future<void> _toggleActive(ProductModel product) async {
    try {
      await widget.apiClient.patch('/products/${product.id}', body: {
        'active': !product.active,
      });
      widget.onRefresh();
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);

    final filteredProducts = widget.products.where((p) {
      if (_searchQuery.isEmpty) return true;
      return p.name.toLowerCase().contains(_searchQuery.toLowerCase());
    }).toList();

    return Scaffold(
      backgroundColor: Colors.transparent,
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _showAddProductDialog,
        backgroundColor: AppColors.primary,
        icon: const Icon(Icons.add_rounded, color: Colors.white),
        label: Text(loc.t('add_product'), style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
      ),
      body: RefreshIndicator(
        onRefresh: () async => widget.onRefresh(),
        color: AppColors.primary,
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
              child: AppSearchField(
                controller: _searchController,
                hint: loc.t('search_products'),
                onChanged: (val) => setState(() => _searchQuery = val.trim()),
              ),
            ),
            Expanded(
              child: widget.isLoading
                  ? ListView.builder(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      itemCount: 4,
                      itemBuilder: (_, __) => const AppCardSkeleton(),
                    )
                  : filteredProducts.isEmpty
                      ? AppEmptyState(
                          icon: Icons.inventory_2_outlined,
                          title: 'No products in catalog',
                          subtitle: 'Tap the button below to add your first product',
                          actionLabel: loc.t('add_product'),
                          onAction: _showAddProductDialog,
                        )
                      : ListView.builder(
                        padding: const EdgeInsets.fromLTRB(16, 4, 16, 120),
                        itemCount: filteredProducts.length,
                        itemBuilder: (ctx, index) {
                          final product = filteredProducts[index];
                          final priceRupees = (product.pricePaise / 100).toStringAsFixed(2);

                          return AppCard(
                            padding: const EdgeInsets.all(12),
                            margin: const EdgeInsets.only(bottom: 10),
                            child: Row(
                              children: [
                                Container(
                                  width: 44,
                                  height: 44,
                                  decoration: BoxDecoration(
                                    color: product.active ? AppColors.mintSoft : AppColors.surfaceMuted,
                                    borderRadius: AppRadii.roundedMd,
                                  ),
                                  child: Icon(
                                    Icons.local_offer_outlined,
                                    color: product.active ? AppColors.primary : AppColors.textTertiary,
                                    size: 22,
                                  ),
                                ),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        product.name,
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
                                        '₹$priceRupees',
                                        style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: AppColors.primary),
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 8),
                                InkWell(
                                  borderRadius: AppRadii.roundedPill,
                                  onTap: () => _toggleActive(product),
                                  child: Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                                    decoration: BoxDecoration(
                                      color: product.active ? AppColors.mintSoft : AppColors.surfaceMuted,
                                      borderRadius: AppRadii.roundedPill,
                                      border: Border.all(
                                        color: product.active ? AppColors.primary : AppColors.border,
                                        width: 1,
                                      ),
                                    ),
                                    child: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Icon(
                                          product.active ? Icons.check_circle_rounded : Icons.pause_circle_outline_rounded,
                                          size: 14,
                                          color: product.active ? AppColors.primary : AppColors.textTertiary,
                                        ),
                                        const SizedBox(width: 4),
                                        Text(
                                          product.active
                                              ? (loc.currentLanguage == 'kn' ? 'ಲಭ್ಯವಿದೆ' : 'Available')
                                              : (loc.currentLanguage == 'kn' ? 'ಲಭ್ಯವಿಲ್ಲ' : 'Disabled'),
                                          style: TextStyle(
                                            fontSize: 11,
                                            fontWeight: FontWeight.bold,
                                            color: product.active ? AppColors.primaryDark : AppColors.textTertiary,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
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

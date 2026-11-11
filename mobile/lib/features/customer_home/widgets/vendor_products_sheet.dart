import 'package:flutter/material.dart';
import '../../../core/api/api_client.dart';
import '../../../core/design_system/app_colors.dart';
import '../../../core/design_system/app_decorations.dart';
import '../../../core/design_system/components/app_badges.dart';
import '../../../core/design_system/components/app_buttons.dart';
import '../../../core/design_system/components/app_cards.dart';
import '../../../core/design_system/components/app_inputs.dart';
import '../../../core/design_system/components/app_states.dart';
import '../../../core/localization/locale_provider.dart';
import '../../../shared/models/models.dart';

/// Bottom sheet displaying vendor's products catalogue with search and Add-to-Cart
class VendorProductsSheet extends StatefulWidget {
  final ApiClient apiClient;
  final NearbyVendorModel vendor;
  final CartModel? cart;
  final Function(ProductModel product, int qty) onAddToCart;
  final VoidCallback onViewCart;

  const VendorProductsSheet({
    super.key,
    required this.apiClient,
    required this.vendor,
    this.cart,
    required this.onAddToCart,
    required this.onViewCart,
  });

  @override
  State<VendorProductsSheet> createState() => _VendorProductsSheetState();
}

class _VendorProductsSheetState extends State<VendorProductsSheet> {
  List<ProductModel> _products = [];
  bool _isLoading = true;
  String? _error;
  String _searchQuery = '';
  final TextEditingController _searchController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _fetchProducts();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _fetchProducts() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final res = await widget.apiClient.get(
        '/products',
        queryParams: {'vendor_id': widget.vendor.vendorId, 'limit': '100'},
      );

      if (mounted) {
        setState(() {
          if (res != null && res['data'] != null) {
            _products = (res['data'] as List).map((e) => ProductModel.fromJson(e)).toList();
          }
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  int _getCartQty(String productId) {
    if (widget.cart == null) return 0;
    for (final item in widget.cart!.items) {
      if (item.productId == productId) return item.quantity;
    }
    return 0;
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);
    final isDifferentVendorCart = widget.cart != null &&
        widget.cart!.items.isNotEmpty &&
        widget.cart!.vendorId != widget.vendor.vendorId;

    final filteredProducts = _products.where((p) {
      if (_searchQuery.isEmpty) return true;
      return p.name.toLowerCase().contains(_searchQuery.toLowerCase());
    }).toList();

    return Container(
      constraints: BoxConstraints(
        maxHeight: MediaQuery.of(context).size.height * 0.85,
      ),
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Vendor Header Card
          Row(
            children: [
              Container(
                width: 48,
                height: 48,
                decoration: BoxDecoration(
                  color: AppColors.mintSoft,
                  borderRadius: AppRadii.roundedMd,
                  border: Border.all(color: AppColors.mint),
                ),
                child: const Icon(Icons.storefront_rounded, color: AppColors.primary, size: 26),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      widget.vendor.businessName,
                      style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: AppColors.textPrimary),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 2),
                    Row(
                      children: [
                        AppStatusBadge.availability(widget.vendor.isOpen),
                        const SizedBox(width: 8),
                        Text(
                          '${widget.vendor.distanceKm.toStringAsFixed(1)} km away',
                          style: const TextStyle(fontSize: 12, color: AppColors.textSecondary, fontWeight: FontWeight.w500),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // Search Field
          AppSearchField(
            controller: _searchController,
            hint: loc.t('search'),
            onChanged: (val) => setState(() => _searchQuery = val.trim()),
          ),
          const SizedBox(height: 12),

          // Conflict banner if user has items from another vendor
          if (isDifferentVendorCart)
            Container(
              padding: const EdgeInsets.all(10),
              margin: const EdgeInsets.only(bottom: 12),
              decoration: BoxDecoration(
                color: AppColors.warningBg,
                borderRadius: AppRadii.roundedMd,
                border: Border.all(color: AppColors.warning.withValues(alpha: 0.3)),
              ),
              child: Row(
                children: [
                  const Icon(Icons.info_outline, color: AppColors.warning, size: 16),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      loc.t('replace_cart_message'),
                      style: const TextStyle(fontSize: 11, color: AppColors.warningText, fontWeight: FontWeight.w500),
                    ),
                  ),
                ],
              ),
            ),

          // Products List View
          Expanded(
            child: _isLoading
                ? ListView.builder(
                    itemCount: 4,
                    itemBuilder: (_, __) => const AppCardSkeleton(),
                  )
                : _error != null
                    ? AppErrorState(message: _error!, onRetry: _fetchProducts)
                    : filteredProducts.isEmpty
                        ? AppEmptyState(
                            icon: Icons.inventory_2_outlined,
                            title: loc.t('no_products_for_forecast'),
                            subtitle: loc.t('search'),
                          )
                        : ListView.builder(
                            padding: const EdgeInsets.only(bottom: 80),
                            itemCount: filteredProducts.length,
                            itemBuilder: (ctx, index) {
                              final product = filteredProducts[index];
                              final inCartQty = _getCartQty(product.id);

                              return AppProductCard(
                                name: product.name,
                                pricePaise: product.pricePaise,
                                isAvailable: product.active,
                                inCartQty: inCartQty,
                                onAdd: () => widget.onAddToCart(product, inCartQty + 1),
                                onRemove: inCartQty > 0
                                    ? () => widget.onAddToCart(product, inCartQty - 1)
                                    : null,
                              );
                            },
                          ),
          ),

          // Sticky View Cart CTA when items exist in cart
          if (widget.cart != null && widget.cart!.items.isNotEmpty)
            SafeArea(
              child: Container(
                padding: const EdgeInsets.symmetric(vertical: 8),
                child: AppPrimaryButton(
                  label: '${loc.t('cart')} (${widget.cart!.items.length} items) • ₹${(widget.cart!.totalPaise / 100).toStringAsFixed(2)}',
                  icon: Icons.shopping_cart_checkout_rounded,
                  onPressed: widget.onViewCart,
                ),
              ),
            ),
        ],
      ),
    );
  }
}

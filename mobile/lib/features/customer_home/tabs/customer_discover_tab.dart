import 'package:flutter/material.dart';
import '../../../core/api/api_client.dart';
import '../../../core/design_system/app_colors.dart';
import '../../../core/design_system/components/app_badges.dart';
import '../../../core/design_system/components/app_cards.dart';
import '../../../core/design_system/components/app_inputs.dart';
import '../../../core/design_system/components/app_sheets.dart';
import '../../../core/design_system/components/app_states.dart';
import '../../../core/localization/locale_provider.dart';
import '../../../shared/models/models.dart';
import '../explore_map_screen.dart';
import '../widgets/vendor_products_sheet.dart';

class CustomerDiscoverTab extends StatefulWidget {
  final ApiClient apiClient;
  final List<NearbyVendorModel> nearbyVendors;
  final List<CategoryModel> categories;
  final CartModel? cart;
  final bool isLoading;
  final double currentLat;
  final double currentLng;
  final VoidCallback onRefresh;
  final VoidCallback onFetchGPS;
  final Function(ProductModel product, int qty) onAddToCart;
  final VoidCallback onViewCart;

  const CustomerDiscoverTab({
    super.key,
    required this.apiClient,
    required this.nearbyVendors,
    required this.categories,
    this.cart,
    required this.isLoading,
    this.currentLat = 12.9767,
    this.currentLng = 77.5713,
    required this.onRefresh,
    required this.onFetchGPS,
    required this.onAddToCart,
    required this.onViewCart,
  });

  @override
  State<CustomerDiscoverTab> createState() => _CustomerDiscoverTabState();
}

class _CustomerDiscoverTabState extends State<CustomerDiscoverTab> {
  String? _selectedCategory;
  String _searchQuery = '';
  final TextEditingController _searchController = TextEditingController();

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _openVendorProducts(NearbyVendorModel vendor) {
    AppBottomSheet.show(
      context: context,
      title: vendor.businessName,
      child: VendorProductsSheet(
        apiClient: widget.apiClient,
        vendor: vendor,
        cart: widget.cart,
        onAddToCart: widget.onAddToCart,
        onViewCart: () {
          Navigator.pop(context);
          widget.onViewCart();
        },
      ),
    );
  }

  Future<void> _openExploreMap({String? selectedVendorId}) async {
    final result = await Navigator.push<ExploreMapResult>(
      context,
      MaterialPageRoute(
        builder: (ctx) => ExploreMapScreen(
          apiClient: widget.apiClient,
          nearbyVendors: widget.nearbyVendors,
          categories: widget.categories,
          cart: widget.cart,
          initialLat: widget.currentLat,
          initialLng: widget.currentLng,
          initialSearchQuery: _searchQuery,
          initialSelectedCategoryId: _selectedCategory,
          initialSelectedVendorId: selectedVendorId,
          onAddToCart: widget.onAddToCart,
          onViewCart: widget.onViewCart,
        ),
      ),
    );

    if (result != null && mounted) {
      setState(() {
        _searchQuery = result.searchQuery;
        _searchController.text = result.searchQuery;
        _selectedCategory = result.selectedCategoryId;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);

    // Filter vendors by search query and category
    final filteredVendors = widget.nearbyVendors.where((v) {
      if (_searchQuery.isNotEmpty) {
        final q = _searchQuery.toLowerCase();
        final matchName = v.businessName.toLowerCase().contains(q);
        final matchDesc = (v.description ?? '').toLowerCase().contains(q);
        if (!matchName && !matchDesc) return false;
      }
      if (_selectedCategory != null) {
        final cat = widget.categories.firstWhere(
          (c) => c.id == _selectedCategory,
          orElse: () => CategoryModel(id: '', name: '', active: true),
        );
        if (cat.name.isNotEmpty &&
            !v.categoryName.toLowerCase().contains(cat.name.toLowerCase())) {
          return false;
        }
      }
      return true;
    }).toList();

    return RefreshIndicator(
      onRefresh: () async => widget.onRefresh(),
      color: AppColors.primary,
      child: CustomScrollView(
        slivers: [
          // Header search, map action and categories
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Search Bar with Dedicated Map Button
                  Row(
                    children: [
                      Expanded(
                        child: AppSearchField(
                          controller: _searchController,
                          hint: loc.t('search_products'),
                          onChanged: (val) => setState(() => _searchQuery = val.trim()),
                          onClear: () => setState(() => _searchQuery = ''),
                        ),
                      ),
                      const SizedBox(width: 8),
                      // Dedicated Fullscreen Map Action
                      Material(
                        color: AppColors.primary,
                        borderRadius: BorderRadius.circular(22),
                        elevation: 1.5,
                        child: InkWell(
                          borderRadius: BorderRadius.circular(22),
                          onTap: () => _openExploreMap(),
                          child: Container(
                            height: 44,
                            padding: const EdgeInsets.symmetric(horizontal: 14),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                const Icon(Icons.map_rounded, color: Colors.white, size: 18),
                                const SizedBox(width: 6),
                                Text(
                                  loc.t('explore_map'),
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 13,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),

                  // Horizontal Category Filters
                  if (widget.categories.isNotEmpty) ...[
                    SizedBox(
                      height: 36,
                      child: ListView(
                        scrollDirection: Axis.horizontal,
                        children: [
                          AppFilterChip(
                            label: loc.t('all'),
                            isSelected: _selectedCategory == null,
                            onTap: () => setState(() => _selectedCategory = null),
                          ),
                          const SizedBox(width: 8),
                          ...widget.categories.map((c) {
                            return Padding(
                              padding: const EdgeInsets.only(right: 8),
                              child: AppFilterChip(
                                label: c.name,
                                isSelected: _selectedCategory == c.id,
                                onTap: () => setState(() {
                                  _selectedCategory = _selectedCategory == c.id ? null : c.id;
                                }),
                              ),
                            );
                          }),
                        ],
                      ),
                    ),
                    const SizedBox(height: 12),
                  ],

                  // Section Title Row
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          '${loc.t('nearby_vendors')} (${filteredVendors.length})',
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w700,
                            color: AppColors.textPrimary,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      TextButton.icon(
                        onPressed: widget.onFetchGPS,
                        icon: const Icon(Icons.my_location_rounded, size: 14, color: AppColors.primary),
                        label: const Text('GPS', style: TextStyle(fontSize: 12, color: AppColors.primary, fontWeight: FontWeight.bold)),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),

          // Vendor List or Empty State
          if (widget.isLoading)
            SliverPadding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              sliver: SliverList(
                delegate: SliverChildBuilderDelegate(
                  (_, __) => const AppCardSkeleton(),
                  childCount: 4,
                ),
              ),
            )
          else if (filteredVendors.isEmpty)
            SliverFillRemaining(
              hasScrollBody: false,
              child: AppEmptyState(
                icon: Icons.storefront_outlined,
                title: loc.t('empty_vendors_title'),
                subtitle: loc.t('empty_vendors_subtitle'),
                actionLabel: loc.t('refresh'),
                onAction: widget.onRefresh,
              ),
            )
          else
            SliverPadding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
              sliver: SliverList(
                delegate: SliverChildBuilderDelegate(
                  (ctx, index) {
                    final vendor = filteredVendors[index];
                    return AppVendorCard(
                      businessName: vendor.businessName,
                      distanceKm: vendor.distanceKm,
                      address: vendor.addressLine1.isNotEmpty ? vendor.addressLine1 : 'Bangalore',
                      isOpen: vendor.isOpen,
                      category: vendor.categoryName,
                      onTap: () => _openVendorProducts(vendor),
                    );
                  },
                  childCount: filteredVendors.length,
                ),
              ),
            ),
        ],
      ),
    );
  }
}

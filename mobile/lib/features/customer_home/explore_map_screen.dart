import 'package:flutter/material.dart';
import '../../core/api/api_client.dart';
import '../../core/design_system/app_colors.dart';
import '../../core/design_system/app_decorations.dart';
import '../../core/design_system/components/app_badges.dart';
import '../../core/design_system/components/app_buttons.dart';
import '../../core/design_system/components/app_cards.dart';
import '../../core/design_system/components/app_inputs.dart';
import '../../core/design_system/components/app_sheets.dart';
import '../../core/design_system/components/app_states.dart';
import '../../core/localization/locale_provider.dart';
import '../../shared/models/models.dart';
import '../../shared/widgets/osm_map_view.dart';
import 'widgets/vendor_products_sheet.dart';

class ExploreMapResult {
  final String searchQuery;
  final String? selectedCategoryId;
  final String? selectedVendorId;

  const ExploreMapResult({
    required this.searchQuery,
    this.selectedCategoryId,
    this.selectedVendorId,
  });
}

class ExploreMapScreen extends StatefulWidget {
  final ApiClient apiClient;
  final List<NearbyVendorModel> nearbyVendors;
  final List<CategoryModel> categories;
  final CartModel? cart;
  final double initialLat;
  final double initialLng;
  final String initialSearchQuery;
  final String? initialSelectedCategoryId;
  final String? initialSelectedVendorId;
  final Function(ProductModel product, int qty) onAddToCart;
  final VoidCallback onViewCart;

  const ExploreMapScreen({
    super.key,
    required this.apiClient,
    required this.nearbyVendors,
    required this.categories,
    this.cart,
    required this.initialLat,
    required this.initialLng,
    this.initialSearchQuery = '',
    this.initialSelectedCategoryId,
    this.initialSelectedVendorId,
    required this.onAddToCart,
    required this.onViewCart,
  });

  @override
  State<ExploreMapScreen> createState() => _ExploreMapScreenState();
}

class _ExploreMapScreenState extends State<ExploreMapScreen> {
  final GlobalKey<OsmMapViewState> _mapKey = GlobalKey<OsmMapViewState>();
  late TextEditingController _searchController;
  late String _searchQuery;
  String? _selectedCategory;
  NearbyVendorModel? _selectedVendor;
  bool _isSearchExpanded = false;
  final DraggableScrollableController _sheetController = DraggableScrollableController();

  @override
  void initState() {
    super.initState();
    _searchQuery = widget.initialSearchQuery;
    _searchController = TextEditingController(text: _searchQuery);
    _selectedCategory = widget.initialSelectedCategoryId;

    if (widget.initialSelectedVendorId != null) {
      try {
        _selectedVendor = widget.nearbyVendors.firstWhere(
          (v) => v.vendorId == widget.initialSelectedVendorId,
        );
      } catch (_) {}
    }

    if (_selectedVendor == null && widget.nearbyVendors.isNotEmpty) {
      _selectedVendor = widget.nearbyVendors.first;
    }
  }

  @override
  void dispose() {
    _searchController.dispose();
    _sheetController.dispose();
    super.dispose();
  }

  List<NearbyVendorModel> get _filteredVendors {
    return widget.nearbyVendors.where((v) {
      if (_searchQuery.isNotEmpty &&
          !v.businessName.toLowerCase().contains(_searchQuery.toLowerCase()) &&
          !(v.description ?? '').toLowerCase().contains(_searchQuery.toLowerCase())) {
        return false;
      }
      return true;
    }).toList();
  }

  void _selectVendor(NearbyVendorModel vendor, {bool panCamera = true}) {
    setState(() {
      _selectedVendor = vendor;
    });

    if (panCamera && vendor.latitude != 0 && vendor.longitude != 0) {
      _mapKey.currentState?.animateToCoordinates(
        vendor.latitude,
        vendor.longitude,
        zoom: 15.0,
      );
    }
  }

  void _openVendorStorefront(NearbyVendorModel vendor) {
    AppBottomSheet.show(
      context: context,
      title: vendor.businessName,
      child: VendorProductsSheet(
        apiClient: widget.apiClient,
        vendor: vendor,
        cart: widget.cart,
        onAddToCart: widget.onAddToCart,
        onViewCart: () {
          Navigator.pop(context); // close sheet
          widget.onViewCart();
        },
      ),
    );
  }

  void _recenterToCurrentLocation() {
    _mapKey.currentState?.animateToCoordinates(
      widget.initialLat,
      widget.initialLng,
      zoom: 14.0,
    );
  }

  void _handleBack() {
    Navigator.pop(
      context,
      ExploreMapResult(
        searchQuery: _searchQuery,
        selectedCategoryId: _selectedCategory,
        selectedVendorId: _selectedVendor?.vendorId,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);
    final filtered = _filteredVendors;

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, _) {
        if (!didPop) _handleBack();
      },
      child: Scaffold(
        backgroundColor: AppColors.background,
        body: Stack(
          children: [
            // 1. Full Screen Interactive Map
            Positioned.fill(
              child: OsmMapView(
                key: _mapKey,
                height: double.infinity,
                apiClient: widget.apiClient,
                initialLat: widget.initialLat,
                initialLng: widget.initialLng,
                customerLat: widget.initialLat,
                customerLng: widget.initialLng,
                customerTitle: 'Your Location',
                initialZoom: 13.5,
                nearbyVendors: filtered,
                selectedVendorId: _selectedVendor?.vendorId,
                onVendorSelected: (vendor) {
                  _selectVendor(vendor, panCamera: true);
                },
              ),
            ),

            // 2. Compact Top Header Overlay
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
                    padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        // Top Bar: Back, Title, Search/Filter Toggles
                        Row(
                          children: [
                            // Back Button
                            Material(
                              color: Colors.white,
                              shape: const CircleBorder(),
                              elevation: 3,
                              child: InkWell(
                                customBorder: const CircleBorder(),
                                onTap: _handleBack,
                                child: const Padding(
                                  padding: EdgeInsets.all(8),
                                  child: Icon(Icons.arrow_back_rounded, size: 20, color: AppColors.textPrimary),
                                ),
                              ),
                            ),
                            const SizedBox(width: 12),
                            // Title & Counter
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Text(
                                    loc.t('explore_map'),
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
                                    '${filtered.length} ${loc.t('nearby_vendors')}',
                                    style: TextStyle(
                                      color: Colors.white.withValues(alpha: 0.9),
                                      fontSize: 12,
                                      fontWeight: FontWeight.w500,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            // Search toggle button
                            Material(
                              color: _isSearchExpanded ? AppColors.mintSoft : Colors.white,
                              shape: const CircleBorder(),
                              elevation: 3,
                              child: InkWell(
                                customBorder: const CircleBorder(),
                                onTap: () => setState(() => _isSearchExpanded = !_isSearchExpanded),
                                child: Padding(
                                  padding: const EdgeInsets.all(8),
                                  child: Icon(
                                    _isSearchExpanded ? Icons.close_rounded : Icons.search_rounded,
                                    size: 20,
                                    color: _isSearchExpanded ? AppColors.primary : AppColors.textPrimary,
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),

                        // Expandable search bar
                        if (_isSearchExpanded) ...[
                          const SizedBox(height: 10),
                          AppSearchField(
                            controller: _searchController,
                            hint: loc.t('search_products'),
                            onChanged: (val) => setState(() => _searchQuery = val.trim()),
                            onClear: () => setState(() => _searchQuery = ''),
                          ),
                        ],

                        // Horizontal Category Filter Pills
                        if (widget.categories.isNotEmpty) ...[
                          const SizedBox(height: 8),
                          SizedBox(
                            height: 32,
                            child: ListView(
                              scrollDirection: Axis.horizontal,
                              children: [
                                AppFilterChip(
                                  label: loc.t('all'),
                                  isSelected: _selectedCategory == null,
                                  onTap: () => setState(() => _selectedCategory = null),
                                ),
                                const SizedBox(width: 6),
                                ...widget.categories.map((c) {
                                  return Padding(
                                    padding: const EdgeInsets.only(right: 6),
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
                        ],
                      ],
                    ),
                  ),
                ),
              ),
            ),

            // 3. Floating Map Controls (Right Side)
            Positioned(
              right: 16,
              bottom: 220,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // GPS / Recenter Action
                  FloatingActionButton.small(
                    heroTag: 'recenter_gps_btn',
                    backgroundColor: Colors.white,
                    foregroundColor: AppColors.primary,
                    elevation: 3,
                    onPressed: _recenterToCurrentLocation,
                    tooltip: 'Current Location',
                    child: const Icon(Icons.my_location_rounded, size: 20),
                  ),
                ],
              ),
            ),

            // 4. Draggable Vendor Bottom Sheet (Compact Preview & Expanded List)
            DraggableScrollableSheet(
              controller: _sheetController,
              initialChildSize: 0.28,
              minChildSize: 0.10,
              maxChildSize: 0.68,
              builder: (ctx, scrollController) {
                return Container(
                  decoration: BoxDecoration(
                    color: AppColors.surface,
                    borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.15),
                        blurRadius: 12,
                        offset: const Offset(0, -3),
                      ),
                    ],
                  ),
                  child: ListView(
                    controller: scrollController,
                    padding: const EdgeInsets.fromLTRB(16, 8, 16, 20),
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

                      // Selected Vendor Card Preview
                      if (_selectedVendor != null) ...[
                        Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: AppColors.mintSoft.withValues(alpha: 0.5),
                            borderRadius: AppRadii.roundedLg,
                            border: Border.all(color: AppColors.mint),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Container(
                                    width: 44,
                                    height: 44,
                                    decoration: const BoxDecoration(
                                      color: AppColors.primary,
                                      borderRadius: AppRadii.roundedMd,
                                    ),
                                    child: const Icon(
                                      Icons.storefront_rounded,
                                      color: Colors.white,
                                      size: 24,
                                    ),
                                  ),
                                  const SizedBox(width: 10),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          _selectedVendor!.businessName,
                                          style: const TextStyle(
                                            fontSize: 16,
                                            fontWeight: FontWeight.w800,
                                            color: AppColors.textPrimary,
                                          ),
                                          maxLines: 1,
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                        const SizedBox(height: 2),
                                        Row(
                                          children: [
                                            AppStatusBadge.availability(_selectedVendor!.isOpen),
                                            const SizedBox(width: 6),
                                            Text(
                                              '${_selectedVendor!.distanceKm.toStringAsFixed(1)} km away',
                                              style: const TextStyle(
                                                fontSize: 12,
                                                fontWeight: FontWeight.w600,
                                                color: AppColors.accent,
                                              ),
                                            ),
                                          ],
                                        ),
                                        if (_selectedVendor!.addressLine1.isNotEmpty) ...[
                                          const SizedBox(height: 3),
                                          Text(
                                            _selectedVendor!.addressLine1,
                                            style: const TextStyle(
                                              fontSize: 11,
                                              color: AppColors.textSecondary,
                                            ),
                                            maxLines: 1,
                                            overflow: TextOverflow.ellipsis,
                                          ),
                                        ],
                                      ],
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 10),
                              // View Store Button
                              AppPrimaryButton(
                                label: loc.t('view_store'),
                                icon: Icons.storefront_outlined,
                                isFullWidth: true,
                                onPressed: () => _openVendorStorefront(_selectedVendor!),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 16),
                      ],

                      // Expanded List Header
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            '${loc.t('nearby_vendors')} (${filtered.length})',
                            style: const TextStyle(
                              fontSize: 15,
                              fontWeight: FontWeight.w700,
                              color: AppColors.textPrimary,
                            ),
                          ),
                          const Text(
                            'Swipe up to see all',
                            style: TextStyle(
                              fontSize: 11,
                              color: AppColors.textTertiary,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 10),

                      // Filtered Nearby Vendor Items
                      if (filtered.isEmpty)
                        AppEmptyState(
                          icon: Icons.storefront_outlined,
                          title: loc.t('empty_vendors_title'),
                          subtitle: loc.t('empty_vendors_subtitle'),
                        )
                      else
                        ...filtered.map((vendor) {
                          final isSelected = _selectedVendor?.vendorId == vendor.vendorId;
                          return Padding(
                            padding: const EdgeInsets.only(bottom: 8),
                            child: AppCard(
                              onTap: () => _selectVendor(vendor, panCamera: true),
                              color: isSelected ? AppColors.mintSoft.withValues(alpha: 0.3) : AppColors.surface,
                              border: Border.all(
                                color: isSelected ? AppColors.primary : AppColors.border,
                                width: isSelected ? 1.5 : 1.0,
                              ),
                              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                              child: Row(
                                children: [
                                  Container(
                                    width: 36,
                                    height: 36,
                                    decoration: BoxDecoration(
                                      color: isSelected ? AppColors.primary : AppColors.mintSoft,
                                      borderRadius: AppRadii.roundedSm,
                                    ),
                                    child: Icon(
                                      Icons.storefront_rounded,
                                      size: 18,
                                      color: isSelected ? Colors.white : AppColors.primary,
                                    ),
                                  ),
                                  const SizedBox(width: 10),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          vendor.businessName,
                                          style: TextStyle(
                                            fontSize: 14,
                                            fontWeight: isSelected ? FontWeight.w800 : FontWeight.w600,
                                            color: AppColors.textPrimary,
                                          ),
                                          maxLines: 1,
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                        Row(
                                          children: [
                                            Text(
                                              '${vendor.distanceKm.toStringAsFixed(1)} km',
                                              style: const TextStyle(
                                                fontSize: 11,
                                                fontWeight: FontWeight.w600,
                                                color: AppColors.accent,
                                              ),
                                            ),
                                            const SizedBox(width: 6),
                                            Flexible(
                                              child: Text(
                                                '•  ${vendor.categoryName}',
                                                style: const TextStyle(fontSize: 11, color: AppColors.textSecondary),
                                                maxLines: 1,
                                                overflow: TextOverflow.ellipsis,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ],
                                    ),
                                  ),
                                  AppStatusBadge.availability(vendor.isOpen),
                                ],
                              ),
                            ),
                          );
                        }),
                    ],
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

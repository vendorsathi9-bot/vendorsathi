import 'dart:async';
import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import '../../core/api/api_client.dart';
import '../../core/design_system/app_colors.dart';
import '../../core/design_system/app_decorations.dart';
import '../../core/design_system/components/app_navigation_bar.dart';
import '../../core/design_system/components/app_sheets.dart';
import '../../core/localization/locale_provider.dart';
import '../../shared/models/models.dart';
import 'tabs/customer_cart_tab.dart';
import 'tabs/customer_discover_tab.dart';
import 'tabs/customer_loyalty_tab.dart';
import 'tabs/customer_orders_tab.dart';
import 'explore_map_screen.dart';

class CustomerHomeScreen extends StatefulWidget {
  final ApiClient apiClient;
  final VoidCallback onLogout;
  final VoidCallback onSwitchToVendor;
  final bool canSwitchToVendor;

  const CustomerHomeScreen({
    super.key,
    required this.apiClient,
    required this.onLogout,
    required this.onSwitchToVendor,
    this.canSwitchToVendor = false,
  });

  @override
  State<CustomerHomeScreen> createState() => _CustomerHomeScreenState();
}

class _CustomerHomeScreenState extends State<CustomerHomeScreen> {
  int _currentIndex = 0;
  bool _isLoading = true;

  // Location (Default: Bangalore Majestic)
  double _currentLat = 12.9767;
  double _currentLng = 77.5713;

  // Data
  List<NearbyVendorModel> _nearbyVendors = [];
  List<CategoryModel> _categories = [];
  CartModel? _cart;
  List<OrderModel> _myOrders = [];
  List<CustomerAddressModel> _myAddresses = [];
  LoyaltySummaryModel? _loyalty;

  @override
  void initState() {
    super.initState();
    _loadCustomerData();
  }

  Future<void> _loadCustomerData() async {
    setState(() => _isLoading = true);
    try {
      // 1. Nearby Vendors
      final vendorRes = await widget.apiClient.get(
        '/mobility/nearby-vendors',
        queryParams: {'lat': _currentLat.toString(), 'lng': _currentLng.toString(), 'radius_km': '10.0'},
      ).catchError((_) => null);
      if (vendorRes != null && vendorRes is List) {
        _nearbyVendors = vendorRes.map((e) => NearbyVendorModel.fromJson(e)).toList();
      }

      // 2. Categories
      final catRes = await widget.apiClient.get('/categories').catchError((_) => null);
      if (catRes != null && catRes is List) {
        _categories = catRes.map((e) => CategoryModel.fromJson(e)).toList();
      }

      // 3. Cart
      final cartRes = await widget.apiClient.get('/carts/me').catchError((_) => null);
      if (cartRes != null) {
        _cart = CartModel.fromJson(cartRes);
      }

      // 4. Orders
      final ordersRes = await widget.apiClient.get('/orders?limit=100').catchError((_) => null);
      if (ordersRes != null && ordersRes['orders'] != null) {
        _myOrders = (ordersRes['orders'] as List).map((e) => OrderModel.fromJson(e)).toList();
      }

      // 5. Addresses
      final addrRes = await widget.apiClient.get('/customers/addresses/me').catchError((_) => null);
      if (addrRes != null && addrRes is List) {
        _myAddresses = addrRes.map((e) => CustomerAddressModel.fromJson(e)).toList();
      }

      // 6. Loyalty
      final loyaltyRes = await widget.apiClient.get('/loyalty/me').catchError((_) => null);
      if (loyaltyRes != null) {
        _loyalty = LoyaltySummaryModel.fromJson(loyaltyRes);
      }
    } catch (_) {
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _fetchGPSLocation() async {
    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        if (mounted) AppFeedback.showError(context, 'Location services are disabled');
        return;
      }
      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
        if (permission == LocationPermission.denied) return;
      }
      if (permission == LocationPermission.deniedForever) return;

      Position pos = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.medium,
        timeLimit: const Duration(seconds: 8),
      );

      setState(() {
        _currentLat = pos.latitude;
        _currentLng = pos.longitude;
      });
      await _loadCustomerData();
      if (mounted) AppFeedback.showSuccess(context, 'Location calibrated');
    } catch (e) {
      if (mounted) AppFeedback.showInfo(context, 'Using Bangalore Majestic anchor');
    }
  }

  // --- Cart Operations ---
  Future<void> _handleAddToCart(ProductModel product, int qty) async {
    final loc = LocaleProvider.of(context);

    // Single-vendor policy check
    if (_cart != null &&
        _cart!.items.isNotEmpty &&
        _cart!.vendorId != product.vendorId) {
      final shouldClear = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          backgroundColor: AppColors.surface,
          shape: const RoundedRectangleBorder(borderRadius: AppRadii.roundedLg),
          title: Text(loc.t('replace_cart_title'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
          content: Text(loc.t('replace_cart_message')),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: Text(loc.t('cancel'), style: const TextStyle(color: AppColors.textSecondary)),
            ),
            ElevatedButton(
              onPressed: () => Navigator.pop(ctx, true),
              style: ElevatedButton.styleFrom(backgroundColor: AppColors.primary),
              child: Text(loc.t('clear_and_add')),
            ),
          ],
        ),
      );

      if (shouldClear == true) {
        await widget.apiClient.delete('/carts/me').catchError((_) => null);
      } else {
        return;
      }
    }

    try {
      if (qty <= 0) {
        // Find cart item ID and delete
        CartItemModel? existingItem;
        try {
          existingItem = _cart?.items.firstWhere((i) => i.productId == product.id);
        } catch (_) {}
        if (existingItem != null && existingItem.id.isNotEmpty) {
          await widget.apiClient.delete('/carts/items/${existingItem.id}');
        }
      } else {
        await widget.apiClient.post('/carts/items', body: {
          'product_id': product.id,
          'quantity': qty,
        });
      }

      // Reload cart
      final updatedCart = await widget.apiClient.get('/carts/me');
      if (mounted && updatedCart != null) {
        setState(() => _cart = CartModel.fromJson(updatedCart));
      }
    } catch (e) {
      if (mounted) AppFeedback.showError(context, 'Cart error: $e');
    }
  }

  Future<void> _handleUpdateItemQty(String itemId, int newQty) async {
    try {
      if (newQty <= 0) {
        await widget.apiClient.delete('/carts/items/$itemId');
      } else {
        await widget.apiClient.patch('/carts/items/$itemId', body: {'quantity': newQty});
      }

      final updatedCart = await widget.apiClient.get('/carts/me');
      if (mounted) {
        setState(() {
          _cart = updatedCart != null ? CartModel.fromJson(updatedCart) : null;
        });
      }
    } catch (e) {
      if (mounted) AppFeedback.showError(context, 'Failed to update item: $e');
    }
  }

  Future<void> _handleClearCart() async {
    try {
      await widget.apiClient.delete('/carts/me');
      if (mounted) setState(() => _cart = null);
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);
    final cartCount = _cart?.items.length ?? 0;

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              loc.t('customer_portal'),
              style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 17, color: AppColors.textPrimary),
            ),
            Row(
              children: [
                const Icon(Icons.place_rounded, size: 12, color: AppColors.primary),
                const SizedBox(width: 3),
                Text(
                  'Bangalore Majestic • ${_nearbyVendors.length} nearby',
                  style: const TextStyle(fontSize: 11, color: AppColors.textSecondary, fontWeight: FontWeight.w500),
                ),
              ],
            ),
          ],
        ),
        actions: [
          IconButton(
            tooltip: loc.t('explore_map'),
            icon: const Icon(Icons.map_rounded, size: 21, color: AppColors.primary),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (ctx) => ExploreMapScreen(
                    apiClient: widget.apiClient,
                    nearbyVendors: _nearbyVendors,
                    categories: _categories,
                    cart: _cart,
                    initialLat: _currentLat,
                    initialLng: _currentLng,
                    onAddToCart: _handleAddToCart,
                    onViewCart: () => setState(() => _currentIndex = 1),
                  ),
                ),
              );
            },
          ),
          IconButton(
            tooltip: 'Live GPS',
            icon: const Icon(Icons.my_location_rounded, size: 20, color: AppColors.primary),
            onPressed: _fetchGPSLocation,
          ),
          IconButton(
            tooltip: loc.t('refresh'),
            icon: const Icon(Icons.refresh_rounded, size: 20, color: AppColors.textPrimary),
            onPressed: _loadCustomerData,
          ),
          PopupMenuButton<String>(
            icon: const Icon(Icons.more_vert_rounded, size: 20, color: AppColors.textPrimary),
            onSelected: (val) {
              if (val == 'switch') {
                widget.onSwitchToVendor();
              } else if (val == 'lang') {
                loc.setLanguage(loc.currentLanguage == 'en' ? 'kn' : 'en');
              } else if (val == 'logout') {
                widget.onLogout();
              }
            },
            itemBuilder: (ctx) => [
              if (widget.canSwitchToVendor)
                PopupMenuItem(
                  value: 'switch',
                  child: Row(
                    children: [
                      const Icon(Icons.storefront_rounded, size: 18, color: AppColors.primary),
                      const SizedBox(width: 8),
                      Text(loc.t('switch_to_vendor'), style: const TextStyle(fontSize: 13)),
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
      body: IndexedStack(
        index: _currentIndex,
        children: [
          CustomerDiscoverTab(
            apiClient: widget.apiClient,
            nearbyVendors: _nearbyVendors,
            categories: _categories,
            cart: _cart,
            isLoading: _isLoading,
            currentLat: _currentLat,
            currentLng: _currentLng,
            onRefresh: _loadCustomerData,
            onFetchGPS: _fetchGPSLocation,
            onAddToCart: _handleAddToCart,
            onViewCart: () => setState(() => _currentIndex = 1),
          ),
          CustomerCartTab(
            apiClient: widget.apiClient,
            cart: _cart,
            addresses: _myAddresses,
            loyalty: _loyalty,
            isLoading: _isLoading,
            onUpdateItemQty: _handleUpdateItemQty,
            onClearCart: _handleClearCart,
            onRefresh: _loadCustomerData,
            onOrderSuccess: () {
              _loadCustomerData();
              setState(() => _currentIndex = 2); // switch to orders tab
            },
            onExploreVendors: () => setState(() => _currentIndex = 0),
            onAddressAdded: _loadCustomerData,
          ),
          CustomerOrdersTab(
            apiClient: widget.apiClient,
            orders: _myOrders,
            isLoading: _isLoading,
            onRefresh: _loadCustomerData,
          ),
          CustomerLoyaltyTab(
            apiClient: widget.apiClient,
            loyalty: _loyalty,
            addresses: _myAddresses,
            onAddressAdded: _loadCustomerData,
          ),
        ],
      ),
      bottomNavigationBar: AppBottomNav(
        currentIndex: _currentIndex,
        onTap: (index) => setState(() => _currentIndex = index),
        items: [
          AppNavItem(
            icon: Icons.explore_outlined,
            activeIcon: Icons.explore_rounded,
            label: loc.t('nearby_vendors'),
          ),
          AppNavItem(
            icon: Icons.shopping_cart_outlined,
            activeIcon: Icons.shopping_cart_rounded,
            label: loc.t('cart'),
            badgeCount: cartCount,
          ),
          AppNavItem(
            icon: Icons.receipt_long_outlined,
            activeIcon: Icons.receipt_long_rounded,
            label: loc.t('my_orders'),
          ),
          AppNavItem(
            icon: Icons.stars_outlined,
            activeIcon: Icons.stars_rounded,
            label: loc.t('loyalty_rewards'),
          ),
        ],
      ),
    );
  }
}

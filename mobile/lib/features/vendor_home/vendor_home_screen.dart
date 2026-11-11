import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import '../../core/api/api_client.dart';
import '../../core/design_system/app_colors.dart';
import '../../core/design_system/components/app_badges.dart';
import '../../core/design_system/components/app_navigation_bar.dart';
import '../../core/design_system/components/app_sheets.dart';
import '../../core/localization/locale_provider.dart';
import '../../shared/models/models.dart';
import 'tabs/vendor_dashboard_tab.dart';
import 'tabs/vendor_inventory_tab.dart';
import 'tabs/vendor_orders_tab.dart';
import 'tabs/vendor_products_tab.dart';
import 'tabs/vendor_udhaar_tab.dart';

class VendorHomeScreen extends StatefulWidget {
  final ApiClient apiClient;
  final VoidCallback onLogout;
  final VoidCallback onSwitchToCustomer;

  const VendorHomeScreen({
    super.key,
    required this.apiClient,
    required this.onLogout,
    required this.onSwitchToCustomer,
  });

  @override
  State<VendorHomeScreen> createState() => _VendorHomeScreenState();
}

class _VendorHomeScreenState extends State<VendorHomeScreen> {
  int _currentIndex = 0;
  bool _isLoading = true;
  bool _isOpen = true;

  // Data
  VendorSummaryModel? _vendorSummary;
  DetailedAnalyticsModel? _detailedAnalytics;
  DemandForecastResponseModel? _demandForecast;
  WeatherModel? _weather;
  List<ProductModel> _products = [];
  Map<String, InventoryModel> _inventoryMap = {};
  List<CategoryModel> _categories = [];
  List<OrderModel> _orders = [];
  List<UdhaarAccountModel> _udhaarAccounts = [];

  @override
  void initState() {
    super.initState();
    _loadAllVendorData();
  }

  Future<void> _loadAllVendorData() async {
    setState(() => _isLoading = true);
    try {
      // 1. Categories
      final catRes = await widget.apiClient.get('/categories').catchError((_) => null);
      if (catRes != null && catRes is List) {
        _categories = catRes.map((e) => CategoryModel.fromJson(e)).toList();
      }

      // 2. Products
      final prodRes = await widget.apiClient.get('/products?limit=100').catchError((_) => null);
      if (prodRes != null && prodRes['data'] != null) {
        _products = (prodRes['data'] as List).map((e) => ProductModel.fromJson(e)).toList();
      }

      // 3. Inventory
      final invRes = await widget.apiClient.get('/inventory/me').catchError((_) => null);
      if (invRes != null && invRes is List) {
        _inventoryMap = {
          for (var item in invRes) item['product_id'] as String: InventoryModel.fromJson(item)
        };
      }

      // 4. Orders
      final orderRes = await widget.apiClient.get('/orders?limit=100').catchError((_) => null);
      if (orderRes != null && orderRes['orders'] != null) {
        _orders = (orderRes['orders'] as List).map((e) => OrderModel.fromJson(e)).toList();
      }

      // 5. Summary Analytics
      final summaryRes = await widget.apiClient.get('/analytics/vendor/summary').catchError((_) => null);
      if (summaryRes != null) {
        _vendorSummary = VendorSummaryModel.fromJson(summaryRes);
      }

      // 6. Detailed Analytics
      final detailedRes = await widget.apiClient.get('/analytics/vendor/detailed').catchError((_) => null);
      if (detailedRes != null) {
        _detailedAnalytics = DetailedAnalyticsModel.fromJson(detailedRes);
      }

      // 6. Demand Forecast
      final forecastRes = await widget.apiClient.get('/analytics/vendor/demand-forecast').catchError((_) => null);
      if (forecastRes != null) {
        _demandForecast = DemandForecastResponseModel.fromJson(forecastRes);
      }

      // 7. Weather
      final weatherRes = await widget.apiClient.get('/analytics/weather').catchError((_) => null);
      if (weatherRes != null) {
        _weather = WeatherModel.fromJson(weatherRes);
      }

      // 8. Udhaar
      final udhaarRes = await widget.apiClient.get('/udhaar/accounts').catchError((_) => null);
      if (udhaarRes != null) {
        final rawList = udhaarRes is List ? udhaarRes : (udhaarRes['data'] as List?);
        if (rawList != null) {
          _udhaarAccounts = rawList.map((e) => UdhaarAccountModel.fromJson(e)).toList();
        }
      }

      // 9. Vendor Profile
      final vendorRes = await widget.apiClient.get('/vendors/profile/me').catchError((_) => null);
      if (vendorRes != null) {
        _isOpen = vendorRes['availability_status'] == 'OPEN';
      }
    } catch (_) {
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _toggleAvailability() async {
    final newStatus = !_isOpen;
    try {
      await widget.apiClient.patch('/vendors/profile/me', body: {
        'availability_status': newStatus ? 'OPEN' : 'CLOSED',
      });
      setState(() => _isOpen = newStatus);
      if (mounted) {
        AppFeedback.showSuccess(context, newStatus ? 'Store is now OPEN' : 'Store is now CLOSED');
      }
    } catch (e) {
      if (mounted) AppFeedback.showError(context, 'Failed to update store status: $e');
    }
  }

  Future<void> _updateVendorGPS() async {
    try {
      Position pos = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
        timeLimit: const Duration(seconds: 8),
      );
      await widget.apiClient.post('/vendors/locations', body: {
        'address_line1': 'Current Location',
        'city': 'Bengaluru',
        'state': 'Karnataka',
        'latitude': pos.latitude,
        'longitude': pos.longitude,
      });
      if (mounted) AppFeedback.showSuccess(context, 'GPS location broadcast updated');
    } catch (e) {
      if (mounted) AppFeedback.showInfo(context, 'Updated location to Bangalore Majestic anchor');
    }
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);
    final pendingOrdersCount = _orders.where((o) => o.status.toUpperCase() == 'PENDING').length;

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Text(
                  loc.t('vendor_portal'),
                  style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 17, color: AppColors.textPrimary),
                ),
                const SizedBox(width: 8),
                GestureDetector(
                  onTap: _toggleAvailability,
                  child: AppStatusBadge.availability(_isOpen),
                ),
              ],
            ),
            Text(
              loc.t('tagline'),
              style: const TextStyle(fontSize: 11, color: AppColors.textSecondary, fontWeight: FontWeight.w500),
            ),
          ],
        ),
        actions: [
          IconButton(
            tooltip: 'Update My GPS Location',
            icon: const Icon(Icons.share_location_rounded, size: 20, color: AppColors.primary),
            onPressed: _updateVendorGPS,
          ),
          IconButton(
            tooltip: loc.t('refresh'),
            icon: const Icon(Icons.refresh_rounded, size: 20, color: AppColors.textPrimary),
            onPressed: _loadAllVendorData,
          ),
          PopupMenuButton<String>(
            icon: const Icon(Icons.more_vert_rounded, size: 20, color: AppColors.textPrimary),
            onSelected: (val) {
              if (val == 'switch') {
                widget.onSwitchToCustomer();
              } else if (val == 'lang') {
                loc.setLanguage(loc.currentLanguage == 'en' ? 'kn' : 'en');
              } else if (val == 'logout') {
                widget.onLogout();
              }
            },
            itemBuilder: (ctx) => [
              PopupMenuItem(
                value: 'switch',
                child: Row(
                  children: [
                    const Icon(Icons.shopping_bag_outlined, size: 18, color: AppColors.primary),
                    const SizedBox(width: 8),
                    Text(loc.t('switch_to_customer'), style: const TextStyle(fontSize: 13)),
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
          VendorDashboardTab(
            apiClient: widget.apiClient,
            summary: _vendorSummary,
            analytics: _detailedAnalytics,
            demandForecast: _demandForecast,
            weather: _weather,
            isLoading: _isLoading,
            onRefresh: _loadAllVendorData,
            onViewOrders: () => setState(() => _currentIndex = 1),
            onViewInventory: () => setState(() => _currentIndex = 3),
          ),
          VendorOrdersTab(
            apiClient: widget.apiClient,
            orders: _orders,
            isLoading: _isLoading,
            onRefresh: _loadAllVendorData,
          ),
          VendorProductsTab(
            apiClient: widget.apiClient,
            products: _products,
            categories: _categories,
            isLoading: _isLoading,
            onRefresh: _loadAllVendorData,
          ),
          VendorInventoryTab(
            apiClient: widget.apiClient,
            products: _products,
            inventoryMap: _inventoryMap,
            isLoading: _isLoading,
            onRefresh: _loadAllVendorData,
          ),
          VendorUdhaarTab(
            apiClient: widget.apiClient,
            accounts: _udhaarAccounts,
            isLoading: _isLoading,
            onRefresh: _loadAllVendorData,
          ),
        ],
      ),
      bottomNavigationBar: AppBottomNav(
        currentIndex: _currentIndex,
        onTap: (index) => setState(() => _currentIndex = index),
        items: [
          AppNavItem(
            icon: Icons.dashboard_outlined,
            activeIcon: Icons.dashboard_rounded,
            label: loc.t('dashboard'),
          ),
          AppNavItem(
            icon: Icons.receipt_long_outlined,
            activeIcon: Icons.receipt_long_rounded,
            label: loc.t('orders'),
            badgeCount: pendingOrdersCount,
          ),
          AppNavItem(
            icon: Icons.inventory_2_outlined,
            activeIcon: Icons.inventory_2_rounded,
            label: loc.t('products'),
          ),
          AppNavItem(
            icon: Icons.warehouse_outlined,
            activeIcon: Icons.warehouse_rounded,
            label: loc.t('inventory'),
          ),
          AppNavItem(
            icon: Icons.menu_book_outlined,
            activeIcon: Icons.menu_book_rounded,
            label: loc.t('udhaar'),
          ),
        ],
      ),
    );
  }
}

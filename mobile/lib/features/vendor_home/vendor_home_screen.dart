import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import '../../core/api/api_client.dart';
import '../../core/localization/locale_provider.dart';
import '../../shared/models/models.dart';

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

class _VendorHomeScreenState extends State<VendorHomeScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  bool _isLoading = true;

  // Data
  DetailedAnalyticsModel? _detailedAnalytics;
  DemandForecastResponseModel? _demandForecast;
  WeatherModel? _weather;
  List<ProductModel> _products = [];
  Map<String, InventoryModel> _inventoryMap = {};
  List<CategoryModel> _categories = [];
  List<OrderModel> _orders = [];
  List<UdhaarAccountModel> _udhaarAccounts = [];
  List<ExpenseModel> _expenses = [];

  // Filter / Search state
  String _productSearchQuery = '';
  String? _selectedCategoryFilter;
  String _orderStatusFilter = 'ALL';

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 5, vsync: this);
    _loadAllVendorData();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
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

      // 3. Inventory for all products
      final invRes = await widget.apiClient.get('/inventory/me').catchError((_) => null);
      if (invRes != null && invRes is List) {
        _inventoryMap = {
          for (var item in invRes)
            item['product_id'] as String: InventoryModel.fromJson(item)
        };
      }

      // 4. Orders
      final orderRes = await widget.apiClient.get('/orders?limit=100').catchError((_) => null);
      if (orderRes != null && orderRes['orders'] != null) {
        _orders = (orderRes['orders'] as List).map((e) => OrderModel.fromJson(e)).toList();
      }

      // 5. Detailed Analytics
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

      // 8. Udhaar accounts
      final udhaarRes = await widget.apiClient.get('/udhaar/accounts').catchError((_) => null);
      if (udhaarRes != null && udhaarRes['data'] != null) {
        _udhaarAccounts = (udhaarRes['data'] as List).map((e) => UdhaarAccountModel.fromJson(e)).toList();
      }

      // 9. Expenses
      final expRes = await widget.apiClient.get('/analytics/expenses').catchError((_) => null);
      if (expRes != null && expRes is List) {
        _expenses = expRes.map((e) => ExpenseModel.fromJson(e)).toList();
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  // --- Location capture ---
  Future<void> _handleCaptureGPSLocation() async {
    final loc = LocaleProvider.of(context);
    try {
      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
        if (permission == LocationPermission.denied) {
          if (!mounted) return;
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Location permission denied')),
          );
          return;
        }
      }

      final position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
      ).catchError((_) => Position(
        latitude: 12.9767,
        longitude: 77.5713,
        timestamp: DateTime.now(),
        accuracy: 10,
        altitude: 900,
        altitudeAccuracy: 1,
        heading: 0,
        headingAccuracy: 1,
        speed: 0,
        speedAccuracy: 1,
      ));

      await widget.apiClient.post('/vendors/locations', body: {
        'address_line1': 'Current GPS Location (Street Vendor)',
        'city': 'Bengaluru',
        'state': 'Karnataka',
        'latitude': position.latitude,
        'longitude': position.longitude,
        'is_primary': true,
      });

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('${loc.t('vendor_location')} updated: ${position.latitude.toStringAsFixed(4)}, ${position.longitude.toStringAsFixed(4)}'),
          backgroundColor: Colors.green,
        ),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Location update error: $e'), backgroundColor: Colors.red),
      );
    }
  }

  // --- Order Lifecycle Actions ---
  Future<void> _handleOrderAction(String orderId, String action, {String? reason}) async {
    try {
      await widget.apiClient.post('/orders/$orderId/$action', body: {
        if (reason != null && reason.isNotEmpty) 'reason': reason,
      });
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Order updated: $action'), backgroundColor: Colors.green),
      );
      _loadAllVendorData();
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Action failed: $e'), backgroundColor: Colors.red),
      );
    }
  }

  // --- Cancellation Decision ---
  Future<void> _handleCancellationDecision(String orderId, String decision) async {
    try {
      await widget.apiClient.post('/orders/$orderId/decide-cancellation', body: {
        'decision': decision,
      });
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Cancellation $decision'), backgroundColor: Colors.green),
      );
      _loadAllVendorData();
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Decision failed: $e'), backgroundColor: Colors.red),
      );
    }
  }

  // --- Inventory Stock Adjustment ---
  Future<void> _handleAdjustStock(String productId, int delta, String reason) async {
    try {
      await widget.apiClient.post('/inventory/adjust', body: {
        'product_id': productId,
        'qty_delta': delta,
        'reason': reason.isNotEmpty ? reason : 'Stock adjustment',
      });
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Stock updated by $delta'), backgroundColor: Colors.green),
      );
      _loadAllVendorData();
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Failed: $e'), backgroundColor: Colors.red),
      );
    }
  }

  // --- Product Creation / Update Dialog ---
  void _showProductDialog({ProductModel? product}) {
    final loc = LocaleProvider.of(context);
    final nameCtrl = TextEditingController(text: product?.name ?? '');
    final descCtrl = TextEditingController(text: product?.description ?? '');
    final priceCtrl = TextEditingController(
      text: product != null ? (product.pricePaise / 100).toStringAsFixed(2) : '',
    );
    final costPriceCtrl = TextEditingController(
      text: product != null ? (product.costPricePaise / 100).toStringAsFixed(2) : '',
    );
    String? categoryId = product?.categoryId ?? (_categories.isNotEmpty ? _categories.first.id : null);
    bool active = product?.active ?? true;

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) => AlertDialog(
          title: Text(product == null ? loc.t('add_product') : loc.t('edit_product')),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: nameCtrl,
                  decoration: InputDecoration(labelText: loc.t('product_name')),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: descCtrl,
                  decoration: InputDecoration(labelText: loc.t('product_desc')),
                ),
                const SizedBox(height: 12),
                if (_categories.isNotEmpty)
                  DropdownButtonFormField<String>(
                    value: categoryId,
                    decoration: InputDecoration(labelText: loc.t('category')),
                    items: _categories.map((c) {
                      return DropdownMenuItem(value: c.id, child: Text(c.name));
                    }).toList(),
                    onChanged: (val) => setDialogState(() => categoryId = val),
                  ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: priceCtrl,
                        keyboardType: TextInputType.number,
                        decoration: InputDecoration(
                          labelText: loc.t('selling_price'),
                          prefixText: '₹ ',
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: TextField(
                        controller: costPriceCtrl,
                        keyboardType: TextInputType.number,
                        decoration: InputDecoration(
                          labelText: loc.t('cost_price'),
                          prefixText: '₹ ',
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                SwitchListTile(
                  title: Text(loc.t('active')),
                  value: active,
                  onChanged: (val) => setDialogState(() => active = val),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: Text(loc.t('cancel')),
            ),
            ElevatedButton(
              onPressed: () async {
                final name = nameCtrl.text.trim();
                final priceFloat = double.tryParse(priceCtrl.text) ?? 0.0;
                final costPriceFloat = double.tryParse(costPriceCtrl.text) ?? 0.0;
                if (name.isEmpty || priceFloat <= 0) return;

                final pricePaise = (priceFloat * 100).toInt();
                final costPricePaise = (costPriceFloat * 100).toInt();

                try {
                  if (product == null) {
                    await widget.apiClient.post('/products', body: {
                      'name': name,
                      'description': descCtrl.text.trim(),
                      'category_id': categoryId,
                      'price_paise': pricePaise,
                      'cost_price_paise': costPricePaise,
                      'currency': 'INR',
                    });
                  } else {
                    await widget.apiClient.patch('/products/${product.id}', body: {
                      'name': name,
                      'description': descCtrl.text.trim(),
                      'category_id': categoryId,
                      'price_paise': pricePaise,
                      'cost_price_paise': costPricePaise,
                      'active': active,
                    });
                  }
                  if (mounted) {
                    Navigator.pop(ctx);
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text(loc.t('success')), backgroundColor: Colors.green),
                    );
                    _loadAllVendorData();
                  }
                } catch (e) {
                  if (mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
                    );
                  }
                }
              },
              child: Text(loc.t('save')),
            ),
          ],
        ),
      ),
    );
  }

  // --- Add Expense Dialog ---
  void _showAddExpenseDialog() {
    final loc = LocaleProvider.of(context);
    final titleCtrl = TextEditingController();
    final amountCtrl = TextEditingController();
    String category = 'OPERATIONAL';

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(loc.t('add_expense')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: titleCtrl,
              decoration: InputDecoration(labelText: loc.t('expense_title')),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: amountCtrl,
              keyboardType: TextInputType.number,
              decoration: InputDecoration(
                labelText: loc.t('expense_amount'),
                prefixText: '₹ ',
              ),
            ),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: Text(loc.t('cancel'))),
          ElevatedButton(
            onPressed: () async {
              final amountFloat = double.tryParse(amountCtrl.text) ?? 0.0;
              if (titleCtrl.text.isEmpty || amountFloat <= 0) return;
              try {
                await widget.apiClient.post('/analytics/expenses', body: {
                  'title': titleCtrl.text.trim(),
                  'amount_paise': (amountFloat * 100).toInt(),
                  'category': category,
                });
                if (mounted) {
                  Navigator.pop(ctx);
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text(loc.t('success')), backgroundColor: Colors.green),
                  );
                  _loadAllVendorData();
                }
              } catch (e) {
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
                  );
                }
              }
            },
            child: Text(loc.t('save')),
          ),
        ],
      ),
    );
  }

  // --- Add Udhaar Entry Dialog ---
  void _showUdhaarEntryDialog(UdhaarAccountModel account) {
    final loc = LocaleProvider.of(context);
    final amountCtrl = TextEditingController();
    final noteCtrl = TextEditingController();
    String entryType = 'CREDIT';

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) => AlertDialog(
          title: Text('${loc.t('udhaar')}: ${account.id.substring(0, 8)}...'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              SegmentedButton<String>(
                segments: [
                  ButtonSegment(value: 'CREDIT', label: Text(loc.t('add_credit'))),
                  ButtonSegment(value: 'REPAYMENT', label: Text(loc.t('record_repayment'))),
                ],
                selected: {entryType},
                onSelectionChanged: (set) => setDialogState(() => entryType = set.first),
              ),
              const SizedBox(height: 16),
              TextField(
                controller: amountCtrl,
                keyboardType: TextInputType.number,
                decoration: InputDecoration(labelText: loc.t('total'), prefixText: '₹ '),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: noteCtrl,
                decoration: InputDecoration(labelText: loc.t('reason')),
              ),
            ],
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx), child: Text(loc.t('cancel'))),
            ElevatedButton(
              onPressed: () async {
                final amt = double.tryParse(amountCtrl.text) ?? 0.0;
                if (amt <= 0) return;
                try {
                  await widget.apiClient.post('/udhaar/accounts/${account.id}/entries', body: {
                    'entry_type': entryType,
                    'amount_paise': (amt * 100).toInt(),
                    'note': noteCtrl.text.trim(),
                  });
                  if (mounted) {
                    Navigator.pop(ctx);
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text(loc.t('success')), backgroundColor: Colors.green),
                    );
                    _loadAllVendorData();
                  }
                } catch (e) {
                  if (mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
                    );
                  }
                }
              },
              child: Text(loc.t('submit')),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Text(loc.t('vendor_portal'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.green.shade100,
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    loc.t('open'),
                    style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Colors.green.shade800),
                  ),
                ),
              ],
            ),
            Text(loc.t('tagline'), style: const TextStyle(fontSize: 11, color: Color(0xFF64748B))),
          ],
        ),
        actions: [
          IconButton(
            tooltip: loc.t('update_my_location'),
            icon: const Icon(Icons.my_location_rounded, color: Color(0xFF2563EB)),
            onPressed: _handleCaptureGPSLocation,
          ),
          IconButton(
            tooltip: loc.t('refresh'),
            icon: const Icon(Icons.refresh),
            onPressed: _loadAllVendorData,
          ),
          PopupMenuButton<String>(
            icon: const Icon(Icons.more_vert),
            onSelected: (val) {
              if (val == 'switch') {
                widget.onSwitchToCustomer();
              } else if (val == 'lang') {
                final newLang = loc.currentLanguage == 'en' ? 'kn' : 'en';
                loc.setLanguage(newLang);
                LanguageController.saveLanguage(newLang, apiClient: widget.apiClient);
              } else if (val == 'logout') {
                widget.onLogout();
              }
            },
            itemBuilder: (ctx) => [
              PopupMenuItem(
                value: 'switch',
                child: Row(
                  children: [
                    const Icon(Icons.shopping_bag_outlined, size: 20),
                    const SizedBox(width: 8),
                    Text(loc.t('switch_to_customer')),
                  ],
                ),
              ),
              PopupMenuItem(
                value: 'lang',
                child: Row(
                  children: [
                    const Icon(Icons.language, size: 20),
                    const SizedBox(width: 8),
                    Text(loc.currentLanguage == 'en' ? 'ಕನ್ನಡಕ್ಕೆ ಬದಲಾಯಿಸಿ' : 'Switch to English'),
                  ],
                ),
              ),
              const PopupMenuDivider(),
              PopupMenuItem(
                value: 'logout',
                child: Row(
                  children: [
                    const Icon(Icons.logout, size: 20, color: Colors.red),
                    const SizedBox(width: 8),
                    Text(loc.t('logout'), style: const TextStyle(color: Colors.red)),
                  ],
                ),
              ),
            ],
          ),
        ],
        bottom: TabBar(
          controller: _tabController,
          isScrollable: true,
          labelColor: const Color(0xFF1E40AF),
          unselectedLabelColor: const Color(0xFF64748B),
          indicatorColor: const Color(0xFF1E40AF),
          tabs: [
            Tab(icon: const Icon(Icons.dashboard_rounded), text: loc.t('dashboard')),
            Tab(icon: const Icon(Icons.inventory_2_rounded), text: loc.t('products')),
            Tab(
              icon: Badge(
                label: Text('${_orders.where((o) => o.status == 'PENDING').length}'),
                isLabelVisible: _orders.any((o) => o.status == 'PENDING'),
                child: const Icon(Icons.receipt_long_rounded),
              ),
              text: loc.t('orders'),
            ),
            Tab(icon: const Icon(Icons.account_balance_wallet_rounded), text: loc.t('udhaar')),
            Tab(icon: const Icon(Icons.auto_graph_rounded), text: loc.t('forecast')),
          ],
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : TabBarView(
              controller: _tabController,
              children: [
                _buildDashboardTab(loc),
                _buildProductsTab(loc),
                _buildOrdersTab(loc),
                _buildUdhaarAndFinanceTab(loc),
                _buildDemandForecastTab(loc),
              ],
            ),
      floatingActionButton: _tabController.index == 1
          ? FloatingActionButton.extended(
              onPressed: () => _showProductDialog(),
              icon: const Icon(Icons.add),
              label: Text(loc.t('add_product')),
            )
          : null,
    );
  }

  // --- TAB 1: DASHBOARD ---
  Widget _buildDashboardTab(LocaleProvider loc) {
    String formatINR(int paise) => '₹${(paise / 100).toStringAsFixed(2)}';

    return RefreshIndicator(
      onRefresh: _loadAllVendorData,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Weather Planning Card
          if (_weather != null)
            Card(
              elevation: 0,
              color: const Color(0xFFEFF6FF),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
                side: const BorderSide(color: Color(0xFFBFDBFE)),
              ),
              child: Padding(
                padding: const EdgeInsets.all(14),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.wb_sunny_rounded, color: Color(0xFF2563EB), size: 28),
                        const SizedBox(width: 10),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                '${_weather!.city} • ${_weather!.temperatureCelsius.toStringAsFixed(1)}°C (${_weather!.condition})',
                                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF1E3A8A)),
                              ),
                              Text(
                                '${_weather!.conditionDescription} • ${loc.t('rain_chance')}: ${_weather!.rainProbability}%',
                                style: const TextStyle(fontSize: 12, color: Color(0xFF3B82F6)),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                    if (_weather!.vendorPlanningTips.isNotEmpty) ...[
                      const SizedBox(height: 10),
                      const Divider(color: Color(0xFFDBEAFE)),
                      const SizedBox(height: 4),
                      Text(
                        loc.t('planning_tips'),
                        style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Color(0xFF1E40AF)),
                      ),
                      const SizedBox(height: 4),
                      for (final tip in _weather!.vendorPlanningTips.take(2))
                        Padding(
                          padding: const EdgeInsets.only(top: 2.0),
                          child: Text('• $tip', style: const TextStyle(fontSize: 11, color: Color(0xFF1E293B))),
                        ),
                    ],
                  ],
                ),
              ),
            ),
          const SizedBox(height: 12),

          // High-level Sales KPI Cards
          GridView.count(
            crossAxisCount: 2,
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            crossAxisSpacing: 10,
            mainAxisSpacing: 10,
            childAspectRatio: 1.5,
            children: [
              _buildKpiCard(
                title: loc.t('daily_sales'),
                value: formatINR(_detailedAnalytics?.dailySalesPaise ?? 0),
                icon: Icons.today_rounded,
                color: Colors.blue,
              ),
              _buildKpiCard(
                title: loc.t('weekly_sales'),
                value: formatINR(_detailedAnalytics?.weeklySalesPaise ?? 0),
                icon: Icons.calendar_view_week_rounded,
                color: Colors.indigo,
              ),
              _buildKpiCard(
                title: loc.t('gross_profit'),
                value: formatINR(_detailedAnalytics?.estimatedGrossProfitPaise ?? 0),
                icon: Icons.trending_up_rounded,
                color: Colors.green,
                subtitle: 'Selling - Cost Price',
              ),
              _buildKpiCard(
                title: loc.t('net_profit'),
                value: formatINR(_detailedAnalytics?.estimatedNetProfitPaise ?? 0),
                icon: Icons.account_balance_rounded,
                color: Colors.teal,
                subtitle: 'Gross - Expenses',
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Low Stock Alerts Card
          if (_detailedAnalytics != null && _detailedAnalytics!.lowStockProducts.isNotEmpty)
            Card(
              color: const Color(0xFFFFFBEB),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
                side: const BorderSide(color: Color(0xFFFDE68A)),
              ),
              child: Padding(
                padding: const EdgeInsets.all(14),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.warning_amber_rounded, color: Color(0xFFD97706), size: 22),
                        const SizedBox(width: 8),
                        Text(
                          loc.t('low_stock_alert'),
                          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF92400E)),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    for (final item in _detailedAnalytics!.lowStockProducts.take(3))
                      Padding(
                        padding: const EdgeInsets.symmetric(vertical: 2.0),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(item['name'] ?? '', style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500)),
                            Text('${item['available_qty']} left', style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Colors.red)),
                          ],
                        ),
                      ),
                  ],
                ),
              ),
            ),
          const SizedBox(height: 16),

          // Quick Action GPS & Location
          Card(
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            child: ListTile(
              leading: const CircleAvatar(
                backgroundColor: Color(0xFFEFF6FF),
                child: Icon(Icons.navigation_rounded, color: Color(0xFF2563EB)),
              ),
              title: Text(loc.t('update_my_location'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
              subtitle: const Text('Broadcast current mobile GPS coordinate to nearby customers', style: TextStyle(fontSize: 11)),
              trailing: ElevatedButton(
                onPressed: _handleCaptureGPSLocation,
                style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF2563EB), foregroundColor: Colors.white),
                child: const Text('Update'),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildKpiCard({required String title, required String value, required IconData icon, required Color color, String? subtitle}) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: Colors.grey.shade200),
      ),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Row(
              children: [
                Icon(icon, color: color, size: 18),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    title,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(fontSize: 11, color: Colors.grey.shade600, fontWeight: FontWeight.w500),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 6),
            Text(
              value,
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Color(0xFF0F172A)),
            ),
            if (subtitle != null) ...[
              const SizedBox(height: 2),
              Text(subtitle, style: TextStyle(fontSize: 9, color: Colors.grey.shade500)),
            ],
          ],
        ),
      ),
    );
  }

  // --- TAB 2: PRODUCTS & STOCK ---
  Widget _buildProductsTab(LocaleProvider loc) {
    var filtered = _products.where((p) {
      final matchesSearch = p.name.toLowerCase().contains(_productSearchQuery.toLowerCase());
      final matchesCategory = _selectedCategoryFilter == null || p.categoryId == _selectedCategoryFilter;
      return matchesSearch && matchesCategory;
    }).toList();

    return Column(
      children: [
        // Search & Category Filter Bar
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: Column(
            children: [
              TextField(
                decoration: InputDecoration(
                  hintText: loc.t('search'),
                  prefixIcon: const Icon(Icons.search),
                  contentPadding: const EdgeInsets.symmetric(vertical: 0, horizontal: 16),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
                ),
                onChanged: (val) => setState(() => _productSearchQuery = val),
              ),
              if (_categories.isNotEmpty) ...[
                const SizedBox(height: 8),
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      ChoiceChip(
                        label: Text(loc.t('all')),
                        selected: _selectedCategoryFilter == null,
                        onSelected: (_) => setState(() => _selectedCategoryFilter = null),
                      ),
                      for (final cat in _categories)
                        Padding(
                          padding: const EdgeInsets.only(left: 6.0),
                          child: ChoiceChip(
                            label: Text(cat.name),
                            selected: _selectedCategoryFilter == cat.id,
                            onSelected: (_) => setState(() => _selectedCategoryFilter = cat.id),
                          ),
                        ),
                    ],
                  ),
                ),
              ],
            ],
          ),
        ),

        // Product Cards List
        Expanded(
          child: filtered.isEmpty
              ? Center(child: Text(loc.t('no_products')))
              : ListView.builder(
                  padding: const EdgeInsets.all(12),
                  itemCount: filtered.length,
                  itemBuilder: (ctx, i) {
                    final p = filtered[i];
                    final inv = _inventoryMap[p.id];
                    final availableStock = inv?.availableQty ?? 0;
                    final isLowStock = availableStock <= 5;

                    return Card(
                      margin: const EdgeInsets.only(bottom: 10),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                        side: BorderSide(
                          color: isLowStock ? Colors.orange.shade300 : Colors.grey.shade200,
                        ),
                      ),
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Container(
                                  width: 44,
                                  height: 44,
                                  decoration: BoxDecoration(
                                    color: Colors.blue.shade50,
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: const Icon(Icons.shopping_bag_outlined, color: Color(0xFF2563EB)),
                                ),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        p.name,
                                        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
                                      ),
                                      if (p.description != null && p.description!.isNotEmpty)
                                        Text(
                                          p.description!,
                                          maxLines: 1,
                                          overflow: TextOverflow.ellipsis,
                                          style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
                                        ),
                                      const SizedBox(height: 4),
                                      Row(
                                        children: [
                                          Text(
                                            'Selling: ${p.formattedPrice}',
                                            style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Colors.blue),
                                          ),
                                          const SizedBox(width: 8),
                                          Text(
                                            'Cost: ${p.formattedCostPrice}',
                                            style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
                                          ),
                                          const SizedBox(width: 8),
                                          Text(
                                            'Margin: ${p.marginPercent.toStringAsFixed(0)}%',
                                            style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: Colors.green.shade700),
                                          ),
                                        ],
                                      ),
                                    ],
                                  ),
                                ),
                                IconButton(
                                  icon: const Icon(Icons.edit_outlined, size: 20),
                                  onPressed: () => _showProductDialog(product: p),
                                ),
                              ],
                            ),
                            const Divider(height: 20),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Row(
                                  children: [
                                    Icon(
                                      isLowStock ? Icons.warning_rounded : Icons.check_circle_rounded,
                                      size: 16,
                                      color: isLowStock ? Colors.orange : Colors.green,
                                    ),
                                    const SizedBox(width: 6),
                                    Text(
                                      '${loc.t('available')}: $availableStock',
                                      style: TextStyle(
                                        fontSize: 13,
                                        fontWeight: FontWeight.bold,
                                        color: isLowStock ? Colors.orange.shade800 : Colors.green.shade800,
                                      ),
                                    ),
                                  ],
                                ),
                                Row(
                                  children: [
                                    IconButton.outlined(
                                      icon: const Icon(Icons.remove, size: 16),
                                      onPressed: availableStock > 0 ? () => _handleAdjustStock(p.id, -1, 'Quick reduction') : null,
                                    ),
                                    const SizedBox(width: 4),
                                    IconButton.outlined(
                                      icon: const Icon(Icons.add, size: 16),
                                      onPressed: () => _handleAdjustStock(p.id, 5, 'Quick replenishment (+5)'),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }

  // --- TAB 3: ORDERS ---
  Widget _buildOrdersTab(LocaleProvider loc) {
    var filtered = _orders.where((o) {
      if (_orderStatusFilter == 'ALL') return true;
      return o.status == _orderStatusFilter;
    }).toList();

    return Column(
      children: [
        // Status Filter Chips
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: Row(
            children: [
              for (final st in ['ALL', 'PENDING', 'ACCEPTED', 'PREPARING', 'READY_FOR_PICKUP', 'OUT_FOR_DELIVERY', 'COMPLETED', 'CANCELLATION_REQUESTED'])
                Padding(
                  padding: const EdgeInsets.only(right: 6.0),
                  child: FilterChip(
                    label: Text(st),
                    selected: _orderStatusFilter == st,
                    onSelected: (_) => setState(() => _orderStatusFilter = st),
                  ),
                ),
            ],
          ),
        ),

        Expanded(
          child: filtered.isEmpty
              ? Center(child: Text(loc.t('no_orders')))
              : ListView.builder(
                  padding: const EdgeInsets.all(12),
                  itemCount: filtered.length,
                  itemBuilder: (ctx, i) {
                    final order = filtered[i];
                    return Card(
                      margin: const EdgeInsets.only(bottom: 12),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      child: Padding(
                        padding: const EdgeInsets.all(14),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'Order #${order.id.substring(0, 8)}',
                                  style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                                ),
                                _buildStatusBadge(order.status),
                              ],
                            ),
                            const SizedBox(height: 6),
                            Text(
                              'Fulfilment: ${order.fulfilmentType} • ${order.formattedTotal}',
                              style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Color(0xFF1E293B)),
                            ),
                            if (order.deliveryAddressSnapshot != null)
                              Padding(
                                padding: const EdgeInsets.only(top: 4.0),
                                child: Text(
                                  'Delivery: ${order.deliveryAddressSnapshot}',
                                  style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
                                ),
                              ),
                            const Divider(height: 20),

                            // Action buttons based on status
                            if (order.status == 'PENDING')
                              Row(
                                children: [
                                  Expanded(
                                    child: OutlinedButton(
                                      onPressed: () => _handleOrderAction(order.id, 'reject', reason: 'Out of stock'),
                                      child: Text(loc.t('reject_order'), style: const TextStyle(color: Colors.red)),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  Expanded(
                                    child: ElevatedButton(
                                      onPressed: () => _handleOrderAction(order.id, 'accept'),
                                      style: ElevatedButton.styleFrom(backgroundColor: Colors.green, foregroundColor: Colors.white),
                                      child: Text(loc.t('accept_order')),
                                    ),
                                  ),
                                ],
                              )
                            else if (order.status == 'ACCEPTED')
                              SizedBox(
                                width: double.infinity,
                                child: ElevatedButton(
                                  onPressed: () => _handleOrderAction(order.id, 'prepare'),
                                  style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF2563EB), foregroundColor: Colors.white),
                                  child: Text(loc.t('prepare_order')),
                                ),
                              )
                            else if (order.status == 'PREPARING')
                              Row(
                                children: [
                                  if (order.fulfilmentType == 'PICKUP')
                                    Expanded(
                                      child: ElevatedButton(
                                        onPressed: () => _handleOrderAction(order.id, 'ready'),
                                        child: Text(loc.t('ready_pickup')),
                                      ),
                                    )
                                  else
                                    Expanded(
                                      child: ElevatedButton(
                                        onPressed: () => _handleOrderAction(order.id, 'out-for-delivery'),
                                        child: Text(loc.t('out_delivery')),
                                      ),
                                    ),
                                ],
                              )
                            else if (order.status == 'READY_FOR_PICKUP' || order.status == 'OUT_FOR_DELIVERY')
                              SizedBox(
                                width: double.infinity,
                                child: ElevatedButton(
                                  onPressed: () => _handleOrderAction(order.id, 'complete'),
                                  style: ElevatedButton.styleFrom(backgroundColor: Colors.green.shade700, foregroundColor: Colors.white),
                                  child: Text(loc.t('complete_order')),
                                ),
                              )
                            else if (order.status == 'CANCELLATION_REQUESTED')
                              Row(
                                children: [
                                  Expanded(
                                    child: OutlinedButton(
                                      onPressed: () => _handleCancellationDecision(order.id, 'DECLINED'),
                                      child: Text(loc.t('cancel_declined')),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  Expanded(
                                    child: ElevatedButton(
                                      onPressed: () => _handleCancellationDecision(order.id, 'APPROVED'),
                                      style: ElevatedButton.styleFrom(backgroundColor: Colors.red, foregroundColor: Colors.white),
                                      child: Text(loc.t('cancel_approved')),
                                    ),
                                  ),
                                ],
                              ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }

  // --- TAB 4: UDHAAR & FINANCE ---
  Widget _buildUdhaarAndFinanceTab(LocaleProvider loc) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        // Udhaar Section Header
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(loc.t('udhaar'), style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            Text('${_udhaarAccounts.length} accounts', style: TextStyle(color: Colors.grey.shade600, fontSize: 13)),
          ],
        ),
        const SizedBox(height: 8),

        if (_udhaarAccounts.isEmpty)
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 16.0),
            child: Text('No customer Udhaar accounts found.'),
          )
        else
          for (final acc in _udhaarAccounts)
            Card(
              margin: const EdgeInsets.only(bottom: 8),
              child: ListTile(
                title: Text('Account: ${acc.id.substring(0, 8)}...'),
                subtitle: Text('Status: ${acc.status}'),
                trailing: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      acc.formattedBalance,
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: acc.balancePaise > 0 ? Colors.red : Colors.green,
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.add_circle_outline, color: Color(0xFF2563EB)),
                      onPressed: () => _showUdhaarEntryDialog(acc),
                    ),
                  ],
                ),
              ),
            ),
        const SizedBox(height: 20),

        // Expense Tracking Section
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(loc.t('expenses'), style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            ElevatedButton.icon(
              onPressed: _showAddExpenseDialog,
              icon: const Icon(Icons.add, size: 16),
              label: Text(loc.t('add_expense')),
            ),
          ],
        ),
        const SizedBox(height: 8),

        if (_expenses.isEmpty)
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 16.0),
            child: Text('No operational expenses recorded.'),
          )
        else
          for (final exp in _expenses)
            Card(
              margin: const EdgeInsets.only(bottom: 8),
              child: ListTile(
                title: Text(exp.title),
                subtitle: Text('${exp.category} • ${exp.expenseDate}'),
                trailing: Text(
                  exp.formattedAmount,
                  style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.red),
                ),
              ),
            ),
      ],
    );
  }

  // --- TAB 5: AI DEMAND PREDICTION ---
  Widget _buildDemandForecastTab(LocaleProvider loc) {
    if (_demandForecast == null) {
      return const Center(child: Text('Generating demand prediction...'));
    }

    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        // AI Model Insight Header Card
        Card(
          color: _demandForecast!.hasSufficientHistory ? const Color(0xFFF0FDF4) : const Color(0xFFFFFBEB),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: BorderSide(
              color: _demandForecast!.hasSufficientHistory ? const Color(0xFFBBF7D0) : const Color(0xFFFDE68A),
            ),
          ),
          child: Padding(
            padding: const EdgeInsets.all(14),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(
                  _demandForecast!.hasSufficientHistory ? Icons.auto_awesome : Icons.info_outline,
                  color: _demandForecast!.hasSufficientHistory ? Colors.green : Colors.amber.shade800,
                  size: 24,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        loc.t('demand_prediction'),
                        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        _demandForecast!.generalInsight,
                        style: TextStyle(fontSize: 12, color: Colors.grey.shade700),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 16),

        // Per-product forecast items
        for (final fc in _demandForecast!.forecasts)
          Card(
            margin: const EdgeInsets.only(bottom: 12),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            child: Padding(
              padding: const EdgeInsets.all(14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        fc.productName,
                        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
                      ),
                      _buildConfidenceBadge(fc.confidence),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      _buildForecastStat(loc.t('on_hand'), '${fc.currentStock}'),
                      _buildForecastStat(loc.t('predicted_demand'), '${fc.predicted7DayDemand}'),
                      _buildForecastStat(
                        loc.t('suggested_stock'),
                        '+${fc.suggestedRestockQty}',
                        highlight: fc.suggestedRestockQty > 0,
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    fc.explanation,
                    style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
                  ),
                ],
              ),
            ),
          ),
      ],
    );
  }

  Widget _buildForecastStat(String label, String value, {bool highlight = false}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: TextStyle(fontSize: 11, color: Colors.grey.shade600)),
        const SizedBox(height: 2),
        Text(
          value,
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.bold,
            color: highlight ? Colors.orange.shade800 : const Color(0xFF0F172A),
          ),
        ),
      ],
    );
  }

  Widget _buildConfidenceBadge(String confidence) {
    Color bg = Colors.grey.shade200;
    Color fg = Colors.grey.shade800;
    if (confidence == 'HIGH') {
      bg = Colors.green.shade100;
      fg = Colors.green.shade800;
    } else if (confidence == 'MEDIUM') {
      bg = Colors.amber.shade100;
      fg = Colors.amber.shade800;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(color: bg, borderRadius: BorderRadius.circular(6)),
      child: Text(confidence, style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: fg)),
    );
  }

  Widget _buildStatusBadge(String status) {
    Color bg = Colors.blue.shade100;
    Color fg = Colors.blue.shade800;

    if (status == 'COMPLETED') {
      bg = Colors.green.shade100;
      fg = Colors.green.shade800;
    } else if (status == 'PENDING') {
      bg = Colors.amber.shade100;
      fg = Colors.amber.shade800;
    } else if (status.contains('CANCEL') || status == 'REJECTED') {
      bg = Colors.red.shade100;
      fg = Colors.red.shade800;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(color: bg, borderRadius: BorderRadius.circular(6)),
      child: Text(status, style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: fg)),
    );
  }
}

class DemandForecastResponseModel {
  final String vendorId;
  final bool hasSufficientHistory;
  final String generalInsight;
  final List<DemandForecastItemModel> forecasts;

  DemandForecastResponseModel({
    required this.vendorId,
    required this.hasSufficientHistory,
    required this.generalInsight,
    required this.forecasts,
  });

  factory DemandForecastResponseModel.fromJson(Map<String, dynamic> json) {
    return DemandForecastResponseModel(
      vendorId: json['vendor_id'] ?? '',
      hasSufficientHistory: json['has_sufficient_history'] ?? false,
      generalInsight: json['general_insight'] ?? '',
      forecasts: (json['forecasts'] as List<dynamic>?)
              ?.map((e) => DemandForecastItemModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

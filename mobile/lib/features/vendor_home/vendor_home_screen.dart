import 'package:flutter/material.dart';
import '../../core/api/api_client.dart';
import '../../shared/models/models.dart';

class VendorHomeScreen extends StatefulWidget {
  final ApiClient apiClient;
  final VoidCallback onLogout;

  const VendorHomeScreen({
    Key? key,
    required this.apiClient,
    required this.onLogout,
  }) : super(key: key);

  @override
  State<VendorHomeScreen> createState() => _VendorHomeScreenState();
}

class _VendorHomeScreenState extends State<VendorHomeScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  bool _isLoading = true;

  Map<String, dynamic>? _salesSummary;
  List<ProductModel> _products = [];
  List<OrderModel> _orders = [];
  List<UdhaarAccountModel> _udhaarAccounts = [];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
    _loadAllVendorData();
  }

  Future<void> _loadAllVendorData() async {
    setState(() => _isLoading = true);
    try {
      // 1. Sales summary
      final summaryRes = await widget.apiClient.get('/analytics/vendor/summary').catchError((_) => null);
      if (summaryRes != null) {
        _salesSummary = summaryRes as Map<String, dynamic>;
      }

      // 2. Products
      final prodRes = await widget.apiClient.get('/products').catchError((_) => null);
      if (prodRes != null && prodRes['data'] != null) {
        _products = (prodRes['data'] as List).map((e) => ProductModel.fromJson(e)).toList();
      }

      // 3. Orders
      final orderRes = await widget.apiClient.get('/orders').catchError((_) => null);
      if (orderRes != null && orderRes['orders'] != null) {
        _orders = (orderRes['orders'] as List).map((e) => OrderModel.fromJson(e)).toList();
      }

      // 4. Udhaar accounts
      final udhaarRes = await widget.apiClient.get('/udhaar/accounts').catchError((_) => null);
      if (udhaarRes != null && udhaarRes['data'] != null) {
        _udhaarAccounts = (udhaarRes['data'] as List).map((e) => UdhaarAccountModel.fromJson(e)).toList();
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _handleStockAdjustment(String productId, int delta) async {
    try {
      await widget.apiClient.post('/inventory/adjust', body: {
        'product_id': productId,
        'qty_delta': delta,
        'reason': 'Manual quick adjustment',
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Stock updated by $delta')),
      );
      _loadAllVendorData();
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Failed to adjust stock: $e'), backgroundColor: Colors.red),
      );
    }
  }

  Future<void> _handleOrderTransition(String orderId, String action) async {
    try {
      await widget.apiClient.post('/orders/$orderId/$action', body: {});
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Order status updated: $action')),
      );
      _loadAllVendorData();
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Action failed: $e'), backgroundColor: Colors.red),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Vendor Dashboard', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
            Text('ವ್ಯಾಪಾರಿ ಮುಖಪುಟ', style: TextStyle(fontSize: 12, color: Color(0xFF64748B))),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadAllVendorData,
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: widget.onLogout,
          ),
        ],
        bottom: TabBar(
          controller: _tabController,
          labelColor: const Color(0xFF1E40AF),
          unselectedLabelColor: const Color(0xFF64748B),
          indicatorColor: const Color(0xFF1E40AF),
          tabs: const [
            Tab(icon: Icon(Icons.dashboard_outlined), text: 'Summary'),
            Tab(icon: Icon(Icons.shopping_bag_outlined), text: 'Orders'),
            Tab(icon: Icon(Icons.inventory_2_outlined), text: 'Products'),
            Tab(icon: Icon(Icons.account_balance_wallet_outlined), text: 'Udhaar'),
          ],
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : TabBarView(
              controller: _tabController,
              children: [
                _buildSummaryTab(),
                _buildOrdersTab(),
                _buildProductsTab(),
                _buildUdhaarTab(),
              ],
            ),
    );
  }

  Widget _buildSummaryTab() {
    final grossSalesPaise = (_salesSummary?['gross_sales_paise'] as num?)?.toInt() ?? 0;
    final totalOrders = (_salesSummary?['total_orders'] as num?)?.toInt() ?? 0;
    final udhaarPaise = (_salesSummary?['outstanding_udhaar_paise'] as num?)?.toInt() ?? 0;

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Card(
            color: const Color(0xFF1E40AF),
            child: Padding(
              padding: const EdgeInsets.all(20.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Completed Sales (ಒಟ್ಟು ಮಾರಾಟ)', style: TextStyle(color: Colors.white70, fontSize: 13)),
                  const SizedBox(height: 8),
                  Text('₹${(grossSalesPaise / 100).toStringAsFixed(2)}',
                      style: const TextStyle(color: Colors.white, fontSize: 28, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      const Icon(Icons.check_circle_outline, color: Colors.greenAccent, size: 16),
                      const SizedBox(width: 4),
                      Text('$totalOrders orders completed', style: const TextStyle(color: Colors.white, fontSize: 13)),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFEF3C7),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: const Icon(Icons.account_balance_wallet, color: Color(0xFFD97706)),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('Customer Udhaar (ಉಧಾರಿ ಬಾಕಿ)', style: TextStyle(color: Color(0xFF64748B), fontSize: 12)),
                        Text('₹${(udhaarPaise / 100).toStringAsFixed(2)}',
                            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18, color: Color(0xFF0F172A))),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildOrdersTab() {
    if (_orders.isEmpty) {
      return const Center(child: Text('No active orders right now'));
    }

    return ListView.builder(
      padding: const EdgeInsets.all(12),
      itemCount: _orders.length,
      itemBuilder: (context, idx) {
        final order = _orders[idx];
        return Card(
          margin: const EdgeInsets.only(bottom: 12),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('Order #${order.id.substring(0, 8)}', style: const TextStyle(fontWeight: FontWeight.bold)),
                    Chip(
                      label: Text(order.status, style: const TextStyle(fontSize: 11, color: Colors.white)),
                      backgroundColor: order.status == 'COMPLETED' ? Colors.green : Colors.orange,
                    ),
                  ],
                ),
                Text('Fulfilment: ${order.fulfilmentType}'),
                Text('Total: ${order.formattedTotal}', style: const TextStyle(fontWeight: FontWeight.bold)),
                const Divider(height: 20),
                Wrap(
                  spacing: 8,
                  children: [
                    if (order.status == 'PENDING') ...[
                      ElevatedButton(
                        onPressed: () => _handleOrderTransition(order.id, 'accept'),
                        child: const Text('Accept'),
                      ),
                      OutlinedButton(
                        onPressed: () => _handleOrderTransition(order.id, 'reject'),
                        child: const Text('Reject'),
                      ),
                    ],
                    if (order.status == 'ACCEPTED')
                      ElevatedButton(
                        onPressed: () => _handleOrderTransition(order.id, 'prepare'),
                        child: const Text('Start Preparing'),
                      ),
                    if (order.status == 'PREPARING') ...[
                      if (order.fulfilmentType == 'PICKUP')
                        ElevatedButton(
                          onPressed: () => _handleOrderTransition(order.id, 'ready'),
                          child: const Text('Ready for Pickup'),
                        )
                      else
                        ElevatedButton(
                          onPressed: () => _handleOrderTransition(order.id, 'out-for-delivery'),
                          child: const Text('Out for Delivery'),
                        ),
                    ],
                    if (order.status == 'READY_FOR_PICKUP' || order.status == 'OUT_FOR_DELIVERY')
                      ElevatedButton(
                        onPressed: () => _handleOrderTransition(order.id, 'complete'),
                        child: const Text('Complete Order'),
                      ),
                  ],
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildProductsTab() {
    if (_products.isEmpty) {
      return const Center(child: Text('No products listed yet'));
    }

    return ListView.builder(
      padding: const EdgeInsets.all(12),
      itemCount: _products.length,
      itemBuilder: (context, idx) {
        final p = _products[idx];
        return Card(
          margin: const EdgeInsets.only(bottom: 12),
          child: ListTile(
            title: Text(p.name, style: const TextStyle(fontWeight: FontWeight.w600)),
            subtitle: Text(p.formattedPrice),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  icon: const Icon(Icons.remove_circle_outline, color: Colors.red),
                  onPressed: () => _handleStockAdjustment(p.id, -1),
                ),
                IconButton(
                  icon: const Icon(Icons.add_circle_outline, color: Colors.green),
                  onPressed: () => _handleStockAdjustment(p.id, 1),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildUdhaarTab() {
    if (_udhaarAccounts.isEmpty) {
      return const Center(child: Text('No customer Udhaar accounts recorded'));
    }

    return ListView.builder(
      padding: const EdgeInsets.all(12),
      itemCount: _udhaarAccounts.length,
      itemBuilder: (context, idx) {
        final acc = _udhaarAccounts[idx];
        return Card(
          child: ListTile(
            leading: const Icon(Icons.person_outline),
            title: Text('Customer ID: ${acc.customerId.substring(0, 8)}...'),
            subtitle: Text('Status: ${acc.status}'),
            trailing: Text(acc.formattedBalance, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
          ),
        );
      },
    );
  }
}

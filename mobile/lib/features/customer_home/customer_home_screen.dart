import 'package:flutter/material.dart';
import '../../core/api/api_client.dart';
import '../../shared/models/models.dart';

class CustomerHomeScreen extends StatefulWidget {
  final ApiClient apiClient;
  final VoidCallback onLogout;

  const CustomerHomeScreen({
    Key? key,
    required this.apiClient,
    required this.onLogout,
  }) : super(key: key);

  @override
  State<CustomerHomeScreen> createState() => _CustomerHomeScreenState();
}

class _CustomerHomeScreenState extends State<CustomerHomeScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  bool _isLoading = true;

  List<NearbyVendorModel> _nearbyVendors = [];
  CartModel? _cart;
  List<OrderModel> _myOrders = [];

  // Default demo coordinate (Bangalore Majestic)
  final double _demoLat = 12.9767;
  final double _demoLng = 77.5713;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _loadCustomerData();
  }

  Future<void> _loadCustomerData() async {
    setState(() => _isLoading = true);
    try {
      // 1. Nearby Vendors
      final vendorRes = await widget.apiClient.get(
        '/mobility/nearby-vendors',
        queryParams: {'lat': _demoLat.toString(), 'lng': _demoLng.toString(), 'radius_km': '10.0'},
      ).catchError((_) => null);
      if (vendorRes != null) {
        _nearbyVendors = (vendorRes as List).map((e) => NearbyVendorModel.fromJson(e)).toList();
      }

      // 2. Cart
      final cartRes = await widget.apiClient.get('/carts/me').catchError((_) => null);
      if (cartRes != null) {
        _cart = CartModel.fromJson(cartRes);
      }

      // 3. Orders
      final ordersRes = await widget.apiClient.get('/orders').catchError((_) => null);
      if (ordersRes != null && ordersRes['orders'] != null) {
        _myOrders = (ordersRes['orders'] as List).map((e) => OrderModel.fromJson(e)).toList();
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _handleCheckout(String fulfilmentType) async {
    if (_cart == null || _cart!.items.isEmpty) return;

    try {
      await widget.apiClient.post('/orders', body: {
        'fulfilment_type': fulfilmentType,
      });

      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Order placed successfully!')),
      );

      _loadCustomerData();
      _tabController.animateTo(2); // Switch to orders tab
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Checkout failed: $e'), backgroundColor: Colors.red),
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
            Text('VendorSaathi Customer', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
            Text('ಹತ್ತಿರದ ವ್ಯಾಪಾರಿಗಳ ಹುಡುಕಾಟ', style: TextStyle(fontSize: 12, color: Color(0xFF64748B))),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadCustomerData,
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
          tabs: [
            const Tab(icon: Icon(Icons.near_me_outlined), text: 'Discover'),
            Tab(
              icon: Badge(
                label: Text('${_cart?.items.length ?? 0}'),
                isLabelVisible: (_cart?.items.length ?? 0) > 0,
                child: const Icon(Icons.shopping_cart_outlined),
              ),
              text: 'Cart',
            ),
            const Tab(icon: Icon(Icons.receipt_long_outlined), text: 'My Orders'),
          ],
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : TabBarView(
              controller: _tabController,
              children: [
                _buildDiscoverTab(),
                _buildCartTab(),
                _buildOrdersTab(),
              ],
            ),
    );
  }

  Widget _buildDiscoverTab() {
    if (_nearbyVendors.isEmpty) {
      return const Center(child: Text('No active vendors found within 10 km'));
    }

    return ListView.builder(
      padding: const EdgeInsets.all(12),
      itemCount: _nearbyVendors.length,
      itemBuilder: (context, idx) {
        final v = _nearbyVendors[idx];
        return Card(
          margin: const EdgeInsets.only(bottom: 12),
          child: ListTile(
            leading: CircleAvatar(
              backgroundColor: const Color(0xFFEFF6FF),
              child: const Icon(Icons.store, color: Color(0xFF1E40AF)),
            ),
            title: Text(v.businessName, style: const TextStyle(fontWeight: FontWeight.bold)),
            subtitle: Text('${v.distanceKm.toStringAsFixed(1)} km away • ${v.city}'),
            trailing: Chip(
              label: Text(v.availabilityStatus, style: const TextStyle(fontSize: 11)),
              backgroundColor: v.availabilityStatus == 'OPEN' ? const Color(0xFFDCFCE7) : const Color(0xFFFEE2E2),
            ),
          ),
        );
      },
    );
  }

  Widget _buildCartTab() {
    if (_cart == null || _cart!.items.isEmpty) {
      return const Center(child: Text('Your cart is empty'));
    }

    return Padding(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        children: [
          Expanded(
            child: ListView.builder(
              itemCount: _cart!.items.length,
              itemBuilder: (context, idx) {
                final it = _cart!.items[idx];
                return ListTile(
                  title: Text(it.productName),
                  subtitle: Text('Qty: ${it.quantity}'),
                  trailing: Text(it.formattedTotal, style: const TextStyle(fontWeight: FontWeight.w600)),
                );
              },
            ),
          ),
          const Divider(),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('Total to Pay:', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
              Text(_cart!.formattedTotal, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Color(0xFF1E40AF))),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: OutlinedButton(
                  onPressed: () => _handleCheckout('PICKUP'),
                  child: const Text('Store Pickup (ಪಿಕಪ್)'),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: ElevatedButton(
                  onPressed: () => _handleCheckout('DELIVERY'),
                  child: const Text('Delivery (ವಿತರಣೆ)'),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildOrdersTab() {
    if (_myOrders.isEmpty) {
      return const Center(child: Text('You have not placed any orders yet'));
    }

    return ListView.builder(
      padding: const EdgeInsets.all(12),
      itemCount: _myOrders.length,
      itemBuilder: (context, idx) {
        final order = _myOrders[idx];
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
                      backgroundColor: order.status == 'COMPLETED' ? Colors.green : Colors.blue,
                    ),
                  ],
                ),
                Text('Fulfilment: ${order.fulfilmentType}'),
                Text('Total: ${order.formattedTotal}', style: const TextStyle(fontWeight: FontWeight.bold)),
                Text('Date: ${order.createdAt.toLocal().toString().substring(0, 16)}', style: const TextStyle(color: Colors.grey, fontSize: 12)),
              ],
            ),
          ),
        );
      },
    );
  }
}

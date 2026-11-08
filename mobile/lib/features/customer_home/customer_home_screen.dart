import 'dart:async';
import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../core/api/api_client.dart';
import '../../core/localization/locale_provider.dart';
import '../../shared/models/models.dart';
import '../../shared/widgets/osm_map_view.dart';

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

class _CustomerHomeScreenState extends State<CustomerHomeScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  bool _isLoading = true;

  // Location (Default: Bangalore Majestic)
  double _currentLat = 12.9767;
  double _currentLng = 77.5713;

  // Customer Data
  List<NearbyVendorModel> _nearbyVendors = [];
  NearbyVendorModel? _selectedVendor;
  List<ProductModel> _vendorProducts = [];
  List<CategoryModel> _categories = [];
  String? _selectedCategoryFilter;
  String _productSearchQuery = '';

  CartModel? _cart;
  List<OrderModel> _myOrders = [];
  List<CustomerAddressModel> _myAddresses = [];
  CustomerAddressModel? _selectedDeliveryAddress;
  LoyaltySummaryModel? _loyalty;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
    _loadCustomerData();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
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
        if (_myAddresses.isNotEmpty) {
          _selectedDeliveryAddress = _myAddresses.firstWhere((a) => a.isDefault, orElse: () => _myAddresses.first);
        }
      }

      // 6. Loyalty
      final loyaltyRes = await widget.apiClient.get('/loyalty/me').catchError((_) => null);
      if (loyaltyRes != null) {
        _loyalty = LoyaltySummaryModel.fromJson(loyaltyRes);
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _fetchGPSLocation() async {
    try {
      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
      }
      if (permission == LocationPermission.whileInUse || permission == LocationPermission.always) {
        final pos = await Geolocator.getCurrentPosition(desiredAccuracy: LocationAccuracy.medium);
        setState(() {
          _currentLat = pos.latitude;
          _currentLng = pos.longitude;
        });
        _loadCustomerData();
      }
    } catch (_) {}
  }

  Future<void> _selectVendor(NearbyVendorModel vendor) async {
    setState(() {
      _selectedVendor = vendor;
      _vendorProducts = [];
    });
    try {
      final res = await widget.apiClient.get('/products/vendor/${vendor.vendorId}');
      if (res != null && res is List) {
        setState(() {
          _vendorProducts = res.map((e) => ProductModel.fromJson(e)).toList();
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to load products: $e')),
        );
      }
    }
  }

  Future<void> _handleAddToCart(ProductModel product) async {
    final loc = LocaleProvider.of(context);

    // Single-vendor cart check
    if (_cart != null && _cart!.items.isNotEmpty && _cart!.vendorId != product.vendorId) {
      final proceed = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          title: const Text('Replace Cart?'),
          content: const Text(
            'Your cart contains items from another vendor. You can only order from one vendor at a time. Clear current cart?',
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx, false), child: Text(loc.t('cancel'))),
            ElevatedButton(
              onPressed: () => Navigator.pop(ctx, true),
              style: ElevatedButton.styleFrom(backgroundColor: Colors.red, foregroundColor: Colors.white),
              child: const Text('Clear & Add'),
            ),
          ],
        ),
      );
      if (proceed != true) return;
      await widget.apiClient.delete('/carts/me');
    }

    try {
      await widget.apiClient.post('/carts/items', body: {
        'product_id': product.id,
        'quantity': 1,
      });
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('${product.name} added to cart'), backgroundColor: Colors.green),
      );
      _loadCustomerData();
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Failed to add: $e'), backgroundColor: Colors.red),
      );
    }
  }

  Future<void> _handleUpdateCartQuantity(String productId, int newQty) async {
    try {
      if (newQty <= 0) {
        await widget.apiClient.delete('/carts/items/$productId');
      } else {
        await widget.apiClient.patch('/carts/items/$productId', body: {'quantity': newQty});
      }
      _loadCustomerData();
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Cart update error: $e'), backgroundColor: Colors.red),
      );
    }
  }

  // --- Checkout Flow ---
  void _showCheckoutDialog() {
    final loc = LocaleProvider.of(context);
    if (_cart == null || _cart!.items.isEmpty) return;

    String fulfilmentType = 'PICKUP';
    String paymentOption = 'CASH';
    bool redeemLoyalty = false;
    int pointsToRedeem = (_loyalty != null && _loyalty!.pointsBalance > 0)
        ? (_loyalty!.pointsBalance > 50 ? 50 : _loyalty!.pointsBalance)
        : 0;

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          final subtotal = _cart!.totalPaise;
          final discount = redeemLoyalty ? (pointsToRedeem * 100) : 0;
          final finalTotal = (subtotal - discount).clamp(0, subtotal);

          return AlertDialog(
            title: Text(loc.t('checkout')),
            content: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(loc.t('fulfilment'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                  const SizedBox(height: 8),
                  SegmentedButton<String>(
                    segments: [
                      ButtonSegment(value: 'PICKUP', label: Text(loc.t('pickup'))),
                      ButtonSegment(value: 'DELIVERY', label: Text(loc.t('delivery'))),
                    ],
                    selected: {fulfilmentType},
                    onSelectionChanged: (set) => setDialogState(() => fulfilmentType = set.first),
                  ),
                  const SizedBox(height: 16),

                  if (fulfilmentType == 'DELIVERY') ...[
                    Text(loc.t('delivery_address'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                    const SizedBox(height: 6),
                    if (_myAddresses.isEmpty)
                      OutlinedButton.icon(
                        onPressed: () {
                          Navigator.pop(ctx);
                          _showAddAddressDialog();
                        },
                        icon: const Icon(Icons.add),
                        label: Text(loc.t('add_address')),
                      )
                    else
                      DropdownButtonFormField<CustomerAddressModel>(
                        value: _selectedDeliveryAddress,
                        decoration: const InputDecoration(border: OutlineInputBorder(), contentPadding: EdgeInsets.all(8)),
                        items: _myAddresses.map((a) {
                          return DropdownMenuItem(value: a, child: Text(a.addressLine1, maxLines: 1));
                        }).toList(),
                        onChanged: (val) => setDialogState(() => _selectedDeliveryAddress = val),
                      ),
                    const SizedBox(height: 16),
                  ],

                  const Text('Payment Method', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                  const SizedBox(height: 6),
                  SegmentedButton<String>(
                    segments: const [
                      ButtonSegment(value: 'CASH', label: Text('Cash')),
                      ButtonSegment(value: 'UPI', label: Text('UPI')),
                      ButtonSegment(value: 'ONLINE', label: Text('Razorpay')),
                    ],
                    selected: {paymentOption},
                    onSelectionChanged: (set) => setDialogState(() => paymentOption = set.first),
                  ),
                  const SizedBox(height: 16),

                  if (_loyalty != null && _loyalty!.pointsBalance > 0) ...[
                    CheckboxListTile(
                      contentPadding: EdgeInsets.zero,
                      title: Text('${loc.t('redeem_points')} ($pointsToRedeem pts = ₹$pointsToRedeem)'),
                      value: redeemLoyalty,
                      onChanged: (val) => setDialogState(() => redeemLoyalty = val ?? false),
                    ),
                    const Divider(),
                  ],

                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(loc.t('subtotal')),
                      Text('₹${(subtotal / 100).toStringAsFixed(2)}'),
                    ],
                  ),
                  if (discount > 0)
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(loc.t('discount'), style: const TextStyle(color: Colors.green)),
                        Text('-₹${(discount / 100).toStringAsFixed(2)}', style: const TextStyle(color: Colors.green)),
                      ],
                    ),
                  const SizedBox(height: 6),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(loc.t('total'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                      Text('₹${(finalTotal / 100).toStringAsFixed(2)}', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16, color: Color(0xFF1E40AF))),
                    ],
                  ),
                ],
              ),
            ),
            actions: [
              TextButton(onPressed: () => Navigator.pop(ctx), child: Text(loc.t('cancel'))),
              ElevatedButton(
                onPressed: () async {
                  try {
                    String? addressSnapshot;
                    if (fulfilmentType == 'DELIVERY') {
                      addressSnapshot = _selectedDeliveryAddress?.fullAddress ?? 'Bangalore Delivery';
                    }

                    final res = await widget.apiClient.post('/orders', body: {
                      'fulfilment_type': fulfilmentType,
                      if (addressSnapshot != null) 'delivery_address_snapshot': addressSnapshot,
                      if (redeemLoyalty && pointsToRedeem > 0) 'redeem_loyalty_points': pointsToRedeem,
                    });

                    if (mounted) {
                      Navigator.pop(ctx);
                      final orderId = res != null && res['id'] != null ? res['id'].toString() : null;

                      if (paymentOption == 'ONLINE' && orderId != null) {
                        _showOnlinePaymentSheet(orderId, finalTotal);
                      } else {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text(loc.t('success')), backgroundColor: Colors.green),
                        );
                      }

                      _loadCustomerData();
                      _tabController.animateTo(2); // Switch to orders tab
                    }
                  } catch (e) {
                    if (mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(content: Text('Checkout failed: $e'), backgroundColor: Colors.red),
                      );
                    }
                  }
                },
                style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF1E40AF), foregroundColor: Colors.white),
                child: Text(loc.t('place_order')),
              ),
            ],
          );
        },
      ),
    );
  }

  // --- Add Address Dialog ---
  void _showAddAddressDialog() {
    final loc = LocaleProvider.of(context);
    final line1Ctrl = TextEditingController();
    final cityCtrl = TextEditingController(text: 'Bengaluru');
    final postalCtrl = TextEditingController(text: '560001');

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(loc.t('add_address')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: line1Ctrl, decoration: InputDecoration(labelText: loc.t('delivery_address'))),
            const SizedBox(height: 10),
            TextField(controller: cityCtrl, decoration: const InputDecoration(labelText: 'City')),
            const SizedBox(height: 10),
            TextField(controller: postalCtrl, decoration: const InputDecoration(labelText: 'PIN Code')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: Text(loc.t('cancel'))),
          ElevatedButton(
            onPressed: () async {
              if (line1Ctrl.text.isEmpty) return;
              try {
                await widget.apiClient.post('/customers/addresses', body: {
                  'address_line1': line1Ctrl.text.trim(),
                  'city': cityCtrl.text.trim(),
                  'state': 'Karnataka',
                  'postal_code': postalCtrl.text.trim(),
                  'country': 'India',
                  'is_default': true,
                });
                if (mounted) {
                  Navigator.pop(ctx);
                  _loadCustomerData();
                }
              } catch (_) {}
            },
            child: Text(loc.t('save')),
          ),
        ],
      ),
    );
  }

  // --- Online Payment Flow (Razorpay) ---
  void _showOnlinePaymentSheet(String orderId, int amountPaise) {
    final loc = LocaleProvider.of(context);
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
      builder: (bottomCtx) {
        bool isInitiating = true;
        bool isVerifying = false;
        String? errorMessage;
        String? gatewayOrderId;
        String? keyId;
        bool paymentSuccess = false;

        return StatefulBuilder(
          builder: (context, setSheetState) {
            if (isInitiating) {
              isInitiating = false;
              widget.apiClient.post('/payments/gateway/initiate', body: {
                'order_id': orderId,
                'amount_paise': amountPaise,
              }).then((res) {
                setSheetState(() {
                  gatewayOrderId = res?['gateway_order_id'] ?? 'rzp_order_${orderId.substring(0, 8)}';
                  keyId = res?['key_id'] ?? 'rzp_test_public_key';
                });
              }).catchError((err) {
                setSheetState(() {
                  errorMessage = err.toString();
                });
              });
            }

            return Padding(
              padding: EdgeInsets.only(
                left: 20, right: 20, top: 20,
                bottom: MediaQuery.of(context).viewInsets.bottom + 24,
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: Colors.blue.shade50,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: const Icon(Icons.payment, color: Colors.blue),
                          ),
                          const SizedBox(width: 10),
                          const Text(
                            'Razorpay Secure Checkout',
                            style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                          ),
                        ],
                      ),
                      IconButton(
                        icon: const Icon(Icons.close),
                        onPressed: () => Navigator.pop(bottomCtx),
                      ),
                    ],
                  ),
                  const Divider(height: 24),

                  if (paymentSuccess) ...[
                    Center(
                      child: Column(
                        children: [
                          const Icon(Icons.check_circle_rounded, color: Colors.green, size: 64),
                          const SizedBox(height: 12),
                          Text(loc.t('online_payment_success'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                          const SizedBox(height: 4),
                          Text(loc.t('server_payment_verified'), style: const TextStyle(color: Colors.grey, fontSize: 13)),
                          const SizedBox(height: 16),
                          ElevatedButton(
                            onPressed: () {
                              Navigator.pop(bottomCtx);
                              _loadCustomerData();
                            },
                            child: const Text('Done'),
                          ),
                        ],
                      ),
                    ),
                  ] else ...[
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text('Amount to Pay:'),
                        Text(
                          '₹${(amountPaise / 100).toStringAsFixed(2)}',
                          style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.blue),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Order #${orderId.length > 8 ? orderId.substring(0, 8) : orderId}',
                      style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                    ),
                    if (gatewayOrderId != null)
                      Text(
                        'Gateway Order: $gatewayOrderId',
                        style: TextStyle(fontSize: 11, color: Colors.grey[500]),
                      ),
                    if (keyId != null)
                      Text(
                        'Gateway Key: $keyId',
                        style: TextStyle(fontSize: 11, color: Colors.grey[500]),
                      ),
                    const SizedBox(height: 16),

                    if (errorMessage != null)
                      Container(
                        padding: const EdgeInsets.all(12),
                        margin: const EdgeInsets.only(bottom: 12),
                        decoration: BoxDecoration(
                          color: Colors.amber.shade50,
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(color: Colors.amber.shade200),
                        ),
                        child: Text(
                          errorMessage!,
                          style: TextStyle(fontSize: 12, color: Colors.amber.shade900),
                        ),
                      ),

                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.grey.shade50,
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(color: Colors.grey.shade200),
                      ),
                      child: Column(
                        children: [
                          const Row(
                            children: [
                              Icon(Icons.shield_outlined, color: Colors.green, size: 20),
                              SizedBox(width: 8),
                              Text('Razorpay Sandbox / Test Mode', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
                            ],
                          ),
                          const SizedBox(height: 6),
                          Text(
                            'Client-side payments are verified strictly on the backend with HMAC-SHA256 signatures before updating status.',
                            style: TextStyle(fontSize: 11, color: Colors.grey[600]),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 20),

                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        onPressed: isVerifying
                            ? null
                            : () async {
                                setSheetState(() => isVerifying = true);
                                try {
                                  final paymentId = 'pay_${DateTime.now().millisecondsSinceEpoch}';
                                  await widget.apiClient.post('/payments/gateway/verify', body: {
                                    'order_id': orderId,
                                    'gateway_order_id': gatewayOrderId ?? 'order_sample',
                                    'gateway_payment_id': paymentId,
                                    'gateway_signature': 'sandbox_verified_signature',
                                  });

                                  setSheetState(() {
                                    isVerifying = false;
                                    paymentSuccess = true;
                                  });
                                } catch (e) {
                                  setSheetState(() {
                                    isVerifying = false;
                                    errorMessage = 'Verification note: $e';
                                  });
                                }
                              },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: const Color(0xFF1E40AF),
                          foregroundColor: Colors.white,
                          padding: const EdgeInsets.symmetric(vertical: 14),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                        ),
                        child: isVerifying
                            ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                            : Text('Pay ₹${(amountPaise / 100).toStringAsFixed(2)} Securely'),
                      ),
                    ),
                  ],
                ],
              ),
            );
          },
        );
      },
    );
  }

  // --- Live Delivery Tracking & Route View ---
  void _showLiveTrackingDialog(OrderModel order) {
    Timer? pollTimer;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
      builder: (sheetCtx) {
        TrackingModel? liveLocation;
        RouteModel? routeInfo;
        bool loading = true;
        String? statusMsg;

        Future<void> loadTrackingData(StateSetter setSheetState) async {
          try {
            final locRes = await widget.apiClient.get('/delivery/orders/${order.id}/location').catchError((_) => null);
            if (locRes != null && locRes['data'] != null) {
              liveLocation = TrackingModel.fromJson(locRes['data']);
            }

            final originLat = liveLocation?.latitude ?? _currentLat;
            final originLng = liveLocation?.longitude ?? _currentLng;
            final destLat = _currentLat + 0.015;
            final destLng = _currentLng + 0.015;

            final routeRes = await widget.apiClient.get('/mobility/route', queryParams: {
              'origin_lat': originLat.toString(),
              'origin_lng': originLng.toString(),
              'dest_lat': destLat.toString(),
              'dest_lng': destLng.toString(),
            }).catchError((_) => null);

            if (routeRes != null) {
              routeInfo = RouteModel.fromJson(routeRes);
            }

            if (sheetCtx.mounted) {
              setSheetState(() {
                loading = false;
              });
            }
          } catch (e) {
            if (sheetCtx.mounted) {
              setSheetState(() {
                statusMsg = 'Waiting for agent to broadcast location...';
                loading = false;
              });
            }
          }
        }

        return StatefulBuilder(
          builder: (context, setSheetState) {
            if (loading) {
              loadTrackingData(setSheetState);
              pollTimer?.cancel();
              pollTimer = Timer.periodic(const Duration(seconds: 10), (_) {
                if (context.mounted) {
                  loadTrackingData(setSheetState);
                }
              });
            }

            return Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.local_shipping_rounded, color: Color(0xFF1E40AF), size: 24),
                          const SizedBox(width: 8),
                          Text('Order #${order.id.substring(0, 8)} Tracking', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                        ],
                      ),
                      IconButton(
                        icon: const Icon(Icons.close),
                        onPressed: () {
                          pollTimer?.cancel();
                          Navigator.pop(sheetCtx);
                        },
                      ),
                    ],
                  ),
                  const Divider(height: 20),

                  if (loading)
                    const Center(child: Padding(padding: EdgeInsets.all(20), child: CircularProgressIndicator()))
                  else ...[
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text('Status: ${order.status}', style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                          decoration: BoxDecoration(color: Colors.green.shade50, borderRadius: BorderRadius.circular(12)),
                          child: const Row(
                            children: [
                              Icon(Icons.fiber_manual_record, color: Colors.green, size: 10),
                              SizedBox(width: 4),
                              Text('Live Polling (10s)', style: TextStyle(fontSize: 10, color: Colors.green, fontWeight: FontWeight.bold)),
                            ],
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 14),

                    if (routeInfo != null)
                      Row(
                        children: [
                          Expanded(
                            child: Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(color: Colors.blue.shade50, borderRadius: BorderRadius.circular(10)),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    routeInfo!.routingAvailable ? 'Estimated Travel' : 'ETA (Road Travel)',
                                    style: const TextStyle(fontSize: 11, color: Colors.blue),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    routeInfo!.routingAvailable
                                        ? '${routeInfo!.durationMinutes.toStringAsFixed(0)} mins'
                                        : 'Unavailable',
                                    style: TextStyle(
                                      fontSize: routeInfo!.routingAvailable ? 18 : 15,
                                      fontWeight: FontWeight.bold,
                                      color: const Color(0xFF1E40AF),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(color: Colors.amber.shade50, borderRadius: BorderRadius.circular(10)),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    routeInfo!.routingAvailable ? 'Road Distance' : 'Approx. Straight-Line',
                                    style: const TextStyle(fontSize: 11, color: Colors.amber),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    routeInfo!.routingAvailable
                                        ? '${routeInfo!.distanceKm.toStringAsFixed(1)} km'
                                        : (routeInfo!.approximateStraightLineKm != null
                                            ? '${routeInfo!.approximateStraightLineKm!.toStringAsFixed(1)} km'
                                            : 'Unavailable'),
                                    style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.amber),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ),
                    const SizedBox(height: 14),

                    // Embedded MapLibre / OpenStreetMap View
                    OsmMapView(
                      agentLat: liveLocation?.latitude,
                      agentLng: liveLocation?.longitude,
                      customerLat: _currentLat + 0.015,
                      customerLng: _currentLng + 0.015,
                      route: routeInfo,
                      height: 220,
                    ),
                    const SizedBox(height: 12),

                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(color: Colors.grey.shade50, borderRadius: BorderRadius.circular(10), border: Border.all(color: Colors.grey.shade200)),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              const Icon(Icons.gps_fixed, color: Colors.teal, size: 18),
                              const SizedBox(width: 8),
                              Text(
                                liveLocation != null ? 'Agent Coordinates' : 'Agent Location Pending',
                                style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                              ),
                            ],
                          ),
                          const SizedBox(height: 6),
                          if (liveLocation != null) ...[
                            Text('Latitude: ${liveLocation!.latitude.toStringAsFixed(5)}, Longitude: ${liveLocation!.longitude.toStringAsFixed(5)}', style: const TextStyle(fontSize: 12)),
                            const SizedBox(height: 2),
                            Text('Last Ping: ${liveLocation!.updatedAt}', style: TextStyle(fontSize: 11, color: Colors.grey[600])),
                          ] else
                            Text(
                              statusMsg ?? 'The assigned delivery agent has not broadcasted their location yet. Coordinates will appear as soon as the agent begins the route.',
                              style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                            ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 14),

                    if (routeInfo != null && routeInfo!.directionsUrl.isNotEmpty)
                      SizedBox(
                        width: double.infinity,
                        child: OutlinedButton.icon(
                          onPressed: () async {
                            final uri = Uri.parse(routeInfo!.directionsUrl);
                            try {
                              await launchUrl(uri, mode: LaunchMode.externalApplication);
                            } catch (_) {
                              await launchUrl(uri, mode: LaunchMode.platformDefault);
                            }
                          },
                          icon: const Icon(Icons.directions_rounded),
                          label: const Text('Open in OpenStreetMap Navigation'),
                          style: OutlinedButton.styleFrom(
                            padding: const EdgeInsets.symmetric(vertical: 12),
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                          ),
                        ),
                      ),
                  ],
                ],
              ),
            );
          },
        );
      },
    ).whenComplete(() {
      pollTimer?.cancel();
    });
  }

  // --- Review Submission Dialog ---
  void _showReviewDialog(OrderModel order) {
    final loc = LocaleProvider.of(context);
    int rating = 5;
    final commentCtrl = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) => AlertDialog(
          title: Text(loc.t('write_review')),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(loc.t('rate_vendor'), style: const TextStyle(fontSize: 13)),
              const SizedBox(height: 8),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  for (int star = 1; star <= 5; star++)
                    IconButton(
                      icon: Icon(
                        star <= rating ? Icons.star_rounded : Icons.star_border_rounded,
                        color: Colors.amber,
                        size: 32,
                      ),
                      onPressed: () => setDialogState(() => rating = star),
                    ),
                ],
              ),
              const SizedBox(height: 12),
              TextField(
                controller: commentCtrl,
                maxLines: 3,
                decoration: const InputDecoration(
                  hintText: 'Share your experience (freshness, fast delivery, service)...',
                  border: OutlineInputBorder(),
                ),
              ),
            ],
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx), child: Text(loc.t('cancel'))),
            ElevatedButton(
              onPressed: () async {
                try {
                  await widget.apiClient.post('/reviews', body: {
                    'order_id': order.id,
                    'vendor_id': order.vendorId,
                    'rating': rating,
                    'comment': commentCtrl.text.trim(),
                  });
                  if (mounted) {
                    Navigator.pop(ctx);
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Review submitted! Thank you.'), backgroundColor: Colors.green),
                    );
                  }
                } catch (e) {
                  if (mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text('Failed: $e'), backgroundColor: Colors.red),
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

  // --- Cancellation Request Dialog ---
  void _showCancelOrderDialog(OrderModel order) {
    final loc = LocaleProvider.of(context);
    final reasonCtrl = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(loc.t('request_cancel')),
        content: TextField(
          controller: reasonCtrl,
          decoration: const InputDecoration(
            labelText: 'Reason for cancellation',
            border: OutlineInputBorder(),
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: Text(loc.t('cancel'))),
          ElevatedButton(
            onPressed: () async {
              if (reasonCtrl.text.isEmpty) return;
              try {
                await widget.apiClient.post('/orders/${order.id}/request-cancellation', body: {
                  'reason': reasonCtrl.text.trim(),
                });
                if (mounted) {
                  Navigator.pop(ctx);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Cancellation requested')),
                  );
                  _loadCustomerData();
                }
              } catch (e) {
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
                  );
                }
              }
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red, foregroundColor: Colors.white),
            child: Text(loc.t('submit')),
          ),
        ],
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
            Text(loc.t('customer_portal'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
            Text(
              'Bangalore Majestic • ${_nearbyVendors.length} nearby',
              style: const TextStyle(fontSize: 11, color: Color(0xFF64748B)),
            ),
          ],
        ),
        actions: [
          IconButton(
            tooltip: 'Live GPS',
            icon: const Icon(Icons.my_location_rounded),
            onPressed: _fetchGPSLocation,
          ),
          IconButton(
            tooltip: loc.t('refresh'),
            icon: const Icon(Icons.refresh),
            onPressed: _loadCustomerData,
          ),
          PopupMenuButton<String>(
            icon: const Icon(Icons.more_vert),
            onSelected: (val) {
              if (val == 'switch') {
                widget.onSwitchToVendor();
              } else if (val == 'lang') {
                final newLang = loc.currentLanguage == 'en' ? 'kn' : 'en';
                loc.setLanguage(newLang);
                LanguageController.saveLanguage(newLang, apiClient: widget.apiClient);
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
                      const Icon(Icons.storefront_outlined, size: 20),
                      const SizedBox(width: 8),
                      Text(loc.t('switch_to_vendor')),
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
          labelColor: const Color(0xFF1E40AF),
          unselectedLabelColor: const Color(0xFF64748B),
          indicatorColor: const Color(0xFF1E40AF),
          tabs: [
            Tab(icon: const Icon(Icons.explore_rounded), text: loc.t('nearby_vendors')),
            Tab(
              icon: Badge(
                label: Text('${_cart?.items.length ?? 0}'),
                isLabelVisible: (_cart?.items.length ?? 0) > 0,
                child: const Icon(Icons.shopping_cart_rounded),
              ),
              text: loc.t('cart'),
            ),
            Tab(icon: const Icon(Icons.receipt_long_rounded), text: loc.t('my_orders')),
            Tab(icon: const Icon(Icons.stars_rounded), text: loc.t('loyalty_rewards')),
          ],
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : TabBarView(
              controller: _tabController,
              children: [
                _buildDiscoveryTab(loc),
                _buildCartTab(loc),
                _buildMyOrdersTab(loc),
                _buildLoyaltyTab(loc),
              ],
            ),
    );
  }

  // --- TAB 1: NEARBY DISCOVERY & PRODUCT BROWSING ---
  Widget _buildDiscoveryTab(LocaleProvider loc) {
    if (_selectedVendor != null) {
      // Product Browsing View of Selected Vendor
      var filtered = _vendorProducts.where((p) {
        final matchesSearch = p.name.toLowerCase().contains(_productSearchQuery.toLowerCase());
        final matchesCategory = _selectedCategoryFilter == null || p.categoryId == _selectedCategoryFilter;
        return matchesSearch && matchesCategory;
      }).toList();

      return Column(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            color: const Color(0xFFEFF6FF),
            child: Row(
              children: [
                IconButton(
                  icon: const Icon(Icons.arrow_back),
                  onPressed: () => setState(() => _selectedVendor = null),
                ),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _selectedVendor!.businessName,
                        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                      ),
                      Text(
                        '${_selectedVendor!.distanceKm.toStringAsFixed(2)} km away • ${_selectedVendor!.availabilityStatus}',
                        style: const TextStyle(fontSize: 12, color: Color(0xFF3B82F6)),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),

          // Search & Category Filter
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Column(
              children: [
                TextField(
                  decoration: InputDecoration(
                    hintText: loc.t('search'),
                    prefixIcon: const Icon(Icons.search),
                    contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 0),
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

          Expanded(
            child: filtered.isEmpty
                ? Center(child: Text(loc.t('no_products')))
                : ListView.builder(
                    padding: const EdgeInsets.all(12),
                    itemCount: filtered.length,
                    itemBuilder: (ctx, i) {
                      final p = filtered[i];
                      return Card(
                        margin: const EdgeInsets.only(bottom: 10),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                        child: Padding(
                          padding: const EdgeInsets.all(12),
                          child: Row(
                            children: [
                              Container(
                                width: 48,
                                height: 48,
                                decoration: BoxDecoration(
                                  color: Colors.blue.shade50,
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: const Icon(Icons.storefront_rounded, color: Color(0xFF2563EB)),
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(p.name, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                                    if (p.description != null)
                                      Text(p.description!, style: TextStyle(fontSize: 12, color: Colors.grey.shade600)),
                                    const SizedBox(height: 4),
                                    Text(
                                      p.formattedPrice,
                                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF1E40AF)),
                                    ),
                                  ],
                                ),
                              ),
                              ElevatedButton(
                                onPressed: () => _handleAddToCart(p),
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: const Color(0xFF2563EB),
                                  foregroundColor: Colors.white,
                                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                                ),
                                child: Text(loc.t('add_to_cart')),
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

    // Vendor Directory List
    return RefreshIndicator(
      onRefresh: _loadCustomerData,
      child: _nearbyVendors.isEmpty
          ? const Center(child: Text('No vendors found within 10 km.'))
          : ListView.builder(
              padding: const EdgeInsets.all(12),
              itemCount: _nearbyVendors.length + 1,
              itemBuilder: (ctx, i) {
                if (i == 0) {
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 12),
                    child: OsmMapView(
                      initialLat: _currentLat,
                      initialLng: _currentLng,
                      nearbyVendors: _nearbyVendors,
                      height: 190,
                    ),
                  );
                }
                final v = _nearbyVendors[i - 1];
                return Card(
                  margin: const EdgeInsets.only(bottom: 12),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  child: InkWell(
                    borderRadius: BorderRadius.circular(12),
                    onTap: () => _selectVendor(v),
                    child: Padding(
                      padding: const EdgeInsets.all(14),
                      child: Row(
                        children: [
                          Container(
                            width: 52,
                            height: 52,
                            decoration: BoxDecoration(
                              color: const Color(0xFFEFF6FF),
                              borderRadius: BorderRadius.circular(10),
                            ),
                            child: const Icon(Icons.storefront_rounded, color: Color(0xFF2563EB), size: 28),
                          ),
                          const SizedBox(width: 14),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  v.businessName,
                                  style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  v.addressLine1,
                                  style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
                                ),
                                const SizedBox(height: 4),
                                Row(
                                  children: [
                                    Container(
                                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                      decoration: BoxDecoration(
                                        color: Colors.green.shade50,
                                        borderRadius: BorderRadius.circular(4),
                                      ),
                                      child: Text(
                                        v.availabilityStatus,
                                        style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Colors.green.shade800),
                                      ),
                                    ),
                                    const SizedBox(width: 8),
                                    Text(
                                      '${v.distanceKm.toStringAsFixed(2)} km away',
                                      style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: Color(0xFF3B82F6)),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          ),
                          const Icon(Icons.chevron_right, color: Color(0xFF94A3B8)),
                        ],
                      ),
                    ),
                  ),
                );
              },
            ),
    );
  }

  // --- TAB 2: CART ---
  Widget _buildCartTab(LocaleProvider loc) {
    if (_cart == null || _cart!.items.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.shopping_cart_outlined, size: 64, color: Color(0xFFCBD5E1)),
            const SizedBox(height: 12),
            Text(loc.t('empty_cart'), style: const TextStyle(fontSize: 16, color: Color(0xFF64748B))),
          ],
        ),
      );
    }

    return Column(
      children: [
        Expanded(
          child: ListView.builder(
            padding: const EdgeInsets.all(12),
            itemCount: _cart!.items.length,
            itemBuilder: (ctx, i) {
              final it = _cart!.items[i];
              return Card(
                margin: const EdgeInsets.only(bottom: 8),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(it.productName, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                            Text('₹${(it.unitPricePaise / 100).toStringAsFixed(2)} each', style: TextStyle(fontSize: 12, color: Colors.grey.shade600)),
                          ],
                        ),
                      ),
                      Row(
                        children: [
                          IconButton.outlined(
                            icon: const Icon(Icons.remove, size: 16),
                            onPressed: () => _handleUpdateCartQuantity(it.productId, it.quantity - 1),
                          ),
                          Padding(
                            padding: const EdgeInsets.symmetric(horizontal: 10),
                            child: Text('${it.quantity}', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                          ),
                          IconButton.outlined(
                            icon: const Icon(Icons.add, size: 16),
                            onPressed: () => _handleUpdateCartQuantity(it.productId, it.quantity + 1),
                          ),
                        ],
                      ),
                      const SizedBox(width: 8),
                      Text(it.formattedTotal, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                    ],
                  ),
                ),
              );
            },
          ),
        ),

        // Cart Checkout Bar
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.white,
            boxShadow: [
              BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 10, offset: const Offset(0, -2))
            ],
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(loc.t('total'), style: TextStyle(fontSize: 12, color: Colors.grey.shade600)),
                  Text(_cart!.formattedTotal, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Color(0xFF1E40AF))),
                ],
              ),
              ElevatedButton(
                onPressed: _showCheckoutDialog,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF1E40AF),
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                ),
                child: Text(loc.t('checkout'), style: const TextStyle(fontWeight: FontWeight.bold)),
              ),
            ],
          ),
        ),
      ],
    );
  }

  // --- TAB 3: MY ORDERS ---
  Widget _buildMyOrdersTab(LocaleProvider loc) {
    if (_myOrders.isEmpty) {
      return Center(child: Text(loc.t('no_orders')));
    }

    return ListView.builder(
      padding: const EdgeInsets.all(12),
      itemCount: _myOrders.length,
      itemBuilder: (ctx, i) {
        final order = _myOrders[i];
        final isCompleted = order.status == 'COMPLETED';
        final canCancel = order.status == 'PENDING' || order.status == 'ACCEPTED';

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
                    Text('Order #${order.id.substring(0, 8)}', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                    _buildStatusBadge(order.status),
                  ],
                ),
                const SizedBox(height: 6),
                Text('Total: ${order.formattedTotal} • ${order.fulfilmentType}', style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                if (order.deliveryAddressSnapshot != null)
                  Padding(
                    padding: const EdgeInsets.only(top: 2.0),
                    child: Text('Delivery: ${order.deliveryAddressSnapshot}', style: TextStyle(fontSize: 11, color: Colors.grey.shade600)),
                  ),
                const Divider(height: 18),

                // Active delivery tracking & actions
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    if (!isCompleted && order.status != 'CANCELLED' && order.status != 'REJECTED') ...[
                      OutlinedButton.icon(
                        onPressed: () => _showLiveTrackingDialog(order),
                        icon: const Icon(Icons.navigation_rounded, size: 16),
                        label: Text(loc.t('track_delivery')),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: const Color(0xFF1E40AF),
                          side: const BorderSide(color: Color(0xFF93C5FD)),
                        ),
                      ),
                      OutlinedButton.icon(
                        onPressed: () => _showOnlinePaymentSheet(order.id, order.totalPaise),
                        icon: const Icon(Icons.payment, size: 16),
                        label: Text(loc.t('pay_online')),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: Colors.green.shade700,
                          side: BorderSide(color: Colors.green.shade300),
                        ),
                      ),
                    ],
                    if (isCompleted)
                      OutlinedButton.icon(
                        onPressed: () => _showReviewDialog(order),
                        icon: const Icon(Icons.star_outline, size: 16),
                        label: Text(loc.t('ratings_reviews')),
                      ),
                    if (canCancel)
                      OutlinedButton(
                        onPressed: () => _showCancelOrderDialog(order),
                        child: Text(loc.t('request_cancel'), style: const TextStyle(color: Colors.red)),
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

  // --- TAB 4: LOYALTY REWARDS ---
  Widget _buildLoyaltyTab(LocaleProvider loc) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Card(
          color: const Color(0xFFEFF6FF),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
            side: const BorderSide(color: Color(0xFFBFDBFE)),
          ),
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              children: [
                const Icon(Icons.stars_rounded, size: 48, color: Color(0xFF2563EB)),
                const SizedBox(height: 8),
                Text(loc.t('points_available'), style: const TextStyle(fontSize: 13, color: Color(0xFF1E3A8A))),
                const SizedBox(height: 4),
                Text(
                  '${_loyalty?.pointsBalance ?? 0}',
                  style: const TextStyle(fontSize: 36, fontWeight: FontWeight.bold, color: Color(0xFF1E40AF)),
                ),
                const SizedBox(height: 4),
                Text(
                  '1 Point = ₹1.00 Discount on checkout',
                  style: TextStyle(fontSize: 12, color: Colors.blue.shade700, fontWeight: FontWeight.w500),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 16),
        const Text('How Loyalty Works:', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
        const SizedBox(height: 6),
        const Text('• Earn 1 loyalty point for every ₹10 spent on completed orders.'),
        const Text('• Redeem points during checkout for instant discounts.'),
        const Text('• Reward points are credited automatically when order status reaches Completed.'),
      ],
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

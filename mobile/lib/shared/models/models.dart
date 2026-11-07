class UserModel {
  final String id;
  final String phone;
  final String status;
  final List<String> roles;

  UserModel({required this.id, required this.phone, required this.status, required this.roles});

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'] ?? '',
      phone: json['phone'] ?? '',
      status: json['status'] ?? 'ACTIVE',
      roles: (json['roles'] as List<dynamic>?)?.map((e) => e.toString()).toList() ?? [],
    );
  }

  bool get isVendor => roles.contains('VENDOR');
  bool get isCustomer => roles.contains('CUSTOMER');
}

class ProductModel {
  final String id;
  final String vendorId;
  final String name;
  final String? description;
  final int pricePaise;
  final String currency;
  final bool active;

  ProductModel({
    required this.id,
    required this.vendorId,
    required this.name,
    this.description,
    required this.pricePaise,
    required this.currency,
    required this.active,
  });

  factory ProductModel.fromJson(Map<String, dynamic> json) {
    return ProductModel(
      id: json['id'] ?? '',
      vendorId: json['vendor_id'] ?? '',
      name: json['name'] ?? '',
      description: json['description'],
      pricePaise: (json['price_paise'] as num?)?.toInt() ?? 0,
      currency: json['currency'] ?? 'INR',
      active: json['active'] ?? true,
    );
  }

  String get formattedPrice => '₹${(pricePaise / 100).toStringAsFixed(2)}';
}

class InventoryModel {
  final String productId;
  final int onHandQty;
  final int reservedQty;
  final int availableQty;

  InventoryModel({
    required this.productId,
    required this.onHandQty,
    required this.reservedQty,
    required this.availableQty,
  });

  factory InventoryModel.fromJson(Map<String, dynamic> json) {
    return InventoryModel(
      productId: json['product_id'] ?? '',
      onHandQty: (json['on_hand_qty'] as num?)?.toInt() ?? 0,
      reservedQty: (json['reserved_qty'] as num?)?.toInt() ?? 0,
      availableQty: (json['available_qty'] as num?)?.toInt() ?? 0,
    );
  }
}

class CartItemModel {
  final String id;
  final String productId;
  final String productName;
  final int quantity;
  final int unitPricePaise;
  final int lineTotalPaise;

  CartItemModel({
    required this.id,
    required this.productId,
    required this.productName,
    required this.quantity,
    required this.unitPricePaise,
    required this.lineTotalPaise,
  });

  factory CartItemModel.fromJson(Map<String, dynamic> json) {
    return CartItemModel(
      id: json['id'] ?? '',
      productId: json['product_id'] ?? '',
      productName: json['product_name'] ?? 'Product',
      quantity: (json['quantity'] as num?)?.toInt() ?? 1,
      unitPricePaise: (json['unit_price_paise'] as num?)?.toInt() ?? 0,
      lineTotalPaise: (json['line_total_paise'] as num?)?.toInt() ?? 0,
    );
  }

  String get formattedTotal => '₹${(lineTotalPaise / 100).toStringAsFixed(2)}';
}

class CartModel {
  final String id;
  final String customerId;
  final String vendorId;
  final String status;
  final List<CartItemModel> items;

  CartModel({
    required this.id,
    required this.customerId,
    required this.vendorId,
    required this.status,
    required this.items,
  });

  factory CartModel.fromJson(Map<String, dynamic> json) {
    return CartModel(
      id: json['id'] ?? '',
      customerId: json['customer_id'] ?? '',
      vendorId: json['vendor_id'] ?? '',
      status: json['status'] ?? 'ACTIVE',
      items: (json['items'] as List<dynamic>?)
              ?.map((e) => CartItemModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  int get totalPaise => items.fold(0, (sum, it) => sum + it.lineTotalPaise);
  String get formattedTotal => '₹${(totalPaise / 100).toStringAsFixed(2)}';
}

class OrderItemModel {
  final String id;
  final String productNameSnapshot;
  final int unitPricePaise;
  final int quantity;
  final int lineTotalPaise;

  OrderItemModel({
    required this.id,
    required this.productNameSnapshot,
    required this.unitPricePaise,
    required this.quantity,
    required this.lineTotalPaise,
  });

  factory OrderItemModel.fromJson(Map<String, dynamic> json) {
    return OrderItemModel(
      id: json['id'] ?? '',
      productNameSnapshot: json['product_name_snapshot'] ?? '',
      unitPricePaise: (json['unit_price_paise'] as num?)?.toInt() ?? 0,
      quantity: (json['quantity'] as num?)?.toInt() ?? 1,
      lineTotalPaise: (json['line_total_paise'] as num?)?.toInt() ?? 0,
    );
  }
}

class OrderModel {
  final String id;
  final String customerId;
  final String vendorId;
  final String fulfilmentType;
  final String status;
  final int totalPaise;
  final List<OrderItemModel> items;
  final DateTime createdAt;

  OrderModel({
    required this.id,
    required this.customerId,
    required this.vendorId,
    required this.fulfilmentType,
    required this.status,
    required this.totalPaise,
    required this.items,
    required this.createdAt,
  });

  factory OrderModel.fromJson(Map<String, dynamic> json) {
    return OrderModel(
      id: json['id'] ?? '',
      customerId: json['customer_id'] ?? '',
      vendorId: json['vendor_id'] ?? '',
      fulfilmentType: json['fulfilment_type'] ?? 'PICKUP',
      status: json['status'] ?? 'PENDING',
      totalPaise: (json['total_paise'] as num?)?.toInt() ?? 0,
      items: (json['items'] as List<dynamic>?)
              ?.map((e) => OrderItemModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at']) ?? DateTime.now()
          : DateTime.now(),
    );
  }

  String get formattedTotal => '₹${(totalPaise / 100).toStringAsFixed(2)}';
}

class NearbyVendorModel {
  final String vendorId;
  final String businessName;
  final String? description;
  final String phone;
  final String status;
  final String availabilityStatus;
  final double distanceKm;
  final String addressLine1;
  final String city;

  NearbyVendorModel({
    required this.vendorId,
    required this.businessName,
    this.description,
    required this.phone,
    required this.status,
    required this.availabilityStatus,
    required this.distanceKm,
    required this.addressLine1,
    required this.city,
  });

  factory NearbyVendorModel.fromJson(Map<String, dynamic> json) {
    return NearbyVendorModel(
      vendorId: json['vendor_id'] ?? '',
      businessName: json['business_name'] ?? 'Vendor',
      description: json['description'],
      phone: json['phone'] ?? '',
      status: json['status'] ?? 'ACTIVE',
      availabilityStatus: json['availability_status'] ?? 'OPEN',
      distanceKm: (json['distance_km'] as num?)?.toDouble() ?? 0.0,
      addressLine1: json['address_line1'] ?? '',
      city: json['city'] ?? '',
    );
  }
}

class UdhaarAccountModel {
  final String id;
  final String vendorId;
  final String customerId;
  final String status;
  final int balancePaise;

  UdhaarAccountModel({
    required this.id,
    required this.vendorId,
    required this.customerId,
    required this.status,
    required this.balancePaise,
  });

  factory UdhaarAccountModel.fromJson(Map<String, dynamic> json) {
    return UdhaarAccountModel(
      id: json['id'] ?? '',
      vendorId: json['vendor_id'] ?? '',
      customerId: json['customer_id'] ?? '',
      status: json['status'] ?? 'ACTIVE',
      balancePaise: (json['balance_paise'] as num?)?.toInt() ?? 0,
    );
  }

  String get formattedBalance => '₹${(balancePaise / 100).toStringAsFixed(2)}';
}

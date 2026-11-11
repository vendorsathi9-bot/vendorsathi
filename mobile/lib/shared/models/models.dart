class UserModel {
  final String id;
  final String phone;
  final String? email;
  final String status;
  final String preferredLanguage;
  final List<String> roles;

  UserModel({
    required this.id,
    required this.phone,
    this.email,
    required this.status,
    this.preferredLanguage = 'en',
    required this.roles,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'] ?? '',
      phone: json['phone'] ?? '',
      email: json['email'],
      status: json['status'] ?? 'ACTIVE',
      preferredLanguage: json['preferred_language'] ?? 'en',
      roles: (json['roles'] as List<dynamic>?)?.map((e) => e.toString()).toList() ?? [],
    );
  }

  bool get isVendor => roles.contains('VENDOR');
  bool get isCustomer => roles.contains('CUSTOMER');
  bool get isDeliveryAgent => roles.contains('DELIVERY_AGENT');
  bool get isAdmin => roles.contains('ADMIN') || roles.contains('SUPER_ADMIN');
}

class CategoryModel {
  final String id;
  final String name;
  final bool active;

  CategoryModel({required this.id, required this.name, required this.active});

  factory CategoryModel.fromJson(Map<String, dynamic> json) {
    return CategoryModel(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      active: json['active'] ?? true,
    );
  }
}

class ProductModel {
  final String id;
  final String vendorId;
  final String? categoryId;
  final String name;
  final String? description;
  final int pricePaise;
  final int costPricePaise;
  final String currency;
  final String? imageObjectKey;
  final bool active;

  ProductModel({
    required this.id,
    required this.vendorId,
    this.categoryId,
    required this.name,
    this.description,
    required this.pricePaise,
    this.costPricePaise = 0,
    required this.currency,
    this.imageObjectKey,
    required this.active,
  });

  factory ProductModel.fromJson(Map<String, dynamic> json) {
    return ProductModel(
      id: json['id'] ?? '',
      vendorId: json['vendor_id'] ?? '',
      categoryId: json['category_id'],
      name: json['name'] ?? '',
      description: json['description'],
      pricePaise: (json['price_paise'] as num?)?.toInt() ?? 0,
      costPricePaise: (json['cost_price_paise'] as num?)?.toInt() ?? 0,
      currency: json['currency'] ?? 'INR',
      imageObjectKey: json['image_object_key'],
      active: json['active'] ?? true,
    );
  }

  String get formattedPrice => '₹${(pricePaise / 100).toStringAsFixed(2)}';
  String get formattedCostPrice => '₹${(costPricePaise / 100).toStringAsFixed(2)}';
  double get marginPercent => pricePaise > 0
      ? ((pricePaise - costPricePaise) / pricePaise * 100).clamp(0, 100)
      : 0.0;
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
  int get pricePaise => unitPricePaise;
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
  String? get vendorName => null;
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
  final int subtotalPaise;
  final int discountPaise;
  final int totalPaise;
  final String? deliveryAddressSnapshot;
  final List<OrderItemModel> items;
  final DateTime createdAt;

  OrderModel({
    required this.id,
    required this.customerId,
    required this.vendorId,
    required this.fulfilmentType,
    required this.status,
    this.subtotalPaise = 0,
    this.discountPaise = 0,
    required this.totalPaise,
    this.deliveryAddressSnapshot,
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
      subtotalPaise: (json['subtotal_paise'] as num?)?.toInt() ?? 0,
      discountPaise: (json['discount_paise'] as num?)?.toInt() ?? 0,
      totalPaise: (json['total_paise'] as num?)?.toInt() ?? 0,
      deliveryAddressSnapshot: json['delivery_address_snapshot'],
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
  String get formattedSubtotal => '₹${(subtotalPaise / 100).toStringAsFixed(2)}';
  String get formattedDiscount => '₹${(discountPaise / 100).toStringAsFixed(2)}';
}

class NearbyVendorModel {
  final String vendorId;
  final String businessName;
  final String? description;
  final String phone;
  final String status;
  final String availabilityStatus;
  final double distanceKm;
  final double latitude;
  final double longitude;
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
    this.latitude = 0.0,
    this.longitude = 0.0,
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
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      addressLine1: json['address_line1'] ?? '',
      city: json['city'] ?? '',
    );
  }

  bool get isOpen => availabilityStatus.toUpperCase() == 'AVAILABLE' || availabilityStatus.toUpperCase() == 'OPEN';
  String get categoryName => description ?? 'General Store';
}

class UdhaarAccountModel {
  final String id;
  final String vendorId;
  final String customerId;
  final String status;
  final int balancePaise;
  final String? customerName;

  UdhaarAccountModel({
    required this.id,
    required this.vendorId,
    required this.customerId,
    required this.status,
    required this.balancePaise,
    this.customerName,
  });

  factory UdhaarAccountModel.fromJson(Map<String, dynamic> json) {
    String? name = json['customer_name'] as String?;
    final phone = json['customer_phone'] as String?;
    final customerId = json['customer_id'] as String? ?? '';
    final id = json['id'] as String? ?? '';

    String identifier;
    if (name != null && name.trim().isNotEmpty) {
      identifier = name.trim();
    } else if (phone != null && phone.trim().isNotEmpty) {
      final p = phone.trim();
      identifier = p.length >= 8 ? 'Customer (${p.substring(0, 5)}...${p.substring(p.length - 2)})' : 'Customer ($p)';
    } else if (customerId.isNotEmpty) {
      identifier = 'Customer #${customerId.length > 6 ? customerId.substring(0, 6) : customerId}';
    } else {
      identifier = 'Customer #${id.length > 6 ? id.substring(0, 6) : id}';
    }

    return UdhaarAccountModel(
      id: id,
      vendorId: json['vendor_id'] ?? '',
      customerId: customerId,
      status: json['status'] ?? 'ACTIVE',
      balancePaise: (json['balance_paise'] as num?)?.toInt() ?? 0,
      customerName: identifier,
    );
  }

  String get formattedBalance => '₹${(balancePaise / 100).toStringAsFixed(2)}';
}

class UdhaarEntryModel {
  final String id;
  final String accountId;
  final String entryType;
  final int amountPaise;
  final String? notes;
  final String createdAt;

  UdhaarEntryModel({
    required this.id,
    required this.accountId,
    required this.entryType,
    required this.amountPaise,
    this.notes,
    required this.createdAt,
  });

  factory UdhaarEntryModel.fromJson(Map<String, dynamic> json) {
    return UdhaarEntryModel(
      id: json['id'] ?? '',
      accountId: json['account_id'] ?? '',
      entryType: json['entry_type'] ?? 'CREDIT',
      amountPaise: (json['amount_paise'] as num?)?.toInt() ?? 0,
      notes: json['notes'] ?? json['note'],
      createdAt: json['created_at'] ?? '',
    );
  }
}

class CustomerAddressModel {
  final String id;
  final String customerId;
  final String addressLine1;
  final String? addressLine2;
  final String city;
  final String state;
  final String? postalCode;
  final String? label;
  final bool isDefault;

  CustomerAddressModel({
    required this.id,
    required this.customerId,
    required this.addressLine1,
    this.addressLine2,
    required this.city,
    required this.state,
    this.postalCode,
    this.label,
    required this.isDefault,
  });

  factory CustomerAddressModel.fromJson(Map<String, dynamic> json) {
    return CustomerAddressModel(
      id: json['id'] ?? '',
      customerId: json['customer_id'] ?? '',
      addressLine1: json['address_line1'] ?? '',
      addressLine2: json['address_line2'],
      city: json['city'] ?? '',
      state: json['state'] ?? 'Karnataka',
      postalCode: json['postal_code'],
      label: json['label'],
      isDefault: json['is_default'] ?? false,
    );
  }

  String get fullAddress => '$addressLine1, $city, $state ${postalCode ?? ''}'.trim();
}

class ReviewModel {
  final String id;
  final String? orderId;
  final String customerId;
  final String vendorId;
  final String customerName;
  final int rating;
  final String? comment;
  final DateTime createdAt;

  ReviewModel({
    required this.id,
    this.orderId,
    required this.customerId,
    required this.vendorId,
    required this.customerName,
    required this.rating,
    this.comment,
    required this.createdAt,
  });

  factory ReviewModel.fromJson(Map<String, dynamic> json) {
    return ReviewModel(
      id: json['id'] ?? '',
      orderId: json['order_id'],
      customerId: json['customer_id'] ?? '',
      vendorId: json['vendor_id'] ?? '',
      customerName: json['customer_name'] ?? 'Customer',
      rating: (json['rating'] as num?)?.toInt() ?? 5,
      comment: json['comment'],
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at']) ?? DateTime.now()
          : DateTime.now(),
    );
  }
}

class VendorReviewSummaryModel {
  final double averageRating;
  final int totalReviews;
  final List<ReviewModel> reviews;

  VendorReviewSummaryModel({
    required this.averageRating,
    required this.totalReviews,
    required this.reviews,
  });

  factory VendorReviewSummaryModel.fromJson(Map<String, dynamic> json) {
    return VendorReviewSummaryModel(
      averageRating: (json['average_rating'] as num?)?.toDouble() ?? 0.0,
      totalReviews: (json['total_reviews'] as num?)?.toInt() ?? 0,
      reviews: (json['reviews'] as List<dynamic>?)
              ?.map((e) => ReviewModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class LoyaltySummaryModel {
  final int pointsBalance;
  final List<dynamic> transactions;

  LoyaltySummaryModel({
    required this.pointsBalance,
    required this.transactions,
  });

  factory LoyaltySummaryModel.fromJson(Map<String, dynamic> json) {
    return LoyaltySummaryModel(
      pointsBalance: (json['points_balance'] as num?)?.toInt() ?? 0,
      transactions: (json['transactions'] as List<dynamic>?) ?? [],
    );
  }

  int get totalSpendPaise => pointsBalance * 100;
}

class ExpenseModel {
  final String id;
  final String title;
  final String category;
  final int amountPaise;
  final String expenseDate;

  ExpenseModel({
    required this.id,
    required this.title,
    required this.category,
    required this.amountPaise,
    required this.expenseDate,
  });

  factory ExpenseModel.fromJson(Map<String, dynamic> json) {
    return ExpenseModel(
      id: json['id'] ?? '',
      title: json['title'] ?? '',
      category: json['category'] ?? 'OPERATIONAL',
      amountPaise: (json['amount_paise'] as num?)?.toInt() ?? 0,
      expenseDate: json['expense_date'] ?? '',
    );
  }

  String get formattedAmount => '₹${(amountPaise / 100).toStringAsFixed(2)}';
}

class SalesTrendModel {
  final String date;
  final int orderCount;
  final int salesPaise;
  final int profitPaise;

  SalesTrendModel({
    required this.date,
    required this.orderCount,
    required this.salesPaise,
    required this.profitPaise,
  });

  factory SalesTrendModel.fromJson(Map<String, dynamic> json) {
    return SalesTrendModel(
      date: json['date'] as String? ?? '',
      orderCount: (json['order_count'] as num?)?.toInt() ?? 0,
      salesPaise: (json['sales_paise'] as num?)?.toInt() ?? 0,
      profitPaise: (json['profit_paise'] as num?)?.toInt() ?? 0,
    );
  }

  double get salesRupees => salesPaise / 100.0;
  double get profitRupees => profitPaise / 100.0;
  String get formattedSales => '₹${salesRupees.toStringAsFixed(2)}';
  String get formattedProfit => '₹${profitRupees.toStringAsFixed(2)}';

  String get shortDate {
    if (date.length >= 10) {
      final parts = date.split('-');
      if (parts.length == 3) {
        return '${parts[2]}/${parts[1]}';
      }
    }
    return date;
  }
}

class DetailedAnalyticsModel {
  final int dailySalesPaise;
  final int weeklySalesPaise;
  final int monthlySalesPaise;
  final int estimatedGrossProfitPaise;
  final int totalExpensesPaise;
  final int estimatedNetProfitPaise;
  final List<dynamic> productProfitMargins;
  final List<dynamic> bestSellingProducts;
  final List<dynamic> lowStockProducts;
  final List<SalesTrendModel> salesTrends;

  DetailedAnalyticsModel({
    required this.dailySalesPaise,
    required this.weeklySalesPaise,
    required this.monthlySalesPaise,
    required this.estimatedGrossProfitPaise,
    required this.totalExpensesPaise,
    required this.estimatedNetProfitPaise,
    required this.productProfitMargins,
    required this.bestSellingProducts,
    required this.lowStockProducts,
    required this.salesTrends,
  });

  factory DetailedAnalyticsModel.fromJson(Map<String, dynamic> json) {
    return DetailedAnalyticsModel(
      dailySalesPaise: (json['daily_sales_paise'] as num?)?.toInt() ?? 0,
      weeklySalesPaise: (json['weekly_sales_paise'] as num?)?.toInt() ?? 0,
      monthlySalesPaise: (json['monthly_sales_paise'] as num?)?.toInt() ?? 0,
      estimatedGrossProfitPaise: (json['estimated_gross_profit_paise'] as num?)?.toInt() ?? 0,
      totalExpensesPaise: (json['total_expenses_paise'] as num?)?.toInt() ?? 0,
      estimatedNetProfitPaise: (json['estimated_net_profit_paise'] as num?)?.toInt() ?? 0,
      productProfitMargins: json['product_profit_margins'] as List<dynamic>? ?? [],
      bestSellingProducts: json['best_selling_products'] as List<dynamic>? ?? [],
      lowStockProducts: json['low_stock_products'] as List<dynamic>? ?? [],
      salesTrends: (json['sales_trends'] as List<dynamic>?)
              ?.map((e) => SalesTrendModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  String get formattedDaily => '₹${(dailySalesPaise / 100).toStringAsFixed(2)}';
  String get formattedWeekly => '₹${(weeklySalesPaise / 100).toStringAsFixed(2)}';
  String get formattedMonthly => '₹${(monthlySalesPaise / 100).toStringAsFixed(2)}';
  String get formattedGrossProfit => '₹${(estimatedGrossProfitPaise / 100).toStringAsFixed(2)}';
  String get formattedNetProfit => '₹${(estimatedNetProfitPaise / 100).toStringAsFixed(2)}';
  String get formattedExpenses => '₹${(totalExpensesPaise / 100).toStringAsFixed(2)}';

  int get grossSalesPaise => dailySalesPaise > 0 ? dailySalesPaise : weeklySalesPaise;
  int get confirmedCollectionsPaise => estimatedGrossProfitPaise;
  int get outstandingUdhaarPaise => totalExpensesPaise;
  int get completedOrdersCount => salesTrends.fold<int>(0, (sum, t) => sum + t.orderCount);
  List<SalesTrendModel> get trends => salesTrends;
}

class VendorSummaryModel {
  final int totalOrders;
  final int completedOrders;
  final int cancelledOrders;
  final int pendingOrders;
  final int grossSalesPaise;
  final int confirmedCollectionsPaise;
  final int pendingCollectionsPaise;
  final int outstandingUdhaarPaise;

  VendorSummaryModel({
    required this.totalOrders,
    required this.completedOrders,
    required this.cancelledOrders,
    required this.pendingOrders,
    required this.grossSalesPaise,
    required this.confirmedCollectionsPaise,
    required this.pendingCollectionsPaise,
    required this.outstandingUdhaarPaise,
  });

  factory VendorSummaryModel.fromJson(Map<String, dynamic> json) {
    return VendorSummaryModel(
      totalOrders: (json['total_orders'] as num?)?.toInt() ?? 0,
      completedOrders: (json['completed_orders'] as num?)?.toInt() ?? 0,
      cancelledOrders: (json['cancelled_orders'] as num?)?.toInt() ?? 0,
      pendingOrders: (json['pending_orders'] as num?)?.toInt() ?? 0,
      grossSalesPaise: (json['gross_sales_paise'] as num?)?.toInt() ?? 0,
      confirmedCollectionsPaise: (json['confirmed_collections_paise'] as num?)?.toInt() ?? 0,
      pendingCollectionsPaise: (json['pending_collections_paise'] as num?)?.toInt() ?? 0,
      outstandingUdhaarPaise: (json['outstanding_udhaar_paise'] as num?)?.toInt() ?? 0,
    );
  }
}

class DemandForecastItemModel {
  final String productId;
  final String productName;
  final int currentStock;
  final double historicalDailyAvg;
  final int predicted7DayDemand;
  final int suggestedRestockQty;
  final String confidence;
  final String explanation;

  DemandForecastItemModel({
    required this.productId,
    required this.productName,
    required this.currentStock,
    required this.historicalDailyAvg,
    required this.predicted7DayDemand,
    required this.suggestedRestockQty,
    required this.confidence,
    required this.explanation,
  });

  factory DemandForecastItemModel.fromJson(Map<String, dynamic> json) {
    return DemandForecastItemModel(
      productId: json['product_id'] ?? '',
      productName: json['product_name'] ?? '',
      currentStock: (json['current_stock'] as num?)?.toInt() ?? 0,
      historicalDailyAvg: (json['historical_daily_avg'] as num?)?.toDouble() ?? 0.0,
      predicted7DayDemand: (json['predicted_7day_demand'] as num?)?.toInt() ?? 0,
      suggestedRestockQty: (json['suggested_restock_qty'] as num?)?.toInt() ?? 0,
      confidence: json['confidence'] ?? 'LOW',
      explanation: json['explanation'] ?? '',
    );
  }
}

class DemandForecastResponseModel {
  final String vendorId;
  final int totalProductsTracked;
  final bool hasSufficientHistory;
  final String generalInsight;
  final List<DemandForecastItemModel> predictions;
  final String methodology;

  DemandForecastResponseModel({
    required this.vendorId,
    required this.totalProductsTracked,
    required this.hasSufficientHistory,
    required this.generalInsight,
    required this.predictions,
    this.methodology = '7-Day Exponential Moving Average',
  });

  factory DemandForecastResponseModel.fromJson(Map<String, dynamic> json) {
    return DemandForecastResponseModel(
      vendorId: json['vendor_id'] ?? '',
      totalProductsTracked: (json['total_products_tracked'] as num?)?.toInt() ?? 0,
      hasSufficientHistory: json['has_sufficient_history'] ?? false,
      generalInsight: json['general_insight'] ?? '',
      predictions: (json['forecasts'] as List<dynamic>?)
              ?.map((e) => DemandForecastItemModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      methodology: json['methodology'] ?? '7-Day Moving Average',
    );
  }
}

class WeatherModel {
  final String city;
  final double temperatureCelsius;
  final String condition;
  final String conditionDescription;
  final int humidityPercent;
  final int rainProbability;
  final double windSpeedKmh;
  final List<String> vendorPlanningTips;

  WeatherModel({
    required this.city,
    required this.temperatureCelsius,
    required this.condition,
    required this.conditionDescription,
    required this.humidityPercent,
    required this.rainProbability,
    required this.windSpeedKmh,
    required this.vendorPlanningTips,
  });

  factory WeatherModel.fromJson(Map<String, dynamic> json) {
    return WeatherModel(
      city: json['city'] ?? 'Bengaluru',
      temperatureCelsius: (json['temperature_celsius'] as num?)?.toDouble() ?? 27.0,
      condition: json['condition'] ?? 'Clear',
      conditionDescription: json['condition_description'] ?? 'Fair weather',
      humidityPercent: (json['humidity_percent'] as num?)?.toInt() ?? 60,
      rainProbability: (json['rain_probability'] as num?)?.toInt() ?? 10,
      windSpeedKmh: (json['wind_speed_kmh'] as num?)?.toDouble() ?? 12.0,
      vendorPlanningTips: (json['vendor_planning_tips'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          [],
    );
  }

  double get windSpeedKmph => windSpeedKmh;
}

class DeliveryAssignmentModel {
  final String id;
  final String orderId;
  final String agentUserId;
  final String status;
  final String assignedAt;
  final String? acceptedAt;
  final String? pickedUpAt;
  final String? deliveredAt;
  final String customerName;
  final String customerPhone;
  final String deliveryAddressLine1;
  final String deliveryCity;
  final double deliveryLat;
  final double deliveryLng;
  final int orderTotalPaise;
  final String orderStatus;
  final String vendorName;
  final String vendorPhone;
  final double vendorLat;
  final double vendorLng;

  DeliveryAssignmentModel({
    required this.id,
    required this.orderId,
    required this.agentUserId,
    required this.status,
    required this.assignedAt,
    this.acceptedAt,
    this.pickedUpAt,
    this.deliveredAt,
    this.customerName = '',
    this.customerPhone = '',
    this.deliveryAddressLine1 = '',
    this.deliveryCity = '',
    this.deliveryLat = 0.0,
    this.deliveryLng = 0.0,
    this.orderTotalPaise = 0,
    this.orderStatus = '',
    this.vendorName = '',
    this.vendorPhone = '',
    this.vendorLat = 0.0,
    this.vendorLng = 0.0,
  });

  factory DeliveryAssignmentModel.fromJson(Map<String, dynamic> json) {
    return DeliveryAssignmentModel(
      id: json['id'] ?? '',
      orderId: json['order_id'] ?? '',
      agentUserId: json['agent_user_id'] ?? '',
      status: json['status'] ?? 'ASSIGNED',
      assignedAt: json['assigned_at'] ?? '',
      acceptedAt: json['accepted_at'],
      pickedUpAt: json['picked_up_at'],
      deliveredAt: json['delivered_at'],
      customerName: json['customer_name'] ?? '',
      customerPhone: json['customer_phone'] ?? '',
      deliveryAddressLine1: json['delivery_address_line1'] ?? '',
      deliveryCity: json['delivery_city'] ?? '',
      deliveryLat: (json['delivery_lat'] as num?)?.toDouble() ?? 0.0,
      deliveryLng: (json['delivery_lng'] as num?)?.toDouble() ?? 0.0,
      orderTotalPaise: (json['order_total_paise'] as num?)?.toInt() ?? 0,
      orderStatus: json['order_status'] ?? '',
      vendorName: json['vendor_name'] ?? '',
      vendorPhone: json['vendor_phone'] ?? '',
      vendorLat: (json['vendor_lat'] as num?)?.toDouble() ?? 0.0,
      vendorLng: (json['vendor_lng'] as num?)?.toDouble() ?? 0.0,
    );
  }

  String get pickupAddress => vendorName.isNotEmpty ? '$vendorName ($vendorPhone)' : 'Vendor Location';
  String get deliveryAddress => deliveryAddressLine1;
  String get deliveryAgentId => agentUserId;
}

class RouteModel {
  final double distanceKm;
  final double durationMinutes;
  final double distanceMeters;
  final double durationSeconds;
  final double? approximateStraightLineKm;
  final String distanceType;
  final bool routingAvailable;
  final bool mapsApiAvailable;
  final String directionsUrl;
  final String? encodedPolyline;
  final String attribution;
  final String provider;
  final String? message;
  final List<List<double>> coordinates; // [[lng, lat], ...]

  RouteModel({
    required this.distanceKm,
    required this.durationMinutes,
    this.distanceMeters = 0.0,
    this.durationSeconds = 0.0,
    this.approximateStraightLineKm,
    this.distanceType = 'road',
    required this.routingAvailable,
    required this.mapsApiAvailable,
    required this.directionsUrl,
    this.encodedPolyline,
    this.attribution = '© OpenStreetMap contributors',
    this.provider = 'osrm',
    this.message,
    this.coordinates = const [],
  });

  factory RouteModel.fromJson(Map<String, dynamic> json) {
    final distKm = (json['distance_km'] as num?)?.toDouble() ?? 0.0;
    final durMin = (json['duration_minutes'] as num?)?.toDouble() ?? 0.0;
    final distMeters = (json['distance_meters'] as num?)?.toDouble() ?? (distKm * 1000.0);
    final durSec = (json['duration_seconds'] as num?)?.toDouble() ?? (durMin * 60.0);
    final avail = (json['routing_available'] as bool?) ?? (json['maps_api_available'] as bool?) ?? false;
    final straightKm = (json['approximate_straight_line_km'] as num?)?.toDouble();
    final distType = json['distance_type'] ?? (avail ? 'road' : 'straight_line_estimate');

    List<List<double>> coords = [];
    if (json['geometry'] != null && json['geometry']['coordinates'] is List) {
      for (var pt in json['geometry']['coordinates']) {
        if (pt is List && pt.length >= 2) {
          final lng = (pt[0] as num).toDouble();
          final lat = (pt[1] as num).toDouble();
          coords.add([lng, lat]);
        }
      }
    }

    return RouteModel(
      distanceKm: distKm,
      durationMinutes: durMin,
      distanceMeters: distMeters,
      durationSeconds: durSec,
      approximateStraightLineKm: straightKm,
      distanceType: distType,
      routingAvailable: avail,
      mapsApiAvailable: avail,
      directionsUrl: json['directions_url'] ?? '',
      encodedPolyline: json['encoded_polyline'],
      attribution: json['attribution'] ?? '© OpenStreetMap contributors',
      provider: json['provider'] ?? 'osrm',
      message: json['message'],
      coordinates: coords,
    );
  }
}

class TrackingModel {
  final String orderId;
  final String agentUserId;
  final double latitude;
  final double longitude;
  final String updatedAt;

  TrackingModel({
    required this.orderId,
    required this.agentUserId,
    required this.latitude,
    required this.longitude,
    required this.updatedAt,
  });

  factory TrackingModel.fromJson(Map<String, dynamic> json) {
    return TrackingModel(
      orderId: json['order_id'] ?? '',
      agentUserId: json['agent_user_id'] ?? '',
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      updatedAt: json['updated_at'] ?? '',
    );
  }
}


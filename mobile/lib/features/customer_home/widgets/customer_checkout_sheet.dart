import 'package:flutter/material.dart';
import '../../../core/api/api_client.dart';
import '../../../core/design_system/app_colors.dart';
import '../../../core/design_system/app_decorations.dart';
import '../../../core/design_system/components/app_badges.dart';
import '../../../core/design_system/components/app_buttons.dart';
import '../../../core/design_system/components/app_inputs.dart';
import '../../../core/design_system/components/app_sheets.dart';
import '../../../core/localization/locale_provider.dart';
import '../../../shared/models/models.dart';

class CustomerCheckoutSheet extends StatefulWidget {
  final ApiClient apiClient;
  final CartModel cart;
  final List<CustomerAddressModel> addresses;
  final LoyaltySummaryModel? loyalty;
  final VoidCallback onOrderSuccess;
  final VoidCallback onAddressAdded;

  const CustomerCheckoutSheet({
    super.key,
    required this.apiClient,
    required this.cart,
    required this.addresses,
    this.loyalty,
    required this.onOrderSuccess,
    required this.onAddressAdded,
  });

  @override
  State<CustomerCheckoutSheet> createState() => _CustomerCheckoutSheetState();
}

class _CustomerCheckoutSheetState extends State<CustomerCheckoutSheet> {
  String _fulfilmentType = 'PICKUP';
  String _paymentOption = 'CASH';
  CustomerAddressModel? _selectedAddress;
  bool _redeemLoyalty = false;
  bool _isSubmitting = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    if (widget.addresses.isNotEmpty) {
      _selectedAddress = widget.addresses.firstWhere(
        (a) => a.isDefault,
        orElse: () => widget.addresses.first,
      );
    }
  }

  void _showAddAddressDialog() {
    final loc = LocaleProvider.of(context);
    final line1 = TextEditingController();
    final city = TextEditingController(text: 'Bengaluru');
    final postal = TextEditingController(text: '560001');

    showDialog(
      context: context,
      builder: (dlgCtx) => AlertDialog(
        backgroundColor: AppColors.surface,
        shape: const RoundedRectangleBorder(borderRadius: AppRadii.roundedLg),
        title: Text(loc.t('add_address'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            AppTextField(label: loc.t('address_line1'), controller: line1, hint: 'No 12, 1st Cross, Majestic'),
            const SizedBox(height: 12),
            AppTextField(label: loc.t('city'), controller: city),
            const SizedBox(height: 12),
            AppTextField(label: loc.t('postal_code'), controller: postal),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dlgCtx),
            child: Text(loc.t('cancel'), style: const TextStyle(color: AppColors.textSecondary)),
          ),
          ElevatedButton(
            onPressed: () async {
              if (line1.text.trim().isEmpty) return;
              try {
                await widget.apiClient.post('/customers/addresses', body: {
                  'address_line1': line1.text.trim(),
                  'city': city.text.trim(),
                  'postal_code': postal.text.trim(),
                  'is_default': true,
                });
                if (mounted) {
                  Navigator.pop(dlgCtx);
                  widget.onAddressAdded();
                }
              } catch (e) {
                // Address error
              }
            },
            style: ElevatedButton.styleFrom(backgroundColor: AppColors.primary),
            child: Text(loc.t('save')),
          ),
        ],
      ),
    );
  }

  Future<void> _handlePlaceOrder() async {
    final loc = LocaleProvider.of(context);

    if (_fulfilmentType == 'DELIVERY' && _selectedAddress == null) {
      setState(() => _errorMessage = loc.t('address_required'));
      return;
    }

    setState(() {
      _isSubmitting = true;
      _errorMessage = null;
    });

    try {
      final orderRes = await widget.apiClient.post('/orders', body: {
        'fulfilment_type': _fulfilmentType,
        'delivery_address_id': _fulfilmentType == 'DELIVERY' ? _selectedAddress?.id : null,
      });

      final orderId = orderRes['id'] as String;

      // If user selected cash or manual UPI, record it
      if (_paymentOption == 'CASH') {
        await widget.apiClient.post('/payments', body: {
          'order_id': orderId,
          'amount_paise': widget.cart.totalPaise,
          'payment_method': 'CASH',
        }).catchError((_) => null);
      } else if (_paymentOption == 'UPI') {
        await widget.apiClient.post('/payments', body: {
          'order_id': orderId,
          'amount_paise': widget.cart.totalPaise,
          'payment_method': 'UPI',
          'transaction_reference': 'UPI-${DateTime.now().millisecondsSinceEpoch}',
        }).catchError((_) => null);
      }

      if (mounted) {
        Navigator.pop(context);
        AppFeedback.showSuccess(context, loc.t('order_placed_success'));
        widget.onOrderSuccess();
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = e.toString().replaceAll('ApiException', '').trim();
          _isSubmitting = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);
    final subtotal = widget.cart.totalPaise;
    final pointsBalance = widget.loyalty?.pointsBalance ?? 0;
    final redeemablePoints = pointsBalance > 50 ? 50 : pointsBalance;
    final discount = (_redeemLoyalty && redeemablePoints > 0) ? (redeemablePoints * 100) : 0;
    final total = (subtotal - discount).clamp(0, subtotal);

    return SingleChildScrollView(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (_errorMessage != null)
            Container(
              padding: const EdgeInsets.all(10),
              margin: const EdgeInsets.only(bottom: 12),
              decoration: BoxDecoration(
                color: AppColors.errorBg,
                borderRadius: AppRadii.roundedMd,
                border: Border.all(color: AppColors.error.withValues(alpha: 0.3)),
              ),
              child: Text(_errorMessage!, style: const TextStyle(color: AppColors.errorText, fontSize: 12)),
            ),

          // Fulfilment Segment
          Text(loc.t('fulfilment'), style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: AppFilterChip(
                  label: loc.t('pickup'),
                  icon: Icons.storefront_rounded,
                  isSelected: _fulfilmentType == 'PICKUP',
                  onTap: () => setState(() => _fulfilmentType = 'PICKUP'),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: AppFilterChip(
                  label: loc.t('delivery'),
                  icon: Icons.delivery_dining_rounded,
                  isSelected: _fulfilmentType == 'DELIVERY',
                  onTap: () => setState(() => _fulfilmentType = 'DELIVERY'),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Delivery Address if delivery chosen
          if (_fulfilmentType == 'DELIVERY') ...[
            Text(loc.t('delivery_address'), style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
            const SizedBox(height: 6),
            if (widget.addresses.isEmpty)
              AppSecondaryButton(
                label: loc.t('add_address'),
                icon: Icons.add_location_alt_rounded,
                onPressed: _showAddAddressDialog,
              )
            else
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                decoration: BoxDecoration(
                  color: AppColors.surface,
                  borderRadius: AppRadii.roundedMd,
                  border: Border.all(color: AppColors.border),
                ),
                child: DropdownButtonHideUnderline(
                  child: DropdownButton<CustomerAddressModel>(
                    value: _selectedAddress,
                    isExpanded: true,
                    items: widget.addresses.map((addr) {
                      return DropdownMenuItem(
                        value: addr,
                        child: Text(
                          '${addr.addressLine1}, ${addr.city}',
                          style: const TextStyle(fontSize: 13, color: AppColors.textPrimary),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      );
                    }).toList(),
                    onChanged: (val) => setState(() => _selectedAddress = val),
                  ),
                ),
              ),
            const SizedBox(height: 16),
          ],

          // Payment Option
          Text(loc.t('payment_method_title'), style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: AppFilterChip(
                  label: loc.t('pay_cash'),
                  icon: Icons.money_rounded,
                  isSelected: _paymentOption == 'CASH',
                  onTap: () => setState(() => _paymentOption = 'CASH'),
                ),
              ),
              const SizedBox(width: 6),
              Expanded(
                child: AppFilterChip(
                  label: loc.t('pay_upi'),
                  icon: Icons.qr_code_scanner_rounded,
                  isSelected: _paymentOption == 'UPI',
                  onTap: () => setState(() => _paymentOption = 'UPI'),
                ),
              ),
              const SizedBox(width: 6),
              Expanded(
                child: AppFilterChip(
                  label: loc.t('pay_online'),
                  icon: Icons.credit_card_rounded,
                  isSelected: _paymentOption == 'ONLINE',
                  onTap: () => setState(() => _paymentOption = 'ONLINE'),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Loyalty Redemption Checkbox
          if (pointsBalance > 0) ...[
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: AppColors.mintSoft,
                borderRadius: AppRadii.roundedMd,
                border: Border.all(color: AppColors.mint),
              ),
              child: CheckboxListTile(
                value: _redeemLoyalty,
                onChanged: (v) => setState(() => _redeemLoyalty = v ?? false),
                contentPadding: EdgeInsets.zero,
                dense: true,
                activeColor: AppColors.primary,
                title: Text(
                  '${loc.t('redeem_points')} ($redeemablePoints pts = ₹$redeemablePoints)',
                  style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: AppColors.primaryDark),
                ),
                subtitle: Text(
                  'Available: $pointsBalance pts',
                  style: const TextStyle(fontSize: 11, color: AppColors.textSecondary),
                ),
              ),
            ),
            const SizedBox(height: 16),
          ],

          // Price Summary Box
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: AppColors.surfaceMuted,
              borderRadius: AppRadii.roundedMd,
              border: Border.all(color: AppColors.borderLight),
            ),
            child: Column(
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(loc.t('subtotal'), style: const TextStyle(fontSize: 13, color: AppColors.textSecondary)),
                    Text('₹${(subtotal / 100).toStringAsFixed(2)}', style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                  ],
                ),
                if (discount > 0) ...[
                  const SizedBox(height: 6),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(loc.t('loyalty_discount'), style: const TextStyle(fontSize: 13, color: AppColors.success, fontWeight: FontWeight.w600)),
                      Text('-₹${(discount / 100).toStringAsFixed(2)}', style: const TextStyle(fontSize: 13, color: AppColors.success, fontWeight: FontWeight.bold)),
                    ],
                  ),
                ],
                const Divider(height: 14),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(loc.t('total_amount'), style: const TextStyle(fontSize: 15, fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
                    Text(
                      '₹${(total / 100).toStringAsFixed(2)}',
                      style: const TextStyle(fontSize: 17, fontWeight: FontWeight.w800, color: AppColors.primary),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),

          // Place Order Button
          AppPrimaryButton(
            label: '${loc.t('place_order')} • ₹${(total / 100).toStringAsFixed(2)}',
            icon: Icons.check_circle_rounded,
            isLoading: _isSubmitting,
            onPressed: _handlePlaceOrder,
          ),
          const SizedBox(height: 12),
        ],
      ),
    );
  }
}

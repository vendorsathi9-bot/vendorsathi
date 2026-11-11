import 'package:flutter/material.dart';
import '../../../core/api/api_client.dart';
import '../../../core/design_system/app_colors.dart';
import '../../../core/design_system/app_decorations.dart';
import '../../../core/design_system/components/app_cards.dart';
import '../../../core/design_system/components/app_inputs.dart';
import '../../../core/design_system/components/app_sheets.dart';
import '../../../core/design_system/components/app_states.dart';
import '../../../core/localization/locale_provider.dart';
import '../../../shared/models/models.dart';

class CustomerLoyaltyTab extends StatelessWidget {
  final ApiClient apiClient;
  final LoyaltySummaryModel? loyalty;
  final List<CustomerAddressModel> addresses;
  final VoidCallback onAddressAdded;

  const CustomerLoyaltyTab({
    super.key,
    required this.apiClient,
    this.loyalty,
    required this.addresses,
    required this.onAddressAdded,
  });

  void _showAddAddressDialog(BuildContext context) {
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
                await apiClient.post('/customers/addresses', body: {
                  'address_line1': line1.text.trim(),
                  'city': city.text.trim(),
                  'postal_code': postal.text.trim(),
                  'is_default': addresses.isEmpty,
                });
                Navigator.pop(dlgCtx);
                AppFeedback.showSuccess(context, 'Address added successfully');
                onAddressAdded();
              } catch (_) {}
            },
            style: ElevatedButton.styleFrom(backgroundColor: AppColors.primary),
            child: Text(loc.t('save')),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);
    final points = loyalty?.pointsBalance ?? 0;
    final totalSpentRupees = (points * 10).toStringAsFixed(2);

    final String tierName;
    if (points >= 500) {
      tierName = 'Gold Member';
    } else if (points >= 100) {
      tierName = 'Silver Member';
    } else {
      tierName = 'Bronze Member';
    }

    return ListView(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 100),
      children: [
        // Loyalty Rewards Hero Card
        Container(
          padding: const EdgeInsets.all(20),
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              colors: [AppColors.primaryDark, AppColors.primaryLight],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: AppRadii.roundedXl,
            boxShadow: AppShadows.floating,
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    loc.t('loyalty_rewards'),
                    style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Colors.white70),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                    decoration: BoxDecoration(
                      color: Colors.white.withValues(alpha: 0.2),
                      borderRadius: AppRadii.roundedPill,
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(Icons.stars_rounded, color: Colors.amber, size: 14),
                        const SizedBox(width: 4),
                        Text(tierName, style: const TextStyle(fontSize: 10, color: Colors.white, fontWeight: FontWeight.bold)),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 14),
              Row(
                crossAxisAlignment: CrossAxisAlignment.baseline,
                textBaseline: TextBaseline.alphabetic,
                children: [
                  Text(
                    '$points',
                    style: const TextStyle(fontSize: 36, fontWeight: FontWeight.w900, color: Colors.white),
                  ),
                  const SizedBox(width: 8),
                  const Text('Points', style: TextStyle(fontSize: 14, color: Colors.white70)),
                ],
              ),
              const SizedBox(height: 6),
              Text(
                'Worth ₹$points discount on your next checkout • Total Spent: ₹$totalSpentRupees',
                style: const TextStyle(fontSize: 12, color: Colors.white70),
              ),
            ],
          ),
        ),
        const SizedBox(height: 20),

        // How Loyalty Works
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: AppColors.surface,
            borderRadius: AppRadii.roundedLg,
            border: Border.all(color: AppColors.border),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                loc.t('how_loyalty_works'),
                style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: AppColors.textPrimary),
              ),
              const SizedBox(height: 10),
              _buildRuleRow(Icons.add_shopping_cart_rounded, loc.t('loyalty_rule_1')),
              const SizedBox(height: 8),
              _buildRuleRow(Icons.discount_rounded, loc.t('loyalty_rule_2')),
              const SizedBox(height: 8),
              _buildRuleRow(Icons.bolt_rounded, loc.t('loyalty_rule_3')),
            ],
          ),
        ),
        const SizedBox(height: 20),

        // Saved Addresses Section
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              loc.t('addresses'),
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: AppColors.textPrimary),
            ),
            TextButton.icon(
              onPressed: () => _showAddAddressDialog(context),
              icon: const Icon(Icons.add_rounded, size: 16, color: AppColors.primary),
              label: Text(loc.t('add_address'), style: const TextStyle(fontSize: 12, color: AppColors.primary, fontWeight: FontWeight.bold)),
            ),
          ],
        ),
        const SizedBox(height: 8),

        if (addresses.isEmpty)
          AppEmptyState(
            icon: Icons.location_off_outlined,
            title: 'No addresses saved',
            subtitle: 'Add a delivery address to order to your doorstep',
            actionLabel: loc.t('add_address'),
            onAction: () => _showAddAddressDialog(context),
          )
        else
          ...addresses.map((addr) {
            return AppCard(
              padding: const EdgeInsets.all(12),
              margin: const EdgeInsets.only(bottom: 8),
              child: Row(
                children: [
                  Container(
                    width: 36,
                    height: 36,
                    decoration: BoxDecoration(
                      color: addr.isDefault ? AppColors.mintSoft : AppColors.surfaceMuted,
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      Icons.place_outlined,
                      color: addr.isDefault ? AppColors.primary : AppColors.textSecondary,
                      size: 20,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Text(
                                addr.addressLine1,
                                style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
                              ),
                            ),
                            if (addr.isDefault)
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                decoration: const BoxDecoration(
                                  color: AppColors.mintSoft,
                                  borderRadius: AppRadii.roundedPill,
                                ),
                                child: const Text('Default', style: TextStyle(fontSize: 9, fontWeight: FontWeight.bold, color: AppColors.primaryDark)),
                              ),
                          ],
                        ),
                        Text(
                          '${addr.city} • ${addr.postalCode ?? ''}',
                          style: const TextStyle(fontSize: 12, color: AppColors.textTertiary),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            );
          }),
      ],
    );
  }

  Widget _buildRuleRow(IconData icon, String text) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 16, color: AppColors.primary),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            text,
            style: const TextStyle(fontSize: 12, color: AppColors.textSecondary, height: 1.3),
          ),
        ),
      ],
    );
  }
}

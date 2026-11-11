import 'package:flutter/material.dart';
import '../../../core/api/api_client.dart';
import '../../../core/design_system/app_colors.dart';
import '../../../core/design_system/app_decorations.dart';
import '../../../core/design_system/components/app_buttons.dart';
import '../../../core/design_system/components/app_cards.dart';
import '../../../core/design_system/components/app_inputs.dart';
import '../../../core/design_system/components/app_sheets.dart';
import '../../../core/design_system/components/app_states.dart';
import '../../../core/localization/locale_provider.dart';
import '../../../shared/models/models.dart';

class VendorUdhaarTab extends StatelessWidget {
  final ApiClient apiClient;
  final List<UdhaarAccountModel> accounts;
  final bool isLoading;
  final VoidCallback onRefresh;

  const VendorUdhaarTab({
    super.key,
    required this.apiClient,
    required this.accounts,
    required this.isLoading,
    required this.onRefresh,
  });

  void _showEntryDialog(BuildContext context, UdhaarAccountModel account, String entryType) {
    final loc = LocaleProvider.of(context);
    final amountCtrl = TextEditingController();
    final notesCtrl = TextEditingController();
    bool isSubmitting = false;

    final isRepayment = entryType == 'REPAYMENT';

    showDialog(
      context: context,
      builder: (dlgCtx) => StatefulBuilder(
        builder: (ctx, setDlgState) => AlertDialog(
          backgroundColor: AppColors.surface,
          shape: const RoundedRectangleBorder(borderRadius: AppRadii.roundedLg),
          title: Text(
            isRepayment ? 'Record Repayment' : 'Add Credit (Udhaar)',
            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                'Customer: ${account.customerName ?? 'Customer'}',
                style: const TextStyle(fontSize: 13, color: AppColors.textSecondary, fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 12),
              AppTextField(label: 'Amount in Rupees (₹)', controller: amountCtrl, keyboardType: TextInputType.number, hint: '100.00'),
              const SizedBox(height: 12),
              AppTextField(label: 'Notes / Reason', controller: notesCtrl, hint: isRepayment ? 'Cash repaid' : 'Grocery purchase on credit'),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(dlgCtx),
              child: Text(loc.t('cancel'), style: const TextStyle(color: AppColors.textSecondary)),
            ),
            ElevatedButton(
              onPressed: isSubmitting
                  ? null
                  : () async {
                      final amountVal = double.tryParse(amountCtrl.text.trim()) ?? 0.0;
                      if (amountVal <= 0) return;
                      final amountPaise = (amountVal * 100).toInt();

                      setDlgState(() => isSubmitting = true);
                      try {
                        await apiClient.post('/udhaar/accounts/${account.id}/entries', body: {
                          'entry_type': entryType,
                          'amount_paise': amountPaise,
                          'notes': notesCtrl.text.trim(),
                        });
                        Navigator.pop(dlgCtx);
                        AppFeedback.showSuccess(context, 'Udhaar transaction recorded');
                        onRefresh();
                      } catch (e) {
                        setDlgState(() => isSubmitting = false);
                        AppFeedback.showError(context, 'Error: $e');
                      }
                    },
              style: ElevatedButton.styleFrom(backgroundColor: isRepayment ? AppColors.success : AppColors.primary),
              child: isSubmitting
                  ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                  : Text(loc.t('save')),
            ),
          ],
        ),
      ),
    );
  }

  void _showLedgerHistory(BuildContext context, UdhaarAccountModel account) async {
    List<UdhaarEntryModel> entries = [];
    bool loading = true;

    AppBottomSheet.show(
      context: context,
      title: 'Ledger: ${account.customerName ?? 'Customer'}',
      child: StatefulBuilder(
        builder: (ctx, setSheetState) {
          if (loading) {
            apiClient.get('/udhaar/accounts/${account.id}/entries').then((res) {
              if (res != null) {
                final raw = res is List ? res : (res['data'] as List?);
                if (raw != null) {
                  entries = raw.map((e) => UdhaarEntryModel.fromJson(e)).toList();
                }
              }
              setSheetState(() => loading = false);
            }).catchError((_) {
              setSheetState(() => loading = false);
            });
          }

          if (loading) {
            return const Padding(
              padding: EdgeInsets.all(32),
              child: Center(child: CircularProgressIndicator()),
            );
          }

          if (entries.isEmpty) {
            return const Padding(
              padding: EdgeInsets.all(32),
              child: Center(child: Text('No ledger entries recorded')),
            );
          }

          return ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: entries.length,
            separatorBuilder: (_, __) => const Divider(),
            itemBuilder: (c, i) {
              final entry = entries[i];
              final isCredit = entry.entryType == 'CREDIT';
              final rupees = (entry.amountPaise / 100).toStringAsFixed(2);

              return Row(
                children: [
                  Container(
                    width: 36,
                    height: 36,
                    decoration: BoxDecoration(
                      color: isCredit ? AppColors.errorBg : AppColors.successBg,
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      isCredit ? Icons.arrow_outward_rounded : Icons.arrow_downward_rounded,
                      color: isCredit ? AppColors.error : AppColors.success,
                      size: 18,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          isCredit ? 'Credit Added' : 'Repayment Received',
                          style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: AppColors.textPrimary),
                        ),
                        if (entry.notes != null && entry.notes!.isNotEmpty)
                          Text(entry.notes!, style: const TextStyle(fontSize: 11, color: AppColors.textSecondary)),
                      ],
                    ),
                  ),
                  Text(
                    '${isCredit ? '+' : '-'}₹$rupees',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: isCredit ? AppColors.error : AppColors.success,
                    ),
                  ),
                ],
              );
            },
          );
        },
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);

    if (isLoading) {
      return Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: List.generate(4, (_) => const AppCardSkeleton()),
        ),
      );
    }

    if (accounts.isEmpty) {
      return AppEmptyState(
        icon: Icons.account_balance_wallet_outlined,
        title: loc.t('no_udhaar_accounts'),
        subtitle: 'Customer credit accounts and repayments will appear here',
      );
    }

    int totalOutstandingPaise = accounts.fold(0, (sum, a) => sum + a.balancePaise);
    final totalOutstandingRupees = (totalOutstandingPaise / 100).toStringAsFixed(2);

    return RefreshIndicator(
      onRefresh: () async => onRefresh(),
      color: AppColors.primary,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 100),
        children: [
          // Total Udhaar Banner
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: AppColors.warningBg,
              borderRadius: AppRadii.roundedLg,
              border: Border.all(color: AppColors.warning.withValues(alpha: 0.3)),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Total Outstanding Udhaar', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.warningText)),
                    const SizedBox(height: 2),
                    Text('₹$totalOutstandingRupees', style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w800, color: AppColors.warningText)),
                  ],
                ),
                Text('${accounts.length} ${loc.t('accounts_count')}', style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: AppColors.warningText)),
              ],
            ),
          ),
          const SizedBox(height: 14),

          // Accounts List
          ...accounts.map((acc) {
            final balanceRupees = (acc.balancePaise / 100).toStringAsFixed(2);
            final hasBalance = acc.balancePaise > 0;

            return AppCard(
              padding: const EdgeInsets.all(14),
              margin: const EdgeInsets.only(bottom: 12),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        acc.customerName ?? 'Customer',
                        style: const TextStyle(fontSize: 15, fontWeight: FontWeight.bold, color: AppColors.textPrimary),
                      ),
                      Text(
                        '₹$balanceRupees',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w800,
                          color: hasBalance ? AppColors.warning : AppColors.success,
                        ),
                      ),
                    ],
                  ),
                  const Divider(height: 18),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      TextButton.icon(
                        icon: const Icon(Icons.history_rounded, size: 16),
                        label: const Text('Ledger History', style: TextStyle(fontSize: 12)),
                        onPressed: () => _showLedgerHistory(context, acc),
                      ),
                      Row(
                        children: [
                          AppOutlinedButton(
                            label: '+ Credit',
                            color: AppColors.primary,
                            onPressed: () => _showEntryDialog(context, acc, 'CREDIT'),
                          ),
                          const SizedBox(width: 8),
                          AppSecondaryButton(
                            label: 'Repay',
                            icon: Icons.check_circle_outline_rounded,
                            isFullWidth: false,
                            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                            onPressed: () => _showEntryDialog(context, acc, 'REPAYMENT'),
                          ),
                        ],
                      ),
                    ],
                  ),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }
}

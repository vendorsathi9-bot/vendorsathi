import 'package:flutter/material.dart';
import '../../../core/api/api_client.dart';
import '../../../core/design_system/app_colors.dart';
import '../../../core/design_system/app_decorations.dart';
import '../../../core/design_system/components/app_cards.dart';
import '../../../core/design_system/components/app_states.dart';
import '../../../core/localization/locale_provider.dart';
import '../../../shared/models/models.dart';
import '../../../shared/widgets/sales_trend_chart.dart';

class VendorDashboardTab extends StatelessWidget {
  final ApiClient apiClient;
  final VendorSummaryModel? summary;
  final DetailedAnalyticsModel? analytics;
  final DemandForecastResponseModel? demandForecast;
  final WeatherModel? weather;
  final bool isLoading;
  final VoidCallback onRefresh;
  final VoidCallback onViewOrders;
  final VoidCallback onViewInventory;

  const VendorDashboardTab({
    super.key,
    required this.apiClient,
    this.summary,
    this.analytics,
    this.demandForecast,
    this.weather,
    required this.isLoading,
    required this.onRefresh,
    required this.onViewOrders,
    required this.onViewInventory,
  });

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

    final totalSalesPaise = summary?.grossSalesPaise ?? analytics?.weeklySalesPaise ?? 0;
    final grossSalesRupees = (totalSalesPaise / 100).toStringAsFixed(2);
    final collectionsRupees = ((summary?.confirmedCollectionsPaise ?? 0) / 100).toStringAsFixed(2);
    final udhaarRupees = ((summary?.outstandingUdhaarPaise ?? 0) / 100).toStringAsFixed(2);
    final ordersCount = summary?.completedOrders ?? analytics?.completedOrdersCount ?? 0;

    return RefreshIndicator(
      onRefresh: () async => onRefresh(),
      color: AppColors.primary,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 32),
        children: [
          // 4 Grid KPI Cards (All-Time Business Summary)
          Row(
            children: [
              Expanded(
                child: AppStatCard(
                  label: loc.currentLanguage == 'kn' ? 'ಒಟ್ಟು ಆದಾಯ (ಸಮಗ್ರ)' : 'Total Revenue (All-Time)',
                  value: '₹$grossSalesRupees',
                  icon: Icons.payments_rounded,
                  iconColor: AppColors.primary,
                  iconBgColor: AppColors.mintSoft,
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: AppStatCard(
                  label: loc.t('total_orders_count'),
                  value: '$ordersCount',
                  icon: Icons.shopping_bag_rounded,
                  iconColor: AppColors.info,
                  iconBgColor: AppColors.infoBg,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Row(
            children: [
              Expanded(
                child: AppStatCard(
                  label: loc.t('collections'),
                  value: '₹$collectionsRupees',
                  icon: Icons.account_balance_wallet_rounded,
                  iconColor: AppColors.success,
                  iconBgColor: AppColors.successBg,
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: AppStatCard(
                  label: loc.t('udhaar'),
                  value: '₹$udhaarRupees',
                  icon: Icons.pending_actions_rounded,
                  iconColor: AppColors.warning,
                  iconBgColor: AppColors.warningBg,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Sales Trend Chart Card
          AppCard(
            padding: const EdgeInsets.all(16),
            child: SalesTrendChart(
              trends: analytics?.trends ?? [],
            ),
          ),
          const SizedBox(height: 16),

          // Weather Forecast Card
          if (weather != null)
            AppCard(
              padding: const EdgeInsets.all(14),
              child: Row(
                children: [
                  Container(
                    width: 44,
                    height: 44,
                    decoration: const BoxDecoration(
                      color: AppColors.mintSoft,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.wb_sunny_rounded, color: Colors.orange, size: 24),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '${weather!.condition} • ${weather!.temperatureCelsius.toStringAsFixed(1)}°C',
                          style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: AppColors.textPrimary),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          'Humidity: ${weather!.humidityPercent}% • Wind: ${weather!.windSpeedKmph} km/h',
                          style: const TextStyle(fontSize: 11, color: AppColors.textSecondary),
                        ),
                      ],
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: const BoxDecoration(
                      color: AppColors.surfaceMuted,
                      borderRadius: AppRadii.roundedPill,
                    ),
                    child: const Text('Live Mandi', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: AppColors.textSecondary)),
                  ),
                ],
              ),
            ),
          const SizedBox(height: 16),

          // AI Demand Forecasting Card
          if (demandForecast != null && demandForecast!.predictions.isNotEmpty) ...[
            Text(
              loc.t('forecast'),
              style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w700, color: AppColors.textPrimary),
            ),
            const SizedBox(height: 8),
            AppCard(
              padding: const EdgeInsets.all(14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.auto_graph_rounded, size: 18, color: AppColors.primary),
                          const SizedBox(width: 6),
                          Text(
                            loc.t('calibrated_forecast'),
                            style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: AppColors.textPrimary),
                          ),
                        ],
                      ),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                        decoration: const BoxDecoration(
                          color: AppColors.mintSoft,
                          borderRadius: AppRadii.roundedPill,
                        ),
                        child: Text(
                          demandForecast!.methodology,
                          style: const TextStyle(fontSize: 9, fontWeight: FontWeight.bold, color: AppColors.primaryDark),
                        ),
                      ),
                    ],
                  ),
                  const Divider(height: 18),
                  ...demandForecast!.predictions.take(4).map((p) {
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 8.0),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            p.productName,
                            style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.textPrimary),
                          ),
                          Row(
                            children: [
                              Text(
                                '${loc.t('suggested_restock')}: ',
                                style: const TextStyle(fontSize: 11, color: AppColors.textSecondary),
                              ),
                              Text(
                                '+${p.suggestedRestockQty}',
                                style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: AppColors.primary),
                              ),
                            ],
                          ),
                        ],
                      ),
                    );
                  }),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }
}

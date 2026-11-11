import 'dart:math' as math;
import 'package:flutter/material.dart';
import '../../core/design_system/app_colors.dart';
import '../../core/design_system/app_decorations.dart';
import '../../core/localization/locale_provider.dart';
import '../models/models.dart';

class SalesTrendChart extends StatefulWidget {
  final List<SalesTrendModel> trends;

  const SalesTrendChart({
    super.key,
    required this.trends,
  });

  @override
  State<SalesTrendChart> createState() => _SalesTrendChartState();
}

class _SalesTrendChartState extends State<SalesTrendChart> {
  int? _selectedBarIndex;

  @override
  Widget build(BuildContext context) {
    final loc = LocaleProvider.of(context);
    final trends = widget.trends;

    final totalSalesPaise = trends.fold<int>(0, (sum, t) => sum + t.salesPaise);
    final totalOrders = trends.fold<int>(0, (sum, t) => sum + t.orderCount);
    final hasData = totalSalesPaise > 0 || totalOrders > 0;

    if (!hasData) {
      return Container(
        padding: const EdgeInsets.symmetric(vertical: 24, horizontal: 16),
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: AppRadii.roundedLg,
          border: Border.all(color: AppColors.border),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 48,
              height: 48,
              decoration: const BoxDecoration(
                color: AppColors.mintSoft,
                shape: BoxShape.circle,
              ),
              child: const Icon(Icons.bar_chart_rounded, size: 26, color: AppColors.primary),
            ),
            const SizedBox(height: 12),
            Text(
              loc.t('no_sales_recorded'),
              textAlign: TextAlign.center,
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: AppColors.textPrimary),
            ),
            const SizedBox(height: 4),
            Text(
              loc.t('no_sales_hint'),
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 12, color: AppColors.textSecondary, height: 1.4),
            ),
          ],
        ),
      );
    }

    final maxSales = trends.fold<int>(0, (max, t) => math.max(max, t.salesPaise));
    final effectiveMax = maxSales > 0 ? maxSales : 10000;
    final avgDailyPaise = trends.isNotEmpty ? (totalSalesPaise / trends.length).round() : 0;

    final selectedTrend = (_selectedBarIndex != null && _selectedBarIndex! < trends.length)
        ? trends[_selectedBarIndex!]
        : null;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: AppRadii.roundedLg,
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header Row
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(
                    width: 32,
                    height: 32,
                    decoration: const BoxDecoration(
                      color: AppColors.mintSoft,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.insights_rounded, color: AppColors.primary, size: 18),
                  ),
                  const SizedBox(width: 8),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        loc.t('sales_trends_title'),
                        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15, color: AppColors.textPrimary),
                      ),
                      Text(
                        loc.t('sales_trends_subtitle'),
                        style: const TextStyle(fontSize: 11, color: AppColors.textTertiary),
                      ),
                    ],
                  ),
                ],
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: AppColors.mintSoft,
                  borderRadius: AppRadii.roundedPill,
                  border: Border.all(color: AppColors.mint),
                ),
                child: Text(
                  '₹${(totalSalesPaise / 100).toStringAsFixed(2)}',
                  style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: AppColors.primaryDark),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),

          // Period Summary Tiles
          Row(
            children: [
              _buildSummaryTile(
                label: loc.currentLanguage == 'kn' ? '7-ದಿನಗಳ ಮಾರಾಟ' : '7-Day Total Sales',
                value: '₹${(totalSalesPaise / 100).toStringAsFixed(2)}',
                color: AppColors.primary,
              ),
              const SizedBox(width: 8),
              _buildSummaryTile(
                label: loc.t('total_orders_count'),
                value: '$totalOrders',
                color: AppColors.success,
              ),
              const SizedBox(width: 8),
              _buildSummaryTile(
                label: loc.currentLanguage == 'kn' ? 'ದೈನಂದಿನ ಸರಾಸರಿ' : '7-Day Daily Avg',
                value: '₹${(avgDailyPaise / 100).toStringAsFixed(2)}',
                color: AppColors.accent,
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Responsive Bar Chart Area
          LayoutBuilder(
            builder: (ctx, constraints) {
              const chartHeight = 160.0;
              final availableWidth = constraints.maxWidth;
              final barWidth = math.min(36.0, math.max(16.0, (availableWidth / (trends.length * 1.8))));

              return Container(
                height: chartHeight,
                padding: const EdgeInsets.only(top: 8, bottom: 4),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    for (int i = 0; i < trends.length; i++)
                      _buildBarColumn(
                        trend: trends[i],
                        index: i,
                        barWidth: barWidth,
                        maxSales: effectiveMax,
                        maxBarHeight: 105.0,
                        isSelected: _selectedBarIndex == i,
                      ),
                  ],
                ),
              );
            },
          ),

          // Interactive Selected Bar Detail Footer
          if (selectedTrend != null) ...[
            const Divider(height: 18),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
              decoration: BoxDecoration(
                color: AppColors.surfaceMuted,
                borderRadius: AppRadii.roundedSm,
                border: Border.all(color: AppColors.borderLight),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    selectedTrend.date,
                    style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12, color: AppColors.textPrimary),
                  ),
                  Text(
                    '${loc.t('orders_metric')}: ${selectedTrend.orderCount}',
                    style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.success),
                  ),
                  Text(
                    '${loc.t('revenue_metric')}: ₹${(selectedTrend.salesPaise / 100).toStringAsFixed(2)}',
                    style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: AppColors.primary),
                  ),
                  if (selectedTrend.profitPaise > 0)
                    Text(
                      '${loc.t('profit_metric')}: ₹${(selectedTrend.profitPaise / 100).toStringAsFixed(2)}',
                      style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: AppColors.accent),
                    ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildSummaryTile({
    required String label,
    required String value,
    required Color color,
  }) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.08),
          borderRadius: AppRadii.roundedSm,
          border: Border.all(color: color.withValues(alpha: 0.15)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(fontSize: 10, color: AppColors.textSecondary, fontWeight: FontWeight.w500),
            ),
            const SizedBox(height: 3),
            Text(
              value,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: color),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBarColumn({
    required SalesTrendModel trend,
    required int index,
    required double barWidth,
    required int maxSales,
    required double maxBarHeight,
    required bool isSelected,
  }) {
    final ratio = trend.salesPaise / maxSales;
    final barHeight = math.max(6.0, ratio * maxBarHeight);
    final hasSales = trend.salesPaise > 0;

    return InkWell(
      onTap: () {
        setState(() {
          _selectedBarIndex = (_selectedBarIndex == index) ? null : index;
        });
      },
      borderRadius: AppRadii.roundedSm,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 2.0),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.end,
          children: [
            // Floating value label above active bars
            SizedBox(
              height: 16,
              child: hasSales
                  ? Center(
                      child: Text(
                        '₹${(trend.salesPaise / 100).round()}',
                        style: TextStyle(
                          fontSize: 9,
                          fontWeight: FontWeight.bold,
                          color: isSelected ? AppColors.primary : AppColors.textSecondary,
                        ),
                      ),
                    )
                  : const SizedBox.shrink(),
            ),
            const SizedBox(height: 4),

            // Bar shape
            AnimatedContainer(
              duration: const Duration(milliseconds: 300),
              width: barWidth,
              height: barHeight,
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.bottomCenter,
                  end: Alignment.topCenter,
                  colors: hasSales
                      ? (isSelected
                          ? [AppColors.primaryDark, AppColors.accentLight]
                          : [AppColors.primary, AppColors.accent])
                      : [AppColors.skeletonBase, AppColors.border],
                ),
                borderRadius: const BorderRadius.vertical(top: Radius.circular(5)),
                boxShadow: (hasSales && isSelected)
                    ? [BoxShadow(color: AppColors.primary.withValues(alpha: 0.3), blurRadius: 6, offset: const Offset(0, 2))]
                    : null,
              ),
            ),
            const SizedBox(height: 4),

            // Short date label (DD/MM)
            SizedBox(
              height: 16,
              child: Center(
                child: Text(
                  trend.shortDate,
                  style: TextStyle(
                    fontSize: 10,
                    fontWeight: isSelected ? FontWeight.bold : FontWeight.w500,
                    color: isSelected ? AppColors.primary : AppColors.textTertiary,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

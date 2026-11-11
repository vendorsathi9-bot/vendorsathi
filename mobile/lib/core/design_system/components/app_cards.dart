import 'package:flutter/material.dart';
import '../app_colors.dart';
import '../app_decorations.dart';
import 'app_badges.dart';
import 'app_buttons.dart';

/// Base modern card with rounded corners, subtle border, and soft shadow
class AppCard extends StatelessWidget {
  final Widget child;
  final EdgeInsets? padding;
  final EdgeInsets? margin;
  final VoidCallback? onTap;
  final Color? color;
  final Border? border;
  final BorderRadius? borderRadius;

  const AppCard({
    super.key,
    required this.child,
    this.padding,
    this.margin,
    this.onTap,
    this.color,
    this.border,
    this.borderRadius,
  });

  @override
  Widget build(BuildContext context) {
    final effectiveRadius = borderRadius ?? AppRadii.roundedLg;

    Widget card = Container(
      margin: margin ?? const EdgeInsets.only(bottom: AppSpacing.sm),
      padding: padding ?? const EdgeInsets.all(AppSpacing.md),
      decoration: BoxDecoration(
        color: color ?? AppColors.surface,
        borderRadius: effectiveRadius,
        border: border ?? Border.all(color: AppColors.border, width: 1),
        boxShadow: AppShadows.card,
      ),
      child: child,
    );

    if (onTap != null) {
      card = PressableScale(
        onPressed: onTap,
        child: card,
      );
    }

    return card;
  }
}

/// Dashboard KPI Metric Card
class AppStatCard extends StatelessWidget {
  final String label;
  final String value;
  final IconData icon;
  final String? subtitle;
  final Color? iconColor;
  final Color? iconBgColor;

  const AppStatCard({
    super.key,
    required this.label,
    required this.value,
    required this.icon,
    this.subtitle,
    this.iconColor,
    this.iconBgColor,
  });

  @override
  Widget build(BuildContext context) {
    return AppCard(
      padding: const EdgeInsets.all(14),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Flexible(
                child: Text(
                  label,
                  style: const TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: AppColors.textSecondary,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              Container(
                width: 32,
                height: 32,
                decoration: BoxDecoration(
                  color: iconBgColor ?? AppColors.mintSoft,
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  icon,
                  size: 16,
                  color: iconColor ?? AppColors.primary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            value,
            style: const TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.w800,
              color: AppColors.textPrimary,
              letterSpacing: -0.3,
            ),
          ),
          if (subtitle != null) ...[
            const SizedBox(height: 4),
            Text(
              subtitle!,
              style: const TextStyle(
                fontSize: 11,
                color: AppColors.textTertiary,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

/// Modern Vendor Discovery Card
class AppVendorCard extends StatelessWidget {
  final String businessName;
  final double distanceKm;
  final String address;
  final bool isOpen;
  final String? category;
  final VoidCallback onTap;

  const AppVendorCard({
    super.key,
    required this.businessName,
    required this.distanceKm,
    required this.address,
    required this.isOpen,
    this.category,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return AppCard(
      onTap: onTap,
      padding: const EdgeInsets.all(14),
      margin: const EdgeInsets.only(bottom: 12),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Store Avatar icon with mint background
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              color: AppColors.mintSoft,
              borderRadius: AppRadii.roundedMd,
              border: Border.all(color: AppColors.mint, width: 1),
            ),
            child: const Icon(
              Icons.storefront_rounded,
              color: AppColors.primary,
              size: 24,
            ),
          ),
          const SizedBox(width: 12),
              // Vendor Info
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Expanded(
                          child: Text(
                            businessName,
                            style: const TextStyle(
                              fontSize: 15,
                              fontWeight: FontWeight.w700,
                              color: AppColors.textPrimary,
                              height: 1.25,
                            ),
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        const SizedBox(width: 8),
                        AppStatusBadge.availability(isOpen),
                      ],
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                          decoration: const BoxDecoration(
                            color: AppColors.surfaceMuted,
                            borderRadius: AppRadii.roundedSm,
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const Icon(Icons.near_me_rounded, size: 11, color: AppColors.accent),
                              const SizedBox(width: 3),
                              Text(
                                '${distanceKm.toStringAsFixed(1)} km',
                                style: const TextStyle(
                                  fontSize: 11,
                                  fontWeight: FontWeight.w600,
                                  color: AppColors.accent,
                                ),
                              ),
                            ],
                          ),
                        ),
                        if (category != null && category!.isNotEmpty) ...[
                          const SizedBox(width: 6),
                          Expanded(
                            child: Text(
                              '•  $category',
                              style: const TextStyle(fontSize: 12, color: AppColors.textSecondary),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ],
                      ],
                    ),
                    if (address.isNotEmpty) ...[
                      const SizedBox(height: 4),
                      Text(
                        address,
                        style: const TextStyle(fontSize: 12, color: AppColors.textTertiary, height: 1.25),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ],
                ),
              ),
          const SizedBox(width: 4),
          const Icon(Icons.chevron_right_rounded, color: AppColors.textTertiary, size: 20),
        ],
      ),
    );
  }
}

/// Modern Product Card with quantity controls and price
class AppProductCard extends StatelessWidget {
  final String name;
  final int pricePaise;
  final String? category;
  final int? stockQty;
  final int inCartQty;
  final VoidCallback onAdd;
  final VoidCallback? onRemove;
  final bool isAvailable;

  const AppProductCard({
    super.key,
    required this.name,
    required this.pricePaise,
    this.category,
    this.stockQty,
    this.inCartQty = 0,
    required this.onAdd,
    this.onRemove,
    this.isAvailable = true,
  });

  @override
  Widget build(BuildContext context) {
    final priceRupees = (pricePaise / 100).toStringAsFixed(2);

    return AppCard(
      padding: const EdgeInsets.all(12),
      margin: const EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          // Item placeholder avatar
          Container(
            width: 46,
            height: 46,
            decoration: BoxDecoration(
              color: isAvailable ? AppColors.mintSoft : AppColors.surfaceMuted,
              borderRadius: AppRadii.roundedMd,
              border: Border.all(color: AppColors.borderLight),
            ),
            child: Icon(
              Icons.shopping_bag_outlined,
              color: isAvailable ? AppColors.primary : AppColors.textTertiary,
              size: 22,
            ),
          ),
          const SizedBox(width: 12),
          // Details
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  name,
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: isAvailable ? AppColors.textPrimary : AppColors.textTertiary,
                    height: 1.25,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 2),
                Row(
                  children: [
                    Text(
                      '₹$priceRupees',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w700,
                        color: isAvailable ? AppColors.primary : AppColors.textTertiary,
                      ),
                    ),
                    if (category != null) ...[
                      const SizedBox(width: 6),
                      Flexible(
                        child: Text(
                          '• $category',
                          style: const TextStyle(fontSize: 11, color: AppColors.textTertiary),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ],
                ),
                if (stockQty != null) ...[
                  const SizedBox(height: 2),
                  Text(
                    stockQty! > 0 ? '$stockQty in stock' : 'Out of stock',
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w500,
                      color: stockQty! > 5
                          ? AppColors.textSecondary
                          : (stockQty! > 0 ? AppColors.warning : AppColors.error),
                    ),
                  ),
                ],
              ],
            ),
          ),
          // Cart Action Stepper
          if (!isAvailable)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: const BoxDecoration(
                color: AppColors.surfaceMuted,
                borderRadius: AppRadii.roundedSm,
              ),
              child: const Text('Unavailable', style: TextStyle(fontSize: 11, color: AppColors.textTertiary)),
            )
          else if (inCartQty > 0)
            Container(
              decoration: BoxDecoration(
                color: AppColors.mintSoft,
                borderRadius: AppRadii.roundedPill,
                border: Border.all(color: AppColors.primary.withValues(alpha: 0.3)),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  IconButton(
                    icon: const Icon(Icons.remove, size: 16, color: AppColors.primary),
                    onPressed: onRemove,
                    constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                    padding: EdgeInsets.zero,
                  ),
                  Text(
                    '$inCartQty',
                    style: const TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.bold,
                      color: AppColors.primary,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.add, size: 16, color: AppColors.primary),
                    onPressed: onAdd,
                    constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                    padding: EdgeInsets.zero,
                  ),
                ],
              ),
            )
          else
            PressableScale(
              onPressed: onAdd,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 7),
                decoration: const BoxDecoration(
                  color: AppColors.primary,
                  borderRadius: AppRadii.roundedPill,
                  boxShadow: AppShadows.card,
                ),
                child: const Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.add_rounded, size: 14, color: Colors.white),
                    SizedBox(width: 4),
                    Text(
                      'ADD',
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                        letterSpacing: 0.5,
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
}

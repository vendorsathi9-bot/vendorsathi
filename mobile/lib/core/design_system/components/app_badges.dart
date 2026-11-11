import 'package:flutter/material.dart';
import '../app_colors.dart';
import '../app_decorations.dart';

enum BadgeVariant {
  neutral,
  success,
  warning,
  error,
  info,
  primary,
}

/// Modern rounded status badge with soft background and dark text
class AppStatusBadge extends StatelessWidget {
  final String label;
  final BadgeVariant variant;
  final IconData? icon;
  final bool hasDot;

  const AppStatusBadge({
    super.key,
    required this.label,
    this.variant = BadgeVariant.neutral,
    this.icon,
    this.hasDot = false,
  });

  /// Factory constructors for order states
  factory AppStatusBadge.fromOrderStatus(String status) {
    switch (status.toUpperCase()) {
      case 'PENDING':
        return AppStatusBadge(label: status, variant: BadgeVariant.warning, hasDot: true);
      case 'ACCEPTED':
        return AppStatusBadge(label: status, variant: BadgeVariant.info, icon: Icons.check_circle_outline);
      case 'PREPARING':
        return AppStatusBadge(label: status, variant: BadgeVariant.primary, icon: Icons.soup_kitchen_rounded);
      case 'READY_FOR_PICKUP':
        return const AppStatusBadge(label: 'READY', variant: BadgeVariant.success, icon: Icons.storefront_rounded);
      case 'OUT_FOR_DELIVERY':
        return const AppStatusBadge(label: 'OUT FOR DELIVERY', variant: BadgeVariant.primary, icon: Icons.moped_rounded);
      case 'CANCELLATION_REQUESTED':
        return const AppStatusBadge(label: 'CANCEL REQUESTED', variant: BadgeVariant.warning, icon: Icons.pending_actions_rounded);
      case 'COMPLETED':
      case 'DELIVERED':
        return AppStatusBadge(label: status, variant: BadgeVariant.success, icon: Icons.check_circle_rounded);
      case 'CANCELLED':
      case 'REJECTED':
        return AppStatusBadge(label: status, variant: BadgeVariant.error, icon: Icons.cancel_outlined);
      default:
        return AppStatusBadge(label: status, variant: BadgeVariant.neutral);
    }
  }

  /// Factory for vendor open/closed
  factory AppStatusBadge.availability(bool isOpen) {
    return AppStatusBadge(
      label: isOpen ? 'OPEN' : 'CLOSED',
      variant: isOpen ? BadgeVariant.success : BadgeVariant.error,
      hasDot: true,
    );
  }

  Color get _bgColor {
    switch (variant) {
      case BadgeVariant.success:
        return AppColors.successBg;
      case BadgeVariant.warning:
        return AppColors.warningBg;
      case BadgeVariant.error:
        return AppColors.errorBg;
      case BadgeVariant.info:
        return AppColors.infoBg;
      case BadgeVariant.primary:
        return AppColors.mintSoft;
      case BadgeVariant.neutral:
        return AppColors.surfaceMuted;
    }
  }

  Color get _textColor {
    switch (variant) {
      case BadgeVariant.success:
        return AppColors.successText;
      case BadgeVariant.warning:
        return AppColors.warningText;
      case BadgeVariant.error:
        return AppColors.errorText;
      case BadgeVariant.info:
        return AppColors.infoText;
      case BadgeVariant.primary:
        return AppColors.primary;
      case BadgeVariant.neutral:
        return AppColors.textSecondary;
    }
  }

  Color get _dotColor {
    switch (variant) {
      case BadgeVariant.success:
        return AppColors.success;
      case BadgeVariant.warning:
        return AppColors.warning;
      case BadgeVariant.error:
        return AppColors.error;
      case BadgeVariant.info:
        return AppColors.info;
      case BadgeVariant.primary:
        return AppColors.primary;
      case BadgeVariant.neutral:
        return AppColors.textTertiary;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: _bgColor,
        borderRadius: AppRadii.roundedPill,
        border: Border.all(
          color: _dotColor.withValues(alpha: 0.2),
          width: 0.8,
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (hasDot) ...[
            Container(
              width: 6,
              height: 6,
              decoration: BoxDecoration(
                color: _dotColor,
                shape: BoxShape.circle,
              ),
            ),
            const SizedBox(width: 5),
          ] else if (icon != null) ...[
            Icon(icon, size: 12, color: _textColor),
            const SizedBox(width: 4),
          ],
          Text(
            label,
            style: TextStyle(
              fontSize: 10,
              fontWeight: FontWeight.w700,
              color: _textColor,
              letterSpacing: 0.3,
            ),
          ),
        ],
      ),
    );
  }
}

/// Filter pill chip with selected/unselected states
class AppFilterChip extends StatelessWidget {
  final String label;
  final bool isSelected;
  final VoidCallback onTap;
  final IconData? icon;

  const AppFilterChip({
    super.key,
    required this.label,
    required this.isSelected,
    required this.onTap,
    this.icon,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 7),
        decoration: BoxDecoration(
          color: isSelected ? AppColors.primary : AppColors.surface,
          borderRadius: AppRadii.roundedPill,
          border: Border.all(
            color: isSelected ? AppColors.primary : AppColors.border,
            width: 1,
          ),
          boxShadow: isSelected ? AppShadows.card : null,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (icon != null) ...[
              Icon(
                icon,
                size: 14,
                color: isSelected ? Colors.white : AppColors.textSecondary,
              ),
              const SizedBox(width: 5),
            ],
            Text(
              label,
              style: TextStyle(
                fontSize: 12,
                fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                color: isSelected ? Colors.white : AppColors.textPrimary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

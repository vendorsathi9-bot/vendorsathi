function readHttpUrl(value: string | undefined): string {
  const raw = (value ?? '').trim();
  if (!raw) return '';

  try {
    const url = new URL(raw);
    return url.protocol === 'http:' || url.protocol === 'https:'
      ? url.toString()
      : '';
  } catch {
    return '';
  }
}

function readEmail(value: string | undefined): string {
  const raw = (value ?? '').trim();
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(raw) ? raw : '';
}

function readText(value: string | undefined, fallback: string): string {
  const raw = (value ?? '').trim();
  return raw || fallback;
}

const siteOrigin = readHttpUrl(import.meta.env.VITE_SITE_URL).replace(/\/$/, '');

// Actual deployed VendorSaathi admin dashboard. Kept as the default so the
// admin entry point still works when the optional env var is not configured.
const deployedAdminDashboardUrl =
  'https://vendorsathi-o9w010418-vendor8.vercel.app/';

// Verified from mobile/build/app/outputs/flutter-apk/app-release.apk
// (package com.example.vendorsaathi_mobile, versionName 1.0.0, versionCode 1).
// Override after a new production build with VITE_APP_VERSION / VITE_APK_SIZE.
const verifiedAppVersion = '1.0.0';
const verifiedApkSize = '85.9 MB';

export const siteConfig = {
  /**
   * Public Android APK download URL. Empty until a real release asset exists.
   * VITE_ANDROID_APK_URL is the preferred name; VITE_APP_DOWNLOAD_URL remains
   * supported for backwards compatibility.
   */
  appDownloadUrl: readHttpUrl(
    import.meta.env.VITE_ANDROID_APK_URL ||
      import.meta.env.VITE_APP_DOWNLOAD_URL,
  ),
  /** Public admin dashboard URL. Defaults to the deployed Vercel dashboard. */
  adminDashboardUrl:
    readHttpUrl(import.meta.env.VITE_ADMIN_DASHBOARD_URL) ||
    deployedAdminDashboardUrl,
  /** Public application version shown next to the Android download. */
  appVersion: readText(import.meta.env.VITE_APP_VERSION, verifiedAppVersion),
  /** Public APK size shown next to the Android download. */
  apkSize: readText(import.meta.env.VITE_APK_SIZE, verifiedApkSize),
  /** Optional public contact address. */
  contactEmail: readEmail(import.meta.env.VITE_CONTACT_EMAIL),
  /** Optional canonical origin, used only to build absolute social URLs. */
  siteOrigin,
};

export const hasAppDownload = siteConfig.appDownloadUrl.length > 0;
export const hasAdminDashboard = siteConfig.adminDashboardUrl.length > 0;
export const hasContactEmail = siteConfig.contactEmail.length > 0;

export function contactHref(subject: string): string {
  return `mailto:${siteConfig.contactEmail}?subject=${encodeURIComponent(subject)}`;
}

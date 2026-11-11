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

const siteOrigin = readHttpUrl(import.meta.env.VITE_SITE_URL).replace(/\/$/, '');

export const siteConfig = {
  /** Public Android download URL. Empty until a real release URL is configured. */
  appDownloadUrl: readHttpUrl(import.meta.env.VITE_APP_DOWNLOAD_URL),
  /** Public admin dashboard URL. Empty until a deployment URL is configured. */
  adminDashboardUrl: readHttpUrl(import.meta.env.VITE_ADMIN_DASHBOARD_URL),
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

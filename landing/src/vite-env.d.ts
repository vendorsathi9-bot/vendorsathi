/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_APP_DOWNLOAD_URL?: string;
  readonly VITE_ADMIN_DASHBOARD_URL?: string;
  readonly VITE_CONTACT_EMAIL?: string;
  readonly VITE_SITE_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

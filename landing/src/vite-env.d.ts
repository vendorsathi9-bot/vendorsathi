/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_ANDROID_APK_URL?: string;
  readonly VITE_APP_DOWNLOAD_URL?: string;
  readonly VITE_APP_VERSION?: string;
  readonly VITE_APK_SIZE?: string;
  readonly VITE_ADMIN_DASHBOARD_URL?: string;
  readonly VITE_CONTACT_EMAIL?: string;
  readonly VITE_SITE_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

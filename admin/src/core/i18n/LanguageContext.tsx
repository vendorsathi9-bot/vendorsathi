import React, { createContext, useContext, useState } from 'react';

export type Language = 'en' | 'kn';

export interface Translations {
  [key: string]: string;
}

const translations: Record<Language, Translations> = {
  en: {
    // Brand & Nav
    brandTitle: 'VendorSaathi',
    brandSub: 'Administration',
    navDashboard: 'Dashboard',
    navVendors: 'Vendors',
    navCustomers: 'Customers & Users',
    navOrders: 'Orders',
    navDelivery: 'Delivery Agents',
    navReviews: 'Ratings & Reviews',
    navAudit: 'Audit Trail',
    logout: 'Log Out',
    langToggle: 'ಕನ್ನಡ',

    // Dashboard Overview
    overviewTitle: 'Platform Overview',
    overviewSubtitle: 'Real-time metrics, order volume, and financial aggregates across VendorSaathi.',
    activeVendors: 'Active Vendors',
    totalRegisteredVendors: 'Total registered',
    totalCustomers: 'Total Customers',
    allPlatformUsers: 'All platform users',
    completedOrders: 'Completed Orders',
    totalOrdersRecorded: 'Total orders',
    grossSalesValue: 'Gross Sales Value',
    collectionsLabel: 'Collections',
    totalProducts: 'Active Products',
    catalogCatalogued: 'Across all vendors',
    totalCategories: 'Active Categories',
    systemHealth: 'System Operational Health',
    coreDb: 'Core PostgreSQL DB',
    apiServices: 'REST API Services',
    gpsMobility: 'GPS & Discovery Engine',
    online: 'ONLINE',
    healthy: 'HEALTHY',
    active: 'ACTIVE',

    // Vendors
    vendorTitle: 'Vendor Management',
    vendorSubtitle: 'Review onboarding, active vendors, product catalog volume, and operational status.',
    approveVendor: 'Approve Vendor',
    suspendVendor: 'Suspend',
    activateVendor: 'Activate',
    pendingVendor: 'Set Pending',
    filterAll: 'All',
    filterPending: 'Pending Approval',
    filterActive: 'Active',
    filterSuspended: 'Suspended',
    vendorId: 'Vendor ID',
    businessName: 'Business Name',
    phone: 'Phone',
    location: 'Primary Location',
    productsCount: 'Products',
    status: 'Status',
    actions: 'Actions',
    noVendorsFound: 'No vendors match this filter.',
    statusUpdated: 'Vendor status successfully updated.',

    // Customers & Users
    userTitle: 'User & Role Management',
    userSubtitle: 'Manage platform accounts, role delegations (Vendor / Customer / Admin), and access.',
    assignRole: 'Assign Role',
    revokeRole: 'Revoke Role',
    userId: 'User ID',
    roles: 'Assigned Roles',
    joinedDate: 'Joined Date',
    selectRole: 'Select Role',
    roleUpdated: 'User role updated successfully.',

    // Orders
    orderTitle: 'Order Monitoring',
    orderSubtitle: 'Track live orders, fulfilment modes, line items, and financial settlements.',
    orderId: 'Order ID',
    fulfilment: 'Fulfilment',
    items: 'Items',
    amount: 'Total Amount',
    date: 'Date & Time',
    viewDetails: 'View Details',
    orderDetails: 'Order Details',
    close: 'Close',
    subtotal: 'Subtotal',
    discount: 'Loyalty Discount',
    total: 'Final Total',
    statusCompleted: 'COMPLETED',
    statusPending: 'PENDING',
    statusCancelled: 'CANCELLED',

    // Reviews & Moderation
    reviewTitle: 'Ratings & Reviews Moderation',
    reviewSubtitle: 'Moderate customer star ratings, comments, and community feedback.',
    rating: 'Rating',
    comment: 'Comment',
    customer: 'Customer',
    moderationStatus: 'Visibility',
    hideReview: 'Hide Review',
    unhideReview: 'Make Visible',
    visible: 'Visible',
    hidden: 'Hidden (Moderated)',
    noReviewsFound: 'No customer reviews submitted yet.',

    // Common
    refresh: 'Refresh',
    loading: 'Loading data...',
    saving: 'Saving...',
    error: 'An error occurred. Please try again.',
  },
  kn: {
    // Brand & Nav
    brandTitle: 'ವೆಂಡರ್‌ಸಾಥಿ',
    brandSub: 'ಆಡಳಿತ ಮಂಡಳಿ',
    navDashboard: 'ಡ್ಯಾಶ್‌ಬೋರ್ಡ್',
    navVendors: 'ಮಾರಾಟಗಾರರು',
    navCustomers: 'ಗ್ರಾಹಕರು & ಬಳಕೆದಾರರು',
    navOrders: 'ಆರ್ಡರ್‌ಗಳು',
    navDelivery: 'ವಿತರಣಾ ಸಿಬ್ಬಂದಿ',
    navReviews: 'ರೇಟಿಂಗ್‌ಗಳು & ವಿಮರ್ಶೆಗಳು',
    navAudit: 'ಆಡಿಟ್ ಲಾಗ್‌ಗಳು',
    logout: 'ನಿರ್ಗಮಿಸಿ',
    langToggle: 'English',

    // Dashboard Overview
    overviewTitle: 'ಪ್ಲಾಟ್‌ಫಾರ್ಮ್ ವಿವರಣೆ',
    overviewSubtitle: 'ವೆಂಡರ್‌ಸಾಥಿಯ ನೈಜ-ಸಮಯದ ಅಂಕಿಅಂಶಗಳು, ಆರ್ಡರ್‌ಗಳ ಪ್ರಮಾಣ ಮತ್ತು ಆರ್ಥಿಕ ವರದಿ.',
    activeVendors: 'ಸಕ್ರಿಯ ಮಾರಾಟಗಾರರು',
    totalRegisteredVendors: 'ಒಟ್ಟು ನೋಂದಾಯಿತ',
    totalCustomers: 'ಒಟ್ಟು ಗ್ರಾಹಕರು',
    allPlatformUsers: 'ಪ್ಲಾಟ್‌ಫಾರ್ಮ್‌ನ ಎಲ್ಲಾ ಬಳಕೆದಾರರು',
    completedOrders: 'ಪೂರ್ಣಗೊಂಡ ಆರ್ಡರ್‌ಗಳು',
    totalOrdersRecorded: 'ಒಟ್ಟು ಆರ್ಡರ್‌ಗಳು',
    grossSalesValue: 'ಒಟ್ಟು ಮಾರಾಟ ಮೌಲ್ಯ',
    collectionsLabel: 'ವಸೂಲಾತಿ',
    totalProducts: 'ಸಕ್ರಿಯ ಉತ್ಪನ್ನಗಳು',
    catalogCatalogued: 'ಎಲ್ಲಾ ಮಾರಾಟಗಾರರಿಂದ',
    totalCategories: 'ವರ್ಗಗಳು',
    systemHealth: 'ವ್ಯವಸ್ಥೆಯ ಕಾರ್ಯಾಚರಣೆ ಸ್ಥಿತಿ',
    coreDb: 'ಕೋರ್ ಪೋಸ್ಟ್‌ಗ್ರೆಸ್ ಡೇಟಾಬೇಸ್',
    apiServices: 'REST API ಸೇವೆಗಳು',
    gpsMobility: 'GPS & ಹುಡುಕಾಟ ಎಂಜಿನ್',
    online: 'ಆನ್‌ಲೈನ್',
    healthy: 'ಉತ್ತಮ ಸ್ಥಿತಿ',
    active: 'ಸಕ್ರಿಯ',

    // Vendors
    vendorTitle: 'ಮಾರಾಟಗಾರರ ನಿರ್ವಹಣೆ',
    vendorSubtitle: 'ಹೊಸ ಮಾರಾಟಗಾರರ ಅನುಮೋದನೆ, ಸಕ್ರಿಯ ಮಾರಾಟಗಾರರು ಮತ್ತು ಕಾರ್ಯಾಚರಣೆಯ ನಿಯಂತ್ರಣ.',
    approveVendor: 'ಅನುಮೋದಿಸಿ (Approve)',
    suspendVendor: 'ಅಮಾನತುಗೊಳಿಸಿ',
    activateVendor: 'ಸಕ್ರಿಯಗೊಳಿಸಿ',
    pendingVendor: 'ಬಾಕಿ ಇರಿಸಿ',
    filterAll: 'ಎಲ್ಲವೂ',
    filterPending: 'ಅನುಮೋದನೆ ಬಾಕಿ',
    filterActive: 'ಸಕ್ರಿಯ',
    filterSuspended: 'ಅಮಾನತಿನಲ್ಲಿರುವವರು',
    vendorId: 'ಮಾರಾಟಗಾರರ ID',
    businessName: 'ವ್ಯಾಪಾರದ ಹೆಸರು',
    phone: 'ದೂರವಾಣಿ ಸಂಖ್ಯೆ',
    location: 'ಮುಖ್ಯ ಸ್ಥಳ',
    productsCount: 'ಉತ್ಪನ್ನಗಳು',
    status: 'ಸ್ಥಿತಿ',
    actions: 'ಕ್ರಮಗಳು',
    noVendorsFound: 'ಯಾವುದೇ ಮಾರಾಟಗಾರರು ಕಂಡುಬಂದಿಲ್ಲ.',
    statusUpdated: 'ಮಾರಾಟಗಾರರ ಸ್ಥಿತಿಯನ್ನು ಯಶಸ್ವಿಯಾಗಿ ನವೀಕರಿಸಲಾಗಿದೆ.',

    // Customers & Users
    userTitle: 'ಬಳಕೆದಾರ ಮತ್ತು ಪಾತ್ರ ನಿರ್ವಹಣೆ',
    userSubtitle: 'ಖಾತೆಗಳು, ಪಾತ್ರಗಳ ಹಂಚಿಕೆ (ಮಾರಾಟಗಾರ / ಗ್ರಾಹಕ / ಅಡ್ಮಿನ್) ಮತ್ತು ಪ್ರವೇಶ ಸ್ಥಿತಿ.',
    assignRole: 'ಪಾತ್ರ ನೀಡಿ',
    revokeRole: 'ಪಾತ್ರ ಹಿಂಪಡೆಯಿರಿ',
    userId: 'ಬಳಕೆದಾರ ID',
    roles: 'ಹೊಂದಿರುವ ಪಾತ್ರಗಳು',
    joinedDate: 'ಸೇರ್ಪಡೆ ದಿನಾಂಕ',
    selectRole: 'ಪಾತ್ರ ಆಯ್ಕೆಮಾಡಿ',
    roleUpdated: 'ಬಳಕೆದಾರರ ಪಾತ್ರವನ್ನು ನವೀಕರಿಸಲಾಗಿದೆ.',

    // Orders
    orderTitle: 'ಆರ್ಡರ್ ಮೇಲ್ವಿಚಾರಣೆ',
    orderSubtitle: 'ನೈಜ ಆರ್ಡರ್‌ಗಳು, ವಿತರಣಾ ವಿಧಾನ, ವಸ್ತುಗಳು ಮತ್ತು ಹಣ ಪಾವತಿಗಳ ಮಾಹಿತಿ.',
    orderId: 'ಆರ್ಡರ್ ID',
    fulfilment: 'ವಿತರಣಾ ವಿಧಾನ',
    items: 'ವಸ್ತುಗಳು',
    amount: 'ಒಟ್ಟು ಮೊತ್ತ',
    date: 'ದಿನಾಂಕ & ಸಮಯ',
    viewDetails: 'ವಿವರಗಳನ್ನು ನೋಡಿ',
    orderDetails: 'ಆರ್ಡರ್ ವಿವರಗಳು',
    close: 'ಮುಚ್ಚಿ',
    subtotal: 'ಉಪಮೊತ್ತ',
    discount: 'ಲಾಯಲ್ಟಿ ರಿಯಾಯಿತಿ',
    total: 'ಅಂತಿಮ ಮೊತ್ತ',
    statusCompleted: 'ಪೂರ್ಣಗೊಂಡಿದೆ',
    statusPending: 'ಬಾಕಿ ಇದೆ',
    statusCancelled: 'ರದ್ದುಗೊಂಡಿದೆ',

    // Reviews & Moderation
    reviewTitle: 'ರೇಟಿಂಗ್ ಮತ್ತು ವಿಮರ್ಶೆಗಳ ಮಾಡರೇಶನ್',
    reviewSubtitle: 'ಗ್ರಾಹಕರ ಸ್ಟಾರ್ ರೇಟಿಂಗ್‌ಗಳು ಮತ್ತು ಅಭಿಪ್ರಾಯಗಳ ಪರಿಶೀಲನೆ.',
    rating: 'ರೇಟಿಂಗ್',
    comment: 'ಅಭಿಪ್ರಾಯ / ಕಾಮೆಂಟ್',
    customer: 'ಗ್ರಾಹಕ',
    moderationStatus: 'ಗೋಚರತೆ',
    hideReview: 'ಮರೆಮಾಡಿ',
    unhideReview: 'ಗೋಚರಿಸುವಂತೆ ಮಾಡಿ',
    visible: 'ಗೋಚರಿಸುತ್ತಿದೆ',
    hidden: 'ಮರೆಮಾಡಲಾಗಿದೆ (Admin)',
    noReviewsFound: 'ಯಾವುದೇ ಗ್ರಾಹಕರ ವಿಮರ್ಶೆಗಳು ಲಭ್ಯವಿಲ್ಲ.',

    // Common
    refresh: 'ನವೀಕರಿಸಿ',
    loading: 'ಲೋಡ್ ಮಾಡಲಾಗುತ್ತಿದೆ...',
    saving: 'ಉಳಿಸಲಾಗುತ್ತಿದೆ...',
    error: 'ದೋಷ ಉಂಟಾಗಿದೆ. ದಯವಿಟ್ಟು ಪುನಃ ಪ್ರಯತ್ನಿಸಿ.',
  },
};

interface LanguageContextType {
  language: Language;
  setLanguage: (lang: Language) => void;
  toggleLanguage: () => void;
  t: (key: string) => string;
}

const LanguageContext = createContext<LanguageContextType>({
  language: 'en',
  setLanguage: () => {},
  toggleLanguage: () => {},
  t: (key: string) => key,
});

export const LanguageProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [language, setLanguageState] = useState<Language>(() => {
    return (localStorage.getItem('vendorsaathi_admin_lang') as Language) || 'en';
  });

  const setLanguage = (lang: Language) => {
    setLanguageState(lang);
    localStorage.setItem('vendorsaathi_admin_lang', lang);
  };

  const toggleLanguage = () => {
    const next = language === 'en' ? 'kn' : 'en';
    setLanguage(next);
  };

  const t = (key: string): string => {
    return translations[language]?.[key] || translations['en']?.[key] || key;
  };

  return (
    <LanguageContext.Provider value={{ language, setLanguage, toggleLanguage, t }}>
      {children}
    </LanguageContext.Provider>
  );
};

export const useLanguage = () => useContext(LanguageContext);

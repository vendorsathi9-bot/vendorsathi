export type LegalDocumentId = 'privacy' | 'terms';

export interface LegalSection {
  heading: string;
  body: string;
}

export interface LegalDocument {
  id: LegalDocumentId;
  title: string;
  updated: string;
  intro: string;
  sections: LegalSection[];
}

export const LEGAL_DOCUMENTS: Record<LegalDocumentId, LegalDocument> = {
  privacy: {
    id: 'privacy',
    title: 'Privacy Policy',
    updated: '11 November 2026',
    intro:
      'This policy explains what information VendorSaathi handles when you use the website and mobile application, and the choices available to you.',
    sections: [
      {
        heading: 'Information we collect',
        body: 'Account details such as your name and phone number; order, cart, and delivery address information; saved addresses and loyalty activity; reviews you submit; payment method and payment status (we do not store full card details); device notification tokens; and device location when you grant permission.',
      },
      {
        heading: 'How we use information',
        body: 'To show nearby vendors, prepare and fulfil orders, coordinate delivery partners, record payments, provide order tracking, send service notifications, support customers and vendors, prevent misuse, and keep administrative audit records.',
      },
      {
        heading: 'Sharing with others',
        body: 'Vendors receive the order and contact details needed to prepare an order. Delivery partners receive pickup, drop-off, and contact details needed to complete a delivery. Payment providers process payment instructions. Infrastructure providers support hosting and notifications under contractual safeguards. We do not sell personal information.',
      },
      {
        heading: 'Location',
        body: 'Location access is requested only with your permission and is used to find nearby vendors, set delivery context, and share delivery progress. You can change location permission at any time in your device settings; some features may not work without it.',
      },
      {
        heading: 'Retention and security',
        body: 'Information is retained for as long as needed to operate the platform, meet legal obligations, resolve disputes, and maintain audit records. We use access controls and encrypted transport, but no system can be guaranteed to be completely secure.',
      },
      {
        heading: 'Your choices',
        body: 'You can update your profile and saved addresses, manage notification and location permissions, and ask for help with your account. For privacy requests, use the contact details published on this website.',
      },
    ],
  },
  terms: {
    id: 'terms',
    title: 'Terms & Conditions',
    updated: '11 November 2026',
    intro:
      'These terms govern your use of the VendorSaathi website and mobile application. By using the platform, you agree to them.',
    sections: [
      {
        heading: 'Using VendorSaathi',
        body: 'You must provide accurate account information and keep your login credentials secure. You are responsible for activity carried out through your account. The platform is intended for lawful personal or business use.',
      },
      {
        heading: 'A marketplace platform',
        body: 'VendorSaathi connects customers, local vendors, and delivery partners. Vendors are responsible for their listings, prices, stock, product quality, and applicable licences. Delivery partners are responsible for the deliveries they accept. VendorSaathi provides the technology that coordinates these interactions.',
      },
      {
        heading: 'Orders and payments',
        body: 'Orders are subject to vendor acceptance and availability. Prices and charges are shown in the app before you confirm an order. Payment may be made by cash, UPI, or an available online payment method. Cancellations, refunds, and payment disputes are handled according to the order details and applicable law.',
      },
      {
        heading: 'Delivery',
        body: 'Delivery estimates are indicative and can change with distance, weather, traffic, vendor preparation time, and network conditions. Live tracking depends on the delivery partner’s device, GPS availability, and connectivity.',
      },
      {
        heading: 'Reviews and content',
        body: 'Reviews should reflect genuine experiences. Abusive, misleading, unlawful, or infringing content is not permitted. VendorSaathi may moderate or remove content that breaks these terms.',
      },
      {
        heading: 'Acceptable use',
        body: 'Do not attempt fraud, impersonation, unauthorised access, scraping, disruption of the service, or any activity that harms other users, vendors, delivery partners, or the platform.',
      },
      {
        heading: 'Availability and liability',
        body: 'The platform is provided on an “as available” basis and may change or be interrupted. To the extent permitted by law, VendorSaathi is not liable for indirect losses arising from reliance on the platform, while nothing in these terms limits rights that cannot be limited by law.',
      },
      {
        heading: 'Changes and contact',
        body: 'We may update these terms as the platform evolves. Material changes will be reflected on this page. Questions about these terms can be sent using the contact details published on this website.',
      },
    ],
  },
};

import {
  BarChart3,
  Bike,
  BookOpen,
  ClipboardCheck,
  ClipboardList,
  Compass,
  CreditCard,
  LayoutDashboard,
  LocateFixed,
  MapPinned,
  Navigation,
  PackageCheck,
  ScanLine,
  Search,
  ShieldCheck,
  ShoppingBag,
  ShoppingBasket,
  Star,
  Store,
  Truck,
  Users,
  Wallet,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

export interface NavLink {
  label: string;
  href: string;
  sectionId: string;
}

export const NAV_LINKS: NavLink[] = [
  { label: 'Home', href: '#home', sectionId: 'home' },
  { label: 'How It Works', href: '#how-it-works', sectionId: 'how-it-works' },
  { label: 'Features', href: '#features', sectionId: 'features' },
  { label: 'For Vendors', href: '#vendors', sectionId: 'vendors' },
  { label: 'About', href: '#about', sectionId: 'about' },
  { label: 'Contact', href: '#contact', sectionId: 'contact' },
];

export const SECTION_IDS = [
  'home',
  'how-it-works',
  'features',
  'customers',
  'vendors',
  'delivery',
  'download',
  'admin',
  'about',
  'contact',
];

export interface ContentItem {
  title: string;
  description: string;
  icon: LucideIcon;
}

export const STEPS: ContentItem[] = [
  {
    title: 'Discover local stores',
    description:
      'Open the app to see nearby vendors on an interactive map and a distance-sorted list, with the products each shop carries.',
    icon: MapPinned,
  },
  {
    title: 'Order conveniently',
    description:
      'Add what you need to a single-vendor cart, apply loyalty rewards when available, and choose cash, UPI, or online payment.',
    icon: ShoppingBasket,
  },
  {
    title: 'Get it delivered',
    description:
      'Follow your order from accepted to ready, then track your delivery partner’s shared location and route on the map.',
    icon: Bike,
  },
];

export const FEATURES: ContentItem[] = [
  {
    title: 'Nearby vendor discovery',
    description:
      'Find local stores around you with an interactive map and a list sorted by distance.',
    icon: MapPinned,
  },
  {
    title: 'Product browsing',
    description:
      'Explore each vendor’s catalogue with prices, availability, and Kannada product names.',
    icon: Store,
  },
  {
    title: 'Cart & checkout',
    description:
      'Build a single-vendor cart, apply rewards, and choose cash, UPI, or online payment.',
    icon: ShoppingBasket,
  },
  {
    title: 'Order management',
    description:
      'Follow every order through placed, accepted, preparing, ready, and picked up.',
    icon: ClipboardList,
  },
  {
    title: 'Live delivery tracking',
    description:
      'See your delivery partner’s shared GPS location and the road route on the map.',
    icon: Navigation,
  },
  {
    title: 'Digital payments',
    description:
      'Choose cash, UPI, or gateway payments, with payment status recorded against the order.',
    icon: CreditCard,
  },
  {
    title: 'Customer reviews',
    description:
      'Rate completed orders and share honest feedback to help neighbours shop well.',
    icon: Star,
  },
  {
    title: 'Vendor management',
    description:
      'Run your storefront, catalogue, inventory, order queue, and udhaar ledger in one place.',
    icon: LayoutDashboard,
  },
];

export const CUSTOMER_HIGHLIGHTS: ContentItem[] = [
  {
    title: 'Discover nearby shops',
    description:
      'A live map and a nearby list make it easy to see who is open around you.',
    icon: Compass,
  },
  {
    title: 'Browse real catalogues',
    description:
      'Products, prices, and availability from the vendor — in English or Kannada.',
    icon: Search,
  },
  {
    title: 'Place orders',
    description:
      'A focused single-vendor cart with rewards and flexible payment options.',
    icon: ShoppingBag,
  },
  {
    title: 'Track deliveries',
    description:
      'Watch order milestones and your delivery partner’s location in real time.',
    icon: LocateFixed,
  },
  {
    title: 'Manage your orders',
    description:
      'Revisit past orders, saved addresses, loyalty points, and reviews.',
    icon: ClipboardCheck,
  },
];

export const VENDOR_BENEFITS: ContentItem[] = [
  {
    title: 'A storefront that stays open',
    description:
      'Publish your shop details and availability so nearby customers can find you.',
    icon: Store,
  },
  {
    title: 'Catalogue & inventory',
    description:
      'Add products, update prices and stock, and keep your catalogue current.',
    icon: BookOpen,
  },
  {
    title: 'Orders in one queue',
    description:
      'Accept, prepare, and update orders with a clear, shared status timeline.',
    icon: ClipboardList,
  },
  {
    title: 'Sales insights',
    description:
      'Review revenue, completed orders, and seven-day sales trends in the vendor portal.',
    icon: BarChart3,
  },
  {
    title: 'Digital customer reach',
    description:
      'Be discoverable to customers who are already shopping in your neighbourhood.',
    icon: Users,
  },
  {
    title: 'Udhaar, digitised',
    description:
      'Keep a customer-wise credit ledger and record repayments without a paper book.',
    icon: Wallet,
  },
];

export const DELIVERY_POINTS: ContentItem[] = [
  {
    title: 'Assigned deliveries',
    description:
      'See active delivery assignments with clear pickup and drop-off details.',
    icon: PackageCheck,
  },
  {
    title: 'Pickup to drop-off flow',
    description:
      'Move each delivery from pickup through to delivered with simple status updates.',
    icon: Truck,
  },
  {
    title: 'Customer coordination',
    description:
      'Call the customer directly from the delivery screen when you need to.',
    icon: Bike,
  },
  {
    title: 'Live location sharing',
    description:
      'Share your GPS location so customers can follow the delivery on the map.',
    icon: Navigation,
  },
];

export const ADMIN_CAPABILITIES: ContentItem[] = [
  {
    title: 'Vendors & customers',
    description: 'Oversee onboarding and account records.',
    icon: Users,
  },
  {
    title: 'Orders & delivery',
    description: 'Monitor operations and delivery agents.',
    icon: ClipboardList,
  },
  {
    title: 'Review moderation',
    description: 'Keep customer feedback helpful and fair.',
    icon: ShieldCheck,
  },
  {
    title: 'Audit logs',
    description: 'Trace administrative actions with a clear record.',
    icon: ScanLine,
  },
];

export const VALUES: ContentItem[] = [
  {
    title: 'Local first',
    description:
      'The shop around the corner should be as easy to reach as the big marketplace.',
    icon: MapPinned,
  },
  {
    title: 'Clear by design',
    description:
      'Simple flows for customers, vendors, and delivery partners — in two languages.',
    icon: Compass,
  },
  {
    title: 'Trust in every order',
    description:
      'Transparent order states, tracked deliveries, and honest customer feedback.',
    icon: ShieldCheck,
  },
];

export const SITE_DESCRIPTION =
  'VendorSaathi is a local commerce platform that connects neighbourhood vendors with nearby customers, with ordering, payments, delivery tracking, and a vendor portal.';

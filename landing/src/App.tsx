import { useCallback, useEffect, useState } from 'react';

import { Footer } from './components/layout/Footer';
import { Navbar } from './components/layout/Navbar';
import { About } from './components/sections/About';
import { AdminDashboard } from './components/sections/AdminDashboard';
import { Contact } from './components/sections/Contact';
import { CustomerExperience } from './components/sections/CustomerExperience';
import { DeliveryPartners } from './components/sections/DeliveryPartners';
import { DownloadApp } from './components/sections/DownloadApp';
import { Features } from './components/sections/Features';
import { Hero } from './components/sections/Hero';
import { HowItWorks } from './components/sections/HowItWorks';
import { Vendors } from './components/sections/Vendors';
import { LegalDialog } from './components/ui/LegalDialog';
import type { LegalDocumentId } from './content/legal';
import { siteConfig } from './lib/config';

function useSocialMetadata() {
  useEffect(() => {
    const origin =
      siteConfig.siteOrigin ||
      (typeof window !== 'undefined' ? window.location.origin : '');

    if (!origin) return;

    const pageUrl = `${origin}${window.location.pathname}`;
    const imageUrl = `${origin}/og-image.png`;

    const setMeta = (
      attribute: 'property' | 'name',
      key: string,
      content: string,
    ) => {
      let element = document.head.querySelector<HTMLMetaElement>(
        `meta[${attribute}="${key}"]`,
      );

      if (!element) {
        element = document.createElement('meta');
        element.setAttribute(attribute, key);
        document.head.appendChild(element);
      }

      element.setAttribute('content', content);
    };

    setMeta('property', 'og:url', pageUrl);
    setMeta('property', 'og:image', imageUrl);
    setMeta('name', 'twitter:image', imageUrl);

    let canonical = document.head.querySelector<HTMLLinkElement>(
      'link[rel="canonical"]',
    );

    if (!canonical) {
      canonical = document.createElement('link');
      canonical.setAttribute('rel', 'canonical');
      document.head.appendChild(canonical);
    }

    canonical.setAttribute('href', pageUrl);
  }, []);
}

export default function App() {
  const [legalDocument, setLegalDocument] = useState<LegalDocumentId | null>(
    null,
  );

  useSocialMetadata();

  const openLegal = useCallback((id: LegalDocumentId) => {
    setLegalDocument(id);
  }, []);

  const closeLegal = useCallback(() => {
    setLegalDocument(null);
  }, []);

  return (
    <div className="flex min-h-screen flex-col bg-brand-cream">
      <a
        href="#main-content"
        className="sr-only z-[120] rounded-full bg-brand-espresso px-5 py-3 text-sm font-semibold text-brand-cream focus:not-sr-only focus:fixed focus:left-4 focus:top-4"
      >
        Skip to main content
      </a>

      <Navbar />

      <main id="main-content" tabIndex={-1} className="flex-1 focus:outline-none">
        <Hero />
        <HowItWorks />
        <Features />
        <CustomerExperience />
        <Vendors />
        <DeliveryPartners />
        <DownloadApp />
        <AdminDashboard />
        <About />
        <Contact />
      </main>

      <Footer onOpenLegal={openLegal} />
      <LegalDialog documentId={legalDocument} onClose={closeLegal} />
    </div>
  );
}

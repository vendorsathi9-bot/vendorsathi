import { AnimatePresence, motion, useReducedMotion } from 'framer-motion';
import { X } from 'lucide-react';
import { useEffect, useRef } from 'react';

import {
  LEGAL_DOCUMENTS,
  type LegalDocumentId,
} from '../../content/legal';
import { hasContactEmail, siteConfig } from '../../lib/config';

interface LegalDialogProps {
  documentId: LegalDocumentId | null;
  onClose: () => void;
}

export function LegalDialog({ documentId, onClose }: LegalDialogProps) {
  const panelRef = useRef<HTMLDivElement>(null);
  const closeRef = useRef<HTMLButtonElement>(null);
  const reduceMotion = useReducedMotion();
  const document = documentId ? LEGAL_DOCUMENTS[documentId] : null;

  useEffect(() => {
    if (!documentId) return;

    const previouslyFocused = window.document.activeElement as HTMLElement | null;
    const previousOverflow = window.document.body.style.overflow;
    window.document.body.style.overflow = 'hidden';

    const focusTimer = window.setTimeout(() => closeRef.current?.focus(), 40);

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        onClose();
        return;
      }

      if (event.key !== 'Tab' || !panelRef.current) return;

      const focusable = panelRef.current.querySelectorAll<HTMLElement>(
        'a[href], button:not([disabled]), textarea, input, select, [tabindex]:not([tabindex="-1"])',
      );

      if (focusable.length === 0) return;

      const first = focusable[0];
      const last = focusable[focusable.length - 1];

      if (event.shiftKey && window.document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && window.document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    };

    window.document.addEventListener('keydown', handleKeyDown);

    return () => {
      window.clearTimeout(focusTimer);
      window.document.body.style.overflow = previousOverflow;
      window.document.removeEventListener('keydown', handleKeyDown);
      previouslyFocused?.focus?.();
    };
  }, [documentId, onClose]);

  return (
    <AnimatePresence>
      {document ? (
        <motion.div
          className="fixed inset-0 z-[100] flex items-end justify-center bg-brand-espresso/55 p-0 backdrop-blur-sm sm:items-center sm:p-6"
          initial={reduceMotion ? false : { opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: reduceMotion ? 0 : 0.2 }}
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) onClose();
          }}
        >
          <motion.div
            ref={panelRef}
            role="dialog"
            aria-modal="true"
            aria-labelledby="legal-dialog-title"
            className="relative flex max-h-[92dvh] w-full max-w-2xl flex-col overflow-hidden rounded-t-4xl border border-brand-border bg-brand-white shadow-lift sm:max-h-[85vh] sm:rounded-4xl"
            initial={reduceMotion ? false : { opacity: 0, y: 24, scale: 0.985 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={reduceMotion ? undefined : { opacity: 0, y: 16, scale: 0.99 }}
            transition={{ duration: reduceMotion ? 0 : 0.28, ease: [0.22, 1, 0.36, 1] }}
          >
            <div className="flex items-start justify-between gap-4 border-b border-brand-border bg-brand-cream/70 px-6 py-5 sm:px-8">
              <div>
                <p className="text-[0.66rem] font-semibold uppercase tracking-[0.18em] text-brand-brown">
                  VendorSaathi
                </p>
                <h2
                  id="legal-dialog-title"
                  className="mt-1 text-xl font-semibold tracking-[-0.02em] text-brand-espresso sm:text-2xl"
                >
                  {document.title}
                </h2>
                <p className="mt-1 text-xs text-brand-brown">
                  Last updated {document.updated}
                </p>
              </div>
              <button
                ref={closeRef}
                type="button"
                onClick={onClose}
                aria-label={`Close ${document.title}`}
                className="grid h-10 w-10 shrink-0 place-items-center rounded-full border border-brand-border bg-brand-white text-brand-espresso transition hover:bg-brand-soft"
              >
                <X className="h-5 w-5" aria-hidden="true" />
              </button>
            </div>

            <div className="overflow-y-auto px-6 py-6 sm:px-8">
              <p className="text-sm leading-relaxed text-brand-brown">
                {document.intro}
              </p>

              <div className="mt-6 space-y-6">
                {document.sections.map((section) => (
                  <section key={section.heading}>
                    <h3 className="text-sm font-semibold text-brand-espresso">
                      {section.heading}
                    </h3>
                    <p className="mt-1.5 text-sm leading-relaxed text-brand-brown">
                      {section.body}
                    </p>
                  </section>
                ))}
              </div>

              <p className="mt-8 rounded-2xl border border-brand-border bg-brand-cream px-4 py-3 text-xs leading-relaxed text-brand-brown">
                {hasContactEmail
                  ? `Questions? Write to ${siteConfig.contactEmail}.`
                  : 'Contact details for privacy and legal questions will be published on this website.'}
              </p>
            </div>
          </motion.div>
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
}

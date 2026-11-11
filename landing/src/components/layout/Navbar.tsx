import {
  AnimatePresence,
  motion,
  useScroll,
  useSpring,
} from 'framer-motion';
import { Download, Menu, X } from 'lucide-react';
import { useEffect, useState } from 'react';

import { NAV_LINKS, SECTION_IDS } from '../../content/site';
import { cn } from '../../lib/cn';
import { Button } from '../ui/Button';
import { Logo } from '../ui/Logo';

function useActiveSection(ids: readonly string[]): string {
  const [active, setActive] = useState('home');

  useEffect(() => {
    if (typeof IntersectionObserver === 'undefined') return;

    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((entry) => entry.isIntersecting)
          .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];

        if (visible?.target.id) {
          setActive(visible.target.id);
        }
      },
      { rootMargin: '-38% 0px -52% 0px', threshold: [0, 0.2, 0.5] },
    );

    ids.forEach((id) => {
      const element = document.getElementById(id);
      if (element) observer.observe(element);
    });

    return () => observer.disconnect();
  }, [ids]);

  return active;
}

export function Navbar() {
  const [scrolled, setScrolled] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const activeSection = useActiveSection(SECTION_IDS);
  const { scrollYProgress } = useScroll();
  const progress = useSpring(scrollYProgress, {
    stiffness: 130,
    damping: 28,
    mass: 0.2,
  });

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 10);
    onScroll();
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  useEffect(() => {
    if (!menuOpen) return;

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setMenuOpen(false);
    };
    const onResize = () => {
      if (window.innerWidth >= 1024) setMenuOpen(false);
    };

    document.addEventListener('keydown', onKeyDown);
    window.addEventListener('resize', onResize);

    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener('keydown', onKeyDown);
      window.removeEventListener('resize', onResize);
    };
  }, [menuOpen]);

  return (
    <header
      className={cn(
        'sticky top-0 z-50 transition-all duration-300',
        scrolled || menuOpen
          ? 'bg-brand-cream/95 shadow-nav backdrop-blur-md'
          : 'bg-brand-cream/70 backdrop-blur-sm',
      )}
    >
      <div className="container-page flex h-[72px] items-center justify-between gap-4">
        <a
          href="#home"
          aria-label="VendorSaathi — back to top"
          className="rounded-2xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-espresso focus-visible:ring-offset-4 focus-visible:ring-offset-brand-cream"
          onClick={() => setMenuOpen(false)}
        >
          <Logo />
        </a>

        <nav aria-label="Primary" className="hidden items-center gap-1 lg:flex">
          {NAV_LINKS.map((link) => {
            const isActive = activeSection === link.sectionId;
            return (
              <a
                key={link.href}
                href={link.href}
                aria-current={isActive ? 'location' : undefined}
                className={cn(
                  'rounded-full px-3.5 py-2 text-[0.86rem] font-medium transition-colors duration-200',
                  isActive
                    ? 'bg-brand-soft text-brand-espresso'
                    : 'text-brand-brown hover:bg-brand-espresso/5 hover:text-brand-espresso',
                )}
              >
                {link.label}
              </a>
            );
          })}
        </nav>

        <div className="flex items-center gap-2">
          <Button
            href="#download"
            size="sm"
            icon={<Download className="h-4 w-4" />}
            className="hidden sm:inline-flex"
          >
            Download App
          </Button>
          <button
            type="button"
            className="grid h-11 w-11 place-items-center rounded-full border border-brand-border bg-brand-white text-brand-espresso transition hover:bg-brand-soft lg:hidden"
            aria-expanded={menuOpen}
            aria-controls="mobile-navigation"
            aria-label={menuOpen ? 'Close navigation menu' : 'Open navigation menu'}
            onClick={() => setMenuOpen((open) => !open)}
          >
            {menuOpen ? (
              <X className="h-5 w-5" aria-hidden="true" />
            ) : (
              <Menu className="h-5 w-5" aria-hidden="true" />
            )}
          </button>
        </div>
      </div>

      <AnimatePresence>
        {menuOpen ? (
          <motion.nav
            id="mobile-navigation"
            aria-label="Mobile"
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.22, ease: 'easeOut' }}
            className="absolute inset-x-0 top-full max-h-[calc(100dvh-72px)] overflow-y-auto border-t border-brand-border bg-brand-cream px-5 pb-7 pt-3 shadow-nav lg:hidden"
          >
            <ul className="flex flex-col">
              {NAV_LINKS.map((link) => (
                <li key={link.href}>
                  <a
                    href={link.href}
                    onClick={() => setMenuOpen(false)}
                    aria-current={
                      activeSection === link.sectionId ? 'location' : undefined
                    }
                    className={cn(
                      'flex items-center justify-between rounded-2xl px-3 py-3.5 text-base font-medium transition-colors',
                      activeSection === link.sectionId
                        ? 'bg-brand-soft text-brand-espresso'
                        : 'text-brand-brown hover:bg-brand-espresso/5 hover:text-brand-espresso',
                    )}
                  >
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
            <Button
              href="#download"
              className="mt-4 w-full"
              icon={<Download className="h-4 w-4" />}
              onClick={() => setMenuOpen(false)}
            >
              Download App
            </Button>
          </motion.nav>
        ) : null}
      </AnimatePresence>

      <motion.div
        aria-hidden="true"
        style={{ scaleX: progress }}
        className="absolute inset-x-0 bottom-0 h-[2px] origin-left bg-brand-yellow/80"
      />
    </header>
  );
}

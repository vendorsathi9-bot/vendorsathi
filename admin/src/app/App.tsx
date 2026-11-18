import React, { useState, useEffect, useRef } from 'react';
import { useAuth } from '../core/auth/AuthContext';
import { useLanguage } from '../core/i18n/LanguageContext';
import { LoginPage } from '../features/auth/LoginPage';
import { DashboardPage } from '../features/dashboard/DashboardPage';
import { VendorsPage } from '../features/vendors/VendorsPage';
import { CustomersPage } from '../features/customers/CustomersPage';
import { OrdersPage } from '../features/orders/OrdersPage';
import { ReviewsModerationPage } from '../features/reviews/ReviewsModerationPage';
import { DeliveryAgentsPage } from '../features/delivery/DeliveryAgentsPage';
import { AuditLogsPage } from '../features/audit/AuditLogsPage';
import { 
  LayoutDashboard, 
  Store, 
  Users, 
  ShoppingBag, 
  Truck,
  MessageSquare,
  ShieldCheck, 
  LogOut,
  Globe,
  Menu,
  X,
} from 'lucide-react';

// Helper: debounced resize listener for sidebar state
const useResponsiveSidebar = () => {
  const [width, setWidth] = useState(window.innerWidth);
  useEffect(() => {
    const handleResize = () => setWidth(window.innerWidth);
    handleResize();
    window.addEventListener('resize', handleResize, { passive: true });
    return () => window.removeEventListener('resize', handleResize);
  }, []);
  return width;
};

export const App: React.FC = () => {
  const { user, loading, logout, isAdmin } = useAuth();
  const { t, language, toggleLanguage } = useLanguage();
  const [currentTab, setCurrentTab] = useState<'dashboard' | 'vendors' | 'customers' | 'orders' | 'delivery' | 'reviews' | 'audit'>('dashboard');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const sidebarRef = useRef<HTMLDivElement>(null);
  const screenWidth = useResponsiveSidebar();

  if (loading) {
    return (
      <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#806B55' }}>
        Verifying admin session...
      </div>
    );
  }

  if (!user || !isAdmin) {
    return <LoginPage />;
  }

  const navItems = [
    { id: 'dashboard', label: t('navDashboard'), icon: LayoutDashboard },
    { id: 'vendors', label: t('navVendors'), icon: Store },
    { id: 'customers', label: t('navCustomers'), icon: Users },
    { id: 'orders', label: t('navOrders'), icon: ShoppingBag },
    { id: 'delivery', label: t('navDelivery'), icon: Truck },
    { id: 'reviews', label: t('navReviews'), icon: MessageSquare },
    { id: 'audit', label: t('navAudit'), icon: ShieldCheck },
  ] as const;

  // Close sidebar on tab selection when expanded
  useEffect(() => {
    if (!sidebarCollapsed && sidebarRef.current) {
      const rect = sidebarRef.current.getBoundingClientRect();
      if (rect.width <= 80) {
        setSidebarCollapsed(true);
      }
    }
  }, [currentTab, sidebarCollapsed]);

  return (
    <div className="min-h-screen" style={{ display: 'flex', backgroundColor: '#FFFCEF', color: '#38200F' }}>
      {/* Sidebar */}
      <aside
        ref={sidebarRef}
        style={{
          width: sidebarCollapsed ? '64px' : '260px',
          backgroundColor: '#38200F',
          color: '#FFFCEF',
          display: 'flex',
          flexDirection: 'column',
          position: 'fixed',
          top: 0,
          bottom: 0,
          left: 0,
          zIndex: 50,
          transition: 'transform 0.3s ease',
          overflowY: 'auto',
          boxShadow: '2px 0 6px rgba(0,0,0,0.1)',
        }}
        onClick={(e) => {
          // Click on sidebar background doesn't toggle, only menu button does
          e.stopPropagation();
        }}
      >
        {/* Brand Header - collapsible */}
        <div style={{
          padding: sidebarCollapsed ? '0.75rem 1rem' : '1.25rem 1.5rem',
          borderBottom: '1px solid rgba(255,252,239,0.12)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          {sidebarCollapsed ? (
            <div
              style={{
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                padding: '0.5rem 0',
              }}
            >
              <div
                style={{
                  width: '32px', height: '32px',
                  padding: '0.5rem',
                  background: '#F4B400',
                  borderRadius: '50%',
                  margin: '0 auto',
                }}
              >
                <ShieldCheck size={16} color="#38200F" />
              </div>
            </div>
          ) : (
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              <div style={{ padding: '0.5rem', background: '#F4B400', borderRadius: '0.5rem' }}>
                <ShieldCheck size={20} color="#38200F" />
              </div>
              <div>
                <div style={{ fontWeight: 700, fontSize: '1.05rem', letterSpacing: '-0.025em', color: '#FFFCEF' }}>{t('brandTitle')}</div>
                <div style={{ fontSize: '0.7rem', color: '#C9B49E', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{t('brandSub')}</div>
              </div>
            </div>
          )}
          {/* Hamburger / Close button */}
          <button
            onClick={() => setSidebarCollapsed(prev => !prev)}
            style={{
              width: '32px', height: '32px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              borderRadius: '0.5rem',
              background: 'rgba(255,252,239,0.08)',
              color: '#F4B400',
              fontSize: '1rem',
              transition: 'background 0.15s',
              flexShrink: 0,
            }}
            aria-label={sidebarCollapsed ? 'Open sidebar' : 'Close sidebar'}
          >
            {sidebarCollapsed ? <Menu size={16} /> : <X size={16} />}
          </button>
        </div>

        {/* Language switch - always visible on hover/collapse */}
        <div style={{
          padding: '0.5rem 1rem',
          borderBottom: '1px solid rgba(255,252,239,0.12)',
          textAlign: 'center',
        }}>
          {sidebarCollapsed ? (
            <button
              onClick={toggleLanguage}
              style={{
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '0.5rem',
                padding: '0.45rem',
                backgroundColor: 'rgba(255,252,239,0.08)',
                color: '#F4B400',
                borderRadius: '0.375rem',
                fontSize: '0.75rem',
                fontWeight: 600,
                marginBottom: '0.25rem',
              }}
              aria-label={t('switchLanguage')}
            >
              <Globe size={12} /> {language === 'en' ? 'kn' : 'EN'}
            </button>
          ) : (
            <div style={{ padding: '0.75rem 1rem', borderBottom: '1px solid rgba(255,252,239,0.12)' }}>
              <button
                onClick={toggleLanguage}
                style={{
                  width: '100%',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '0.5rem',
                  padding: '0.45rem',
                  backgroundColor: 'rgba(255,252,239,0.08)',
                  color: '#F4B400',
                  borderRadius: '0.375rem',
                  fontSize: '0.8rem',
                  fontWeight: 600,
                  transition: 'background 0.15s',
                }}
              >
                <Globe size={14} /> {language === 'en' ? 'ಕನ್ನಡಕ್ಕೆ ಬದಲಾಯಿಸಿ' : 'Switch to English'}
              </button>
            </div>
          )}
        </div>

        {/* Navigation */}
        <nav style={{
          padding: sidebarCollapsed ? '0.5rem 0.75rem 1rem' : '1rem',
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          gap: sidebarCollapsed ? 0 : 0.35,
        }}>
          {navItems.map((item) => {
            const Icon = item.icon;
            const active = currentTab === item.id;
            const isLast = navItems.indexOf(item) === navItems.length - 1;
            return (
              <button
                key={item.id}
                onClick={() => {
                  setCurrentTab(item.id);
                  // Auto-collapse sidebar on mobile after navigation
                  if (window.innerWidth < 768) {
                    setSidebarCollapsed(true);
                  }
                }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: sidebarCollapsed ? 0.5 : 0.75,
                  padding: sidebarCollapsed ? '0.5rem 0.625rem' : '0.75rem 1rem',
                  borderRadius: sidebarCollapsed ? '0.375rem' : '0.5rem',
                  color: active ? '#F4B400' : '#D9C7B2',
                  backgroundColor: active ? 'rgba(244,180,0,0.16)' : 'transparent',
                  fontWeight: active ? 600 : 500,
                  fontSize: sidebarCollapsed ? '0.7rem' : '0.9rem',
                  width: '100%',
                  textAlign: 'left',
                  overflow: 'hidden',
                  whiteSpace: 'nowrap',
                  textOverflow: 'ellipsis',
                  transition: 'all 0.15s ease',
                  marginBottom: isLast ? '0.25rem' : 0,
                }}
                aria-label={item.label}
              >
                <Icon
                  size={sidebarCollapsed ? 14 : 18}
                  color={active ? '#F4B400' : '#C9B49E'}
                />
                {sidebarCollapsed ? null : item.label}
              </button>
            );
          })}
        </nav>

        {/* User Info & Logout */}
        <div style={{
          padding: sidebarCollapsed ? '0.75rem 1rem' : '1.25rem',
          borderTop: '1px solid rgba(255,252,239,0.12)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          {sidebarCollapsed ? (
            <div style={{ textAlign: 'center' }}>
              <div
                style={{
                  width: '32px', height: '32px',
                  padding: '0.5rem',
                  background: '#F4B400',
                  borderRadius: '50%',
                  margin: '0 auto 0.5rem',
                }}
              >
                <ShieldCheck size={14} color="#38200F" />
              </div>
              <div
                style={{
                  fontSize: '0.7rem', color: '#C9B49E',
                  whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis',
                }}
              >
                {user.phone.substring(0, 8)}...
              </div>
            </div>
          ) : (
            <div style={{ overflow: 'hidden' }}>
              <div style={{ fontSize: '0.875rem', fontWeight: 600, color: '#FFFCEF', whiteSpace: 'nowrap', textOverflow: 'ellipsis', overflow: 'hidden' }}>
                {user.phone}
              </div>
              <div style={{ fontSize: '0.75rem', color: '#B9A48F' }}>
                {user.roles.join(', ')}
              </div>
            </div>
          )}
          <button
            onClick={logout}
            title={t('logout')}
            style={{
              padding: '0.5rem',
              borderRadius: '0.375rem',
              color: '#D9C7B2',
              background: 'rgba(255,252,239,0.08)',
              flexShrink: 0,
            }}
          >
            <LogOut size={16} />
          </button>
        </div>
      </aside>

      {/* Main Content Area */}
      <main style={{
        marginLeft: sidebarCollapsed ? '64px' : '260px',
        flex: 1,
        padding: '2rem 2.5rem 2rem',
        maxWidth: '1400px',
        width: '100%',
        transition: 'margin-left 0.3s ease',
      }}>
        {/* Mobile sidebar toggle bar - visible only on small screens */}
        {screenWidth < 768 && (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'flex-start',
              gap: '0.75rem',
              padding: '0.5rem 1rem',
              background: '#38200F',
              borderBottom: '1px solid rgba(255,252,239,0.12)',
            }}
          >
            <button
              onClick={() => setSidebarCollapsed(prev => !prev)}
              style={{
                width: '32px', height: '32px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                borderRadius: '0.375rem',
                color: '#F4B400',
                fontSize: '1.25rem',
              }}
              aria-label={sidebarCollapsed ? 'Open sidebar' : 'Close sidebar'}
            >
              {sidebarCollapsed ? <Menu size={16} /> : <X size={16} />}
            </button>
            <span style={{ fontWeight: 600, fontSize: '0.9rem' }}>{t('menu')}</span>
          </div>
        )}

        {currentTab === 'dashboard' && <DashboardPage />}
        {currentTab === 'vendors' && <VendorsPage />}
        {currentTab === 'customers' && <CustomersPage />}
        {currentTab === 'orders' && <OrdersPage />}
        {currentTab === 'delivery' && <DeliveryAgentsPage />}
        {currentTab === 'reviews' && <ReviewsModerationPage />}
        {currentTab === 'audit' && <AuditLogsPage />}
      </main>
    </div>
  );
};

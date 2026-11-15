import React, { useState } from 'react';
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
  Globe
} from 'lucide-react';

export const App: React.FC = () => {
  const { user, loading, logout, isAdmin } = useAuth();
  const { t, language, toggleLanguage } = useLanguage();
  const [currentTab, setCurrentTab] = useState<'dashboard' | 'vendors' | 'customers' | 'orders' | 'delivery' | 'reviews' | 'audit'>('dashboard');

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

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      {/* Sidebar */}
      <aside style={{
        width: '260px',
        backgroundColor: '#38200F',
        color: '#FFFCEF',
        display: 'flex',
        flexDirection: 'column',
        position: 'fixed',
        top: 0,
        bottom: 0,
        left: 0,
        zIndex: 40,
        transition: 'transform 0.2s',
      }}>
        {/* Brand */}
        <div style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid rgba(255,252,239,0.12)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <div style={{ padding: '0.5rem', background: '#F4B400', borderRadius: '0.5rem' }}>
              <ShieldCheck size={22} color="#38200F" />
            </div>
            <div>
              <div style={{ fontWeight: 700, fontSize: '1.05rem', letterSpacing: '-0.025em', color: '#FFFCEF' }}>{t('brandTitle')}</div>
              <div style={{ fontSize: '0.7rem', color: '#C9B49E', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{t('brandSub')}</div>
            </div>
          </div>
        </div>

        {/* Language switch button */}
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
            <Globe size={14} />
            {language === 'en' ? 'ಕನ್ನಡಕ್ಕೆ ಬದಲಾಯಿಸಿ' : 'Switch to English'}
          </button>
        </div>

        {/* Navigation */}
        <nav style={{ padding: '1rem', flex: 1, display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
          {navItems.map((item) => {
            const Icon = item.icon;
            const active = currentTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => {
                  setCurrentTab(item.id);
                }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.75rem',
                  padding: '0.75rem 1rem',
                  borderRadius: '0.5rem',
                  color: active ? '#FFFCEF' : '#D9C7B2',
                  backgroundColor: active ? 'rgba(244,180,0,0.16)' : 'transparent',
                  fontWeight: active ? 600 : 500,
                  fontSize: '0.9rem',
                  width: '100%',
                  textAlign: 'left',
                  transition: 'background-color 0.15s, color 0.15s',
                }}
              >
                <Icon size={18} color={active ? '#F4B400' : '#C9B49E'} />
                {item.label}
              </button>
            );
          })}
        </nav>

        {/* User Info & Logout */}
        <div style={{ padding: '1.25rem', borderTop: '1px solid rgba(255,252,239,0.12)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ overflow: 'hidden' }}>
            <div style={{ fontSize: '0.875rem', fontWeight: 600, color: '#FFFCEF', whiteSpace: 'nowrap', textOverflow: 'ellipsis', overflow: 'hidden' }}>
              {user.phone}
            </div>
            <div style={{ fontSize: '0.75rem', color: '#B9A48F' }}>
              {user.roles.join(', ')}
            </div>
          </div>
          <button
            onClick={logout}
            title={t('logout')}
            style={{ padding: '0.5rem', borderRadius: '0.375rem', color: '#D9C7B2', background: 'rgba(255,252,239,0.08)' }}
          >
            <LogOut size={16} />
          </button>
        </div>
      </aside>

      {/* Main Content Area */}
      <main style={{ marginLeft: '260px', flex: 1, padding: '2rem 2.5rem', maxWidth: '1400px' }}>
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

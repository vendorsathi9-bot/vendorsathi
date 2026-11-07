import React, { useState } from 'react';
import { useAuth } from '../core/auth/AuthContext';
import { LoginPage } from '../features/auth/LoginPage';
import { DashboardPage } from '../features/dashboard/DashboardPage';
import { VendorsPage } from '../features/vendors/VendorsPage';
import { CustomersPage } from '../features/customers/CustomersPage';
import { OrdersPage } from '../features/orders/OrdersPage';
import { AuditLogsPage } from '../features/audit/AuditLogsPage';
import { 
  LayoutDashboard, 
  Store, 
  Users, 
  ShoppingBag, 
  ShieldCheck, 
  LogOut
} from 'lucide-react';

export const App: React.FC = () => {
  const { user, loading, logout, isAdmin } = useAuth();
  const [currentTab, setCurrentTab] = useState<'dashboard' | 'vendors' | 'customers' | 'orders' | 'audit'>('dashboard');

  if (loading) {
    return (
      <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#64748b' }}>
        Verifying admin session...
      </div>
    );
  }

  if (!user || !isAdmin) {
    return <LoginPage />;
  }

  const navItems = [
    { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { id: 'vendors', label: 'Vendors', icon: Store },
    { id: 'customers', label: 'Customers', icon: Users },
    { id: 'orders', label: 'Orders', icon: ShoppingBag },
    { id: 'audit', label: 'Audit Trail', icon: ShieldCheck },
  ] as const;

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      {/* Sidebar */}
      <aside style={{
        width: '260px',
        backgroundColor: '#0f172a',
        color: 'white',
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
        <div style={{ padding: '1.5rem', borderBottom: '1px solid #1e293b', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <div style={{ padding: '0.5rem', background: '#2563eb', borderRadius: '0.5rem' }}>
            <ShieldCheck size={22} color="white" />
          </div>
          <div>
            <div style={{ fontWeight: 700, fontSize: '1.1rem', letterSpacing: '-0.025em' }}>VendorSaathi</div>
            <div style={{ fontSize: '0.7rem', color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Administration</div>
          </div>
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
                  color: active ? '#ffffff' : '#94a3b8',
                  backgroundColor: active ? '#1e293b' : 'transparent',
                  fontWeight: active ? 600 : 500,
                  fontSize: '0.9rem',
                  width: '100%',
                  textAlign: 'left',
                  transition: 'background-color 0.15s, color 0.15s',
                }}
              >
                <Icon size={18} color={active ? '#3b82f6' : '#94a3b8'} />
                {item.label}
              </button>
            );
          })}
        </nav>

        {/* User Info & Logout */}
        <div style={{ padding: '1.25rem', borderTop: '1px solid #1e293b', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ overflow: 'hidden' }}>
            <div style={{ fontSize: '0.875rem', fontWeight: 600, color: 'white', whiteSpace: 'nowrap', textOverflow: 'ellipsis', overflow: 'hidden' }}>
              {user.phone}
            </div>
            <div style={{ fontSize: '0.75rem', color: '#64748b' }}>
              {user.roles.join(', ')}
            </div>
          </div>
          <button
            onClick={logout}
            title="Log out"
            style={{ padding: '0.5rem', borderRadius: '0.375rem', color: '#94a3b8', background: '#1e293b' }}
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
        {currentTab === 'audit' && <AuditLogsPage />}
      </main>
    </div>
  );
};

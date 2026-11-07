import React, { useEffect, useState } from 'react';
import { ApiClient, PlatformOverview } from '../../core/api/apiClient';
import { Store, Users, ShoppingBag, IndianRupee, CheckCircle, TrendingUp } from 'lucide-react';

export const DashboardPage: React.FC = () => {
  const [overview, setOverview] = useState<PlatformOverview | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchOverview = async () => {
      try {
        const data = await ApiClient.getOverview();
        setOverview(data);
      } catch (err) {
        console.error('Failed to load overview', err);
      } finally {
        setLoading(false);
      }
    };
    fetchOverview();
  }, []);

  if (loading) {
    return <div style={{ padding: '2rem', color: '#64748b' }}>Loading metrics...</div>;
  }

  const formatINR = (paise: number) => {
    return '₹' + (paise / 100).toLocaleString('en-IN', { maximumFractionDigits: 2 });
  };

  return (
    <div>
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a' }}>Platform Overview</h1>
        <p style={{ color: '#64748b', fontSize: '0.9rem', marginTop: '0.25rem' }}>Real-time metrics, order volume, and financial aggregates across VendorSaathi.</p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.25rem', marginBottom: '2rem' }}>
        <div className="card" style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{ padding: '0.875rem', borderRadius: '0.75rem', background: '#eff6ff', color: '#2563eb' }}>
            <Store size={28} />
          </div>
          <div>
            <div style={{ fontSize: '0.875rem', color: '#64748b', fontWeight: 500 }}>Active Vendors</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#0f172a' }}>{overview?.active_vendors ?? 0}</div>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Total registered: {overview?.total_vendors ?? 0}</div>
          </div>
        </div>

        <div className="card" style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{ padding: '0.875rem', borderRadius: '0.75rem', background: '#f0fdf4', color: '#16a34a' }}>
            <Users size={28} />
          </div>
          <div>
            <div style={{ fontSize: '0.875rem', color: '#64748b', fontWeight: 500 }}>Total Customers</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#0f172a' }}>{overview?.total_customers ?? 0}</div>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>All platform users: {overview?.total_users ?? 0}</div>
          </div>
        </div>

        <div className="card" style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{ padding: '0.875rem', borderRadius: '0.75rem', background: '#fefce8', color: '#ca8a04' }}>
            <ShoppingBag size={28} />
          </div>
          <div>
            <div style={{ fontSize: '0.875rem', color: '#64748b', fontWeight: 500 }}>Completed Orders</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#0f172a' }}>{overview?.completed_orders ?? 0}</div>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Total orders: {overview?.total_orders ?? 0}</div>
          </div>
        </div>

        <div className="card" style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{ padding: '0.875rem', borderRadius: '0.75rem', background: '#faf5ff', color: '#9333ea' }}>
            <IndianRupee size={28} />
          </div>
          <div>
            <div style={{ fontSize: '0.875rem', color: '#64748b', fontWeight: 500 }}>Gross Sales Value</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#0f172a' }}>{formatINR(overview?.total_gross_sales_paise ?? 0)}</div>
            <div style={{ fontSize: '0.75rem', color: '#16a34a', display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
              <TrendingUp size={12} /> Collections: {formatINR(overview?.total_collections_paise ?? 0)}
            </div>
          </div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '1.5rem' }}>
        <div className="card">
          <h2 style={{ fontSize: '1.125rem', fontWeight: 600, marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <CheckCircle size={18} color="#2563eb" /> System Operational Health
          </h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem 0', borderBottom: '1px solid #f1f5f9' }}>
              <span style={{ color: '#64748b', fontSize: '0.875rem' }}>Core PostgreSQL DB</span>
              <span className="badge badge-success">ONLINE</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem 0', borderBottom: '1px solid #f1f5f9' }}>
              <span style={{ color: '#64748b', fontSize: '0.875rem' }}>REST API Services</span>
              <span className="badge badge-success">HEALTHY</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem 0', borderBottom: '1px solid #f1f5f9' }}>
              <span style={{ color: '#64748b', fontSize: '0.875rem' }}>Offline Queue Sync Processor</span>
              <span className="badge badge-info">IDLE</span>
            </div>
          </div>
        </div>

        <div className="card">
          <h2 style={{ fontSize: '1.125rem', fontWeight: 600, marginBottom: '1rem' }}>Quick Actions</h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
            <a href="#/vendors" className="btn btn-secondary" style={{ textDecoration: 'none' }}>
              Review Vendor Onboarding
            </a>
            <a href="#/orders" className="btn btn-secondary" style={{ textDecoration: 'none' }}>
              Monitor Active Orders
            </a>
            <a href="#/audit-logs" className="btn btn-secondary" style={{ textDecoration: 'none' }}>
              View Security & Audit Trail
            </a>
          </div>
        </div>
      </div>
    </div>
  );
};

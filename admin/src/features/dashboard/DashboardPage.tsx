import React, { useEffect, useState } from 'react';
import { ApiClient, PlatformOverview } from '../../core/api/apiClient';
import { useLanguage } from '../../core/i18n/LanguageContext';
import { 
  Store, 
  Users, 
  ShoppingBag, 
  IndianRupee, 
  CheckCircle, 
  TrendingUp, 
  Package, 
  Layers, 
  RefreshCw,
  Navigation
} from 'lucide-react';

export const DashboardPage: React.FC = () => {
  const { t, language } = useLanguage();
  const [overview, setOverview] = useState<PlatformOverview | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const fetchOverview = async () => {
    try {
      const data = await ApiClient.getOverview();
      setOverview(data);
    } catch (err) {
      console.error('Failed to load overview', err);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    fetchOverview();
  }, []);

  const handleRefresh = () => {
    setRefreshing(true);
    fetchOverview();
  };

  const formatINR = (paise: number) => {
    return '₹' + (paise / 100).toLocaleString(language === 'kn' ? 'kn-IN' : 'en-IN', { maximumFractionDigits: 2 });
  };

  if (loading) {
    return (
      <div style={{ padding: '3rem', textAlign: 'center', color: '#64748b' }}>
        <RefreshCw size={24} className="animate-spin" style={{ margin: '0 auto 1rem', display: 'block' }} />
        {t('loading')}
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a' }}>{t('overviewTitle')}</h1>
          <p style={{ color: '#64748b', fontSize: '0.9rem', marginTop: '0.25rem' }}>{t('overviewSubtitle')}</p>
        </div>
        <button 
          onClick={handleRefresh} 
          className="btn btn-secondary"
          disabled={refreshing}
          style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}
        >
          <RefreshCw size={14} className={refreshing ? 'animate-spin' : ''} />
          {t('refresh')}
        </button>
      </div>

      {/* Primary KPI Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.25rem', marginBottom: '2rem' }}>
        {/* Vendors */}
        <div className="card" style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{ padding: '0.875rem', borderRadius: '0.75rem', background: '#eff6ff', color: '#2563eb' }}>
            <Store size={28} />
          </div>
          <div>
            <div style={{ fontSize: '0.875rem', color: '#64748b', fontWeight: 500 }}>{t('activeVendors')}</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#0f172a' }}>{overview?.active_vendors ?? 0}</div>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>{t('totalRegisteredVendors')}: {overview?.total_vendors ?? 0}</div>
          </div>
        </div>

        {/* Customers */}
        <div className="card" style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{ padding: '0.875rem', borderRadius: '0.75rem', background: '#f0fdf4', color: '#16a34a' }}>
            <Users size={28} />
          </div>
          <div>
            <div style={{ fontSize: '0.875rem', color: '#64748b', fontWeight: 500 }}>{t('totalCustomers')}</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#0f172a' }}>{overview?.total_customers ?? 0}</div>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>{t('allPlatformUsers')}: {overview?.total_users ?? 0}</div>
          </div>
        </div>

        {/* Orders */}
        <div className="card" style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{ padding: '0.875rem', borderRadius: '0.75rem', background: '#fefce8', color: '#ca8a04' }}>
            <ShoppingBag size={28} />
          </div>
          <div>
            <div style={{ fontSize: '0.875rem', color: '#64748b', fontWeight: 500 }}>{t('completedOrders')}</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#0f172a' }}>{overview?.completed_orders ?? 0}</div>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>{t('totalOrdersRecorded')}: {overview?.total_orders ?? 0}</div>
          </div>
        </div>

        {/* Gross Sales */}
        <div className="card" style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{ padding: '0.875rem', borderRadius: '0.75rem', background: '#faf5ff', color: '#9333ea' }}>
            <IndianRupee size={28} />
          </div>
          <div>
            <div style={{ fontSize: '0.875rem', color: '#64748b', fontWeight: 500 }}>{t('grossSalesValue')}</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#0f172a' }}>{formatINR(overview?.total_gross_sales_paise ?? 0)}</div>
            <div style={{ fontSize: '0.75rem', color: '#16a34a', display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
              <TrendingUp size={12} /> {t('collectionsLabel')}: {formatINR(overview?.total_collections_paise ?? 0)}
            </div>
          </div>
        </div>
      </div>

      {/* Secondary Metrics: Products, Categories & Health */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '1.5rem', marginBottom: '2rem' }}>
        {/* Products & Catalog Card */}
        <div className="card">
          <h2 style={{ fontSize: '1.125rem', fontWeight: 600, marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#0f172a' }}>
            <Package size={18} color="#2563eb" /> Product & Category Statistics
          </h2>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
            <div style={{ padding: '1rem', background: '#f8fafc', borderRadius: '0.5rem', border: '1px solid #e2e8f0' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#64748b', fontSize: '0.8rem', fontWeight: 500, marginBottom: '0.25rem' }}>
                <Package size={14} /> {t('totalProducts')}
              </div>
              <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#0f172a' }}>{overview?.total_products ?? 0}</div>
              <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>{t('catalogCatalogued')}</div>
            </div>

            <div style={{ padding: '1rem', background: '#f8fafc', borderRadius: '0.5rem', border: '1px solid #e2e8f0' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#64748b', fontSize: '0.8rem', fontWeight: 500, marginBottom: '0.25rem' }}>
                <Layers size={14} /> {t('totalCategories')}
              </div>
              <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#0f172a' }}>{overview?.total_categories ?? 0}</div>
              <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Local commerce taxonomy</div>
            </div>
          </div>
        </div>

        {/* Operational Health */}
        <div className="card">
          <h2 style={{ fontSize: '1.125rem', fontWeight: 600, marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#0f172a' }}>
            <CheckCircle size={18} color="#16a34a" /> {t('systemHealth')}
          </h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid #f1f5f9' }}>
              <span style={{ color: '#475569', fontSize: '0.875rem' }}>{t('coreDb')}</span>
              <span className="badge badge-success">{t('online')}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid #f1f5f9' }}>
              <span style={{ color: '#475569', fontSize: '0.875rem' }}>{t('apiServices')}</span>
              <span className="badge badge-success">{t('healthy')}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid #f1f5f9' }}>
              <span style={{ color: '#475569', fontSize: '0.875rem', display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                <Navigation size={13} color="#2563eb" /> {t('gpsMobility')}
              </span>
              <span className="badge badge-success">{t('active')}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

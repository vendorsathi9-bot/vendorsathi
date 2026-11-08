import React, { useEffect, useState } from 'react';
import { ApiClient, VendorItem } from '../../core/api/apiClient';
import { useLanguage } from '../../core/i18n/LanguageContext';
import { 
  Store, 
  Check, 
  ShieldAlert, 
  Clock, 
  MapPin, 
  Package, 
  RefreshCw, 
  Filter
} from 'lucide-react';

export const VendorsPage: React.FC = () => {
  const { t } = useLanguage();
  const [vendors, setVendors] = useState<VendorItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'ALL' | 'PENDING' | 'ACTIVE' | 'SUSPENDED'>('ALL');
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);

  const loadVendors = async () => {
    setLoading(true);
    try {
      const data = await ApiClient.listVendors();
      setVendors(data || []);
    } catch (err) {
      console.error('Failed to load vendors', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadVendors();
  }, []);

  const handleStatusChange = async (vendorId: string, newStatus: 'ACTIVE' | 'SUSPENDED' | 'PENDING') => {
    const confirmMessage = newStatus === 'ACTIVE'
      ? 'Approve and activate this vendor on VendorSaathi?'
      : newStatus === 'SUSPENDED'
      ? 'Suspend this vendor? Their products will be hidden from discovery.'
      : 'Reset vendor status to pending?';

    if (!window.confirm(confirmMessage)) return;

    setActionLoadingId(vendorId);
    try {
      await ApiClient.updateVendorStatus(vendorId, newStatus);
      await loadVendors();
    } catch (err: any) {
      alert(`Action failed: ${err.message}`);
    } finally {
      setActionLoadingId(null);
    }
  };

  const filteredVendors = vendors.filter(v => {
    if (filter === 'ALL') return true;
    return v.status === filter;
  });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'ACTIVE':
        return <span className="badge badge-success">ACTIVE</span>;
      case 'PENDING':
        return <span className="badge badge-warning">PENDING APPROVAL</span>;
      case 'SUSPENDED':
        return <span className="badge badge-danger">SUSPENDED</span>;
      default:
        return <span className="badge badge-neutral">{status}</span>;
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a' }}>{t('vendorTitle')}</h1>
          <p style={{ color: '#64748b', fontSize: '0.9rem', marginTop: '0.25rem' }}>{t('vendorSubtitle')}</p>
        </div>
        <button onClick={loadVendors} className="btn btn-secondary" disabled={loading}>
          <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
          {t('refresh')}
        </button>
      </div>

      {/* Filter Tabs */}
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem', borderBottom: '1px solid #e2e8f0', paddingBottom: '0.75rem' }}>
        {(['ALL', 'PENDING', 'ACTIVE', 'SUSPENDED'] as const).map(tab => {
          const count = tab === 'ALL' ? vendors.length : vendors.filter(v => v.status === tab).length;
          const active = filter === tab;
          return (
            <button
              key={tab}
              onClick={() => setFilter(tab)}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.35rem',
                padding: '0.4rem 0.85rem',
                borderRadius: '0.375rem',
                fontSize: '0.85rem',
                fontWeight: active ? 600 : 500,
                backgroundColor: active ? '#2563eb' : '#f1f5f9',
                color: active ? '#ffffff' : '#475569',
                transition: 'all 0.15s ease-in-out',
              }}
            >
              <Filter size={13} />
              {tab === 'ALL' ? t('filterAll') : tab === 'PENDING' ? t('filterPending') : tab === 'ACTIVE' ? t('filterActive') : t('filterSuspended')}
              <span style={{
                background: active ? 'rgba(255,255,255,0.25)' : '#e2e8f0',
                padding: '0.1rem 0.4rem',
                borderRadius: '9999px',
                fontSize: '0.75rem',
              }}>
                {count}
              </span>
            </button>
          );
        })}
      </div>

      {/* Vendors Table */}
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>{t('businessName')}</th>
              <th>{t('phone')}</th>
              <th>{t('location')}</th>
              <th>{t('productsCount')}</th>
              <th>{t('status')}</th>
              <th>{t('joinedDate')}</th>
              <th style={{ textAlign: 'right' }}>{t('actions')}</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={7} style={{ textAlign: 'center', padding: '3rem', color: '#64748b' }}>
                  <RefreshCw size={24} className="animate-spin" style={{ margin: '0 auto 0.75rem', display: 'block' }} />
                  {t('loading')}
                </td>
              </tr>
            ) : filteredVendors.length === 0 ? (
              <tr>
                <td colSpan={7} style={{ textAlign: 'center', padding: '3.5rem', color: '#64748b' }}>
                  <Store size={36} style={{ margin: '0 auto 0.75rem', opacity: 0.4, display: 'block' }} />
                  <div style={{ fontWeight: 600, fontSize: '0.95rem' }}>{t('noVendorsFound')}</div>
                </td>
              </tr>
            ) : (
              filteredVendors.map(v => (
                <tr key={v.id}>
                  <td>
                    <div style={{ fontWeight: 600, color: '#0f172a' }}>{v.business_name}</div>
                    <div style={{ fontSize: '0.75rem', color: '#64748b', fontFamily: 'monospace' }}>ID: {v.id.substring(0, 8)}...</div>
                  </td>
                  <td style={{ fontWeight: 500 }}>{v.phone}</td>
                  <td>
                    {v.city || v.address_line1 ? (
                      <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem', fontSize: '0.85rem', color: '#475569' }}>
                        <MapPin size={13} color="#2563eb" />
                        {v.city ? `${v.city}${v.address_line1 ? `, ${v.address_line1}` : ''}` : v.address_line1}
                      </span>
                    ) : (
                      <span style={{ color: '#94a3b8', fontSize: '0.8rem' }}>Not configured</span>
                    )}
                  </td>
                  <td>
                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.3rem', fontSize: '0.85rem', fontWeight: 600, color: '#0f172a' }}>
                      <Package size={14} color="#64748b" />
                      {v.product_count}
                    </span>
                  </td>
                  <td>
                    {getStatusBadge(v.status)}
                    <div style={{ fontSize: '0.7rem', color: '#94a3b8', marginTop: '0.15rem' }}>
                      Shop: {v.availability_status}
                    </div>
                  </td>
                  <td style={{ color: '#64748b', fontSize: '0.85rem' }}>
                    {new Date(v.created_at).toLocaleDateString()}
                  </td>
                  <td style={{ textAlign: 'right' }}>
                    <div style={{ display: 'inline-flex', gap: '0.4rem', justifyContent: 'flex-end' }}>
                      {v.status === 'PENDING' && (
                        <button
                          onClick={() => handleStatusChange(v.id, 'ACTIVE')}
                          disabled={actionLoadingId === v.id}
                          className="btn btn-primary"
                          style={{ padding: '0.35rem 0.65rem', fontSize: '0.75rem' }}
                          title="Approve this vendor profile"
                        >
                          <Check size={14} /> {t('approveVendor')}
                        </button>
                      )}

                      {v.status === 'ACTIVE' && (
                        <button
                          onClick={() => handleStatusChange(v.id, 'SUSPENDED')}
                          disabled={actionLoadingId === v.id}
                          className="btn btn-danger"
                          style={{ padding: '0.35rem 0.65rem', fontSize: '0.75rem' }}
                          title="Suspend vendor access"
                        >
                          <ShieldAlert size={14} /> {t('suspendVendor')}
                        </button>
                      )}

                      {v.status === 'SUSPENDED' && (
                        <button
                          onClick={() => handleStatusChange(v.id, 'ACTIVE')}
                          disabled={actionLoadingId === v.id}
                          className="btn btn-primary"
                          style={{ padding: '0.35rem 0.65rem', fontSize: '0.75rem' }}
                          title="Re-activate vendor"
                        >
                          <Check size={14} /> {t('activateVendor')}
                        </button>
                      )}

                      {v.status !== 'PENDING' && (
                        <button
                          onClick={() => handleStatusChange(v.id, 'PENDING')}
                          disabled={actionLoadingId === v.id}
                          className="btn btn-secondary"
                          style={{ padding: '0.35rem 0.5rem', fontSize: '0.75rem' }}
                          title="Set back to pending"
                        >
                          <Clock size={13} />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

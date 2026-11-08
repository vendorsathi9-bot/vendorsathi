import React, { useEffect, useState } from 'react';
import { ApiClient, Order } from '../../core/api/apiClient';
import { useLanguage } from '../../core/i18n/LanguageContext';
import { 
  ShoppingBag, 
  Eye, 
  MapPin, 
  RefreshCw, 
  Filter, 
  Tag
} from 'lucide-react';

export const OrdersPage: React.FC = () => {
  const { t } = useLanguage();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null);
  const [statusFilter, setStatusFilter] = useState<'ALL' | 'PENDING' | 'ACTIVE' | 'COMPLETED' | 'CANCELLED'>('ALL');

  const loadOrders = async () => {
    setLoading(true);
    try {
      const res = await ApiClient.listOrders(100);
      setOrders(res.orders || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadOrders();
  }, []);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'COMPLETED':
        return 'badge-success';
      case 'PENDING':
      case 'PREPARING':
      case 'READY_FOR_PICKUP':
      case 'OUT_FOR_DELIVERY':
        return 'badge-warning';
      case 'REJECTED':
      case 'CANCELLED':
        return 'badge-danger';
      default:
        return 'badge-info';
    }
  };

  const filteredOrders = orders.filter(o => {
    if (statusFilter === 'ALL') return true;
    if (statusFilter === 'PENDING') return o.status === 'PENDING';
    if (statusFilter === 'ACTIVE') return ['ACCEPTED', 'PREPARING', 'READY_FOR_PICKUP', 'OUT_FOR_DELIVERY'].includes(o.status);
    if (statusFilter === 'COMPLETED') return o.status === 'COMPLETED';
    if (statusFilter === 'CANCELLED') return ['CANCELLED', 'REJECTED'].includes(o.status);
    return true;
  });

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a' }}>{t('orderTitle')}</h1>
          <p style={{ color: '#64748b', fontSize: '0.9rem', marginTop: '0.25rem' }}>{t('orderSubtitle')}</p>
        </div>
        <button onClick={loadOrders} className="btn btn-secondary" disabled={loading}>
          <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
          {t('refresh')}
        </button>
      </div>

      {/* Filter Tabs */}
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem', borderBottom: '1px solid #e2e8f0', paddingBottom: '0.75rem', flexWrap: 'wrap' }}>
        {(['ALL', 'PENDING', 'ACTIVE', 'COMPLETED', 'CANCELLED'] as const).map(tab => {
          const count = orders.filter(o => {
            if (tab === 'ALL') return true;
            if (tab === 'PENDING') return o.status === 'PENDING';
            if (tab === 'ACTIVE') return ['ACCEPTED', 'PREPARING', 'READY_FOR_PICKUP', 'OUT_FOR_DELIVERY'].includes(o.status);
            if (tab === 'COMPLETED') return o.status === 'COMPLETED';
            if (tab === 'CANCELLED') return ['CANCELLED', 'REJECTED'].includes(o.status);
            return true;
          }).length;
          const active = statusFilter === tab;
          return (
            <button
              key={tab}
              onClick={() => setStatusFilter(tab)}
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
              {tab === 'ALL' ? t('filterAll') : tab === 'PENDING' ? 'Pending' : tab === 'ACTIVE' ? 'In Progress' : tab === 'COMPLETED' ? 'Completed' : 'Cancelled'}
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

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>{t('orderId')}</th>
              <th>{t('fulfilment')}</th>
              <th>{t('status')}</th>
              <th>{t('items')}</th>
              <th>{t('amount')}</th>
              <th>{t('date')}</th>
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
            ) : filteredOrders.length === 0 ? (
              <tr>
                <td colSpan={7} style={{ textAlign: 'center', padding: '3.5rem', color: '#64748b' }}>
                  <ShoppingBag size={36} style={{ margin: '0 auto 0.75rem', opacity: 0.4, display: 'block' }} />
                  <div style={{ fontWeight: 600, fontSize: '0.95rem' }}>No orders found for this filter.</div>
                </td>
              </tr>
            ) : (
              filteredOrders.map(o => (
                <tr key={o.id}>
                  <td style={{ fontFamily: 'monospace', fontSize: '0.8rem', color: '#0f172a', fontWeight: 600 }}>
                    #{o.id.substring(0, 8)}
                  </td>
                  <td>
                    <span className="badge badge-neutral" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}>
                      <MapPin size={12} /> {o.fulfilment_type}
                    </span>
                  </td>
                  <td>
                    <span className={`badge ${getStatusBadge(o.status)}`}>
                      {o.status}
                    </span>
                  </td>
                  <td>{o.items?.length || 0} items</td>
                  <td style={{ fontWeight: 700, color: '#0f172a' }}>
                    ₹{(o.total_paise / 100).toFixed(2)}
                    {(o.discount_paise ?? 0) > 0 && (
                      <div style={{ fontSize: '0.7rem', color: '#16a34a', display: 'flex', alignItems: 'center', gap: '0.2rem' }}>
                        <Tag size={10} /> -₹{((o.discount_paise ?? 0) / 100).toFixed(2)}
                      </div>
                    )}
                  </td>
                  <td style={{ color: '#64748b', fontSize: '0.85rem' }}>
                    {new Date(o.created_at).toLocaleString()}
                  </td>
                  <td style={{ textAlign: 'right' }}>
                    <button
                      onClick={() => setSelectedOrder(o)}
                      className="btn btn-secondary"
                      style={{ padding: '0.35rem 0.65rem', fontSize: '0.75rem' }}
                    >
                      <Eye size={14} /> {t('viewDetails')}
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {selectedOrder && (
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '1rem', zIndex: 50 }}>
          <div style={{ background: 'white', borderRadius: '1rem', maxWidth: '550px', width: '100%', padding: '1.75rem', maxHeight: '90vh', overflowY: 'auto' }}>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 700, marginBottom: '0.25rem' }}>{t('orderDetails')}</h2>
            <div style={{ fontSize: '0.85rem', color: '#64748b', marginBottom: '1.25rem', fontFamily: 'monospace' }}>
              ID: {selectedOrder.id}
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', marginBottom: '1.5rem', background: '#f8fafc', padding: '1rem', borderRadius: '0.5rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ color: '#64748b', fontSize: '0.875rem' }}>{t('status')}</span>
                <span className={`badge ${getStatusBadge(selectedOrder.status)}`}>{selectedOrder.status}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ color: '#64748b', fontSize: '0.875rem' }}>{t('fulfilment')}</span>
                <span style={{ fontWeight: 600 }}>{selectedOrder.fulfilment_type}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ color: '#64748b', fontSize: '0.875rem' }}>Customer ID</span>
                <span style={{ fontFamily: 'monospace', fontSize: '0.8rem' }}>{selectedOrder.customer_id}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ color: '#64748b', fontSize: '0.875rem' }}>Vendor ID</span>
                <span style={{ fontFamily: 'monospace', fontSize: '0.8rem' }}>{selectedOrder.vendor_id}</span>
              </div>
            </div>

            <h3 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '0.75rem' }}>{t('items')}</h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginBottom: '1.5rem' }}>
              {selectedOrder.items?.map(it => (
                <div key={it.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.875rem', padding: '0.6rem 0.75rem', background: '#f8fafc', borderRadius: '0.375rem' }}>
                  <div>
                    <div style={{ fontWeight: 600, color: '#0f172a' }}>{it.product_name_snapshot}</div>
                    <div style={{ fontSize: '0.75rem', color: '#64748b' }}>₹{(it.unit_price_paise / 100).toFixed(2)} × {it.quantity}</div>
                  </div>
                  <span style={{ fontWeight: 700, color: '#0f172a' }}>₹{(it.line_total_paise / 100).toFixed(2)}</span>
                </div>
              ))}
            </div>

            {/* Financial summary */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', borderTop: '1px solid #e2e8f0', paddingTop: '1rem', marginBottom: '1.5rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.9rem', color: '#64748b' }}>
                <span>{t('subtotal')}</span>
                <span>₹{(selectedOrder.subtotal_paise / 100).toFixed(2)}</span>
              </div>
              {(selectedOrder.discount_paise ?? 0) > 0 && (
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.9rem', color: '#16a34a' }}>
                  <span>{t('discount')}</span>
                  <span>-₹{((selectedOrder.discount_paise ?? 0) / 100).toFixed(2)}</span>
                </div>
              )}
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '1.1rem', fontWeight: 700, color: '#0f172a' }}>
                <span>{t('total')}</span>
                <span>₹{(selectedOrder.total_paise / 100).toFixed(2)}</span>
              </div>
            </div>

            <button onClick={() => setSelectedOrder(null)} className="btn btn-secondary" style={{ width: '100%', justifyContent: 'center' }}>
              {t('close')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

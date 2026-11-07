import React, { useEffect, useState } from 'react';
import { ApiClient, Order } from '../../core/api/apiClient';
import { ShoppingBag, Eye, MapPin } from 'lucide-react';

export const OrdersPage: React.FC = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null);

  const loadOrders = async () => {
    setLoading(true);
    try {
      const res = await ApiClient.listOrders(50);
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

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a' }}>Order Monitoring</h1>
          <p style={{ color: '#64748b', fontSize: '0.9rem', marginTop: '0.25rem' }}>Track lifecycle transitions, fulfilment modes, and financial line items.</p>
        </div>
        <button onClick={loadOrders} className="btn btn-secondary">Refresh</button>
      </div>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Order ID</th>
              <th>Fulfilment</th>
              <th>Status</th>
              <th>Items</th>
              <th>Total Amount</th>
              <th>Date</th>
              <th style={{ textAlign: 'right' }}>Details</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '2rem' }}>Loading orders...</td></tr>
            ) : orders.length === 0 ? (
              <tr>
                <td colSpan={7} style={{ textAlign: 'center', padding: '3rem', color: '#64748b' }}>
                  <ShoppingBag size={32} style={{ margin: '0 auto 0.5rem', opacity: 0.5 }} />
                  <div>No orders recorded in the system yet.</div>
                </td>
              </tr>
            ) : (
              orders.map(o => (
                <tr key={o.id}>
                  <td style={{ fontFamily: 'monospace', fontSize: '0.8rem', color: '#64748b' }}>{o.id.substring(0, 8)}...</td>
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
                  <td style={{ fontWeight: 600 }}>₹{(o.total_paise / 100).toFixed(2)}</td>
                  <td>{new Date(o.created_at).toLocaleString()}</td>
                  <td style={{ textAlign: 'right' }}>
                    <button
                      onClick={() => setSelectedOrder(o)}
                      className="btn btn-secondary"
                      style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                    >
                      <Eye size={14} /> View
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
            <h2 style={{ fontSize: '1.25rem', fontWeight: 700, marginBottom: '0.5rem' }}>Order Details</h2>
            <div style={{ fontSize: '0.85rem', color: '#64748b', marginBottom: '1.25rem', fontFamily: 'monospace' }}>ID: {selectedOrder.id}</div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', marginBottom: '1.5rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid #f1f5f9', paddingBottom: '0.5rem' }}>
                <span style={{ color: '#64748b' }}>Status</span>
                <span className={`badge ${getStatusBadge(selectedOrder.status)}`}>{selectedOrder.status}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid #f1f5f9', paddingBottom: '0.5rem' }}>
                <span style={{ color: '#64748b' }}>Fulfilment</span>
                <span>{selectedOrder.fulfilment_type}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid #f1f5f9', paddingBottom: '0.5rem' }}>
                <span style={{ color: '#64748b' }}>Subtotal</span>
                <span>₹{(selectedOrder.subtotal_paise / 100).toFixed(2)}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid #f1f5f9', paddingBottom: '0.5rem', fontWeight: 700 }}>
                <span>Total Amount</span>
                <span>₹{(selectedOrder.total_paise / 100).toFixed(2)}</span>
              </div>
            </div>

            <h3 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '0.75rem' }}>Purchased Items</h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginBottom: '1.5rem' }}>
              {selectedOrder.items?.map(it => (
                <div key={it.id} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.875rem', padding: '0.5rem', background: '#f8fafc', borderRadius: '0.375rem' }}>
                  <span>{it.product_name_snapshot} × {it.quantity}</span>
                  <span style={{ fontWeight: 500 }}>₹{(it.line_total_paise / 100).toFixed(2)}</span>
                </div>
              ))}
            </div>

            <button onClick={() => setSelectedOrder(null)} className="btn btn-secondary" style={{ width: '100%', justifyContent: 'center' }}>
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

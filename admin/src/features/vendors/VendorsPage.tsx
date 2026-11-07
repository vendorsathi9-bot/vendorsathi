import React, { useEffect, useState } from 'react';
import { ApiClient, User } from '../../core/api/apiClient';
import { Check, ShieldAlert, Store } from 'lucide-react';

export const VendorsPage: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  const loadVendors = async () => {
    setLoading(true);
    try {
      const res = await ApiClient.listUsers(50);
      // Filter users who have the VENDOR role
      const vendorUsers = (res.data || []).filter(u => u.roles?.includes('VENDOR'));
      setUsers(vendorUsers);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadVendors();
  }, []);

  const handleStatusChange = async (userId: string, newStatus: 'ACTIVE' | 'SUSPENDED') => {
    const reason = prompt(`Reason for setting status to ${newStatus}:`);
    if (!reason) return;
    try {
      await ApiClient.updateUserStatus(userId, newStatus);
      await loadVendors();
    } catch (err: any) {
      alert(`Action failed: ${err.message}`);
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a' }}>Vendor Management</h1>
          <p style={{ color: '#64748b', fontSize: '0.9rem', marginTop: '0.25rem' }}>Review onboarding, active vendors, and operational status controls.</p>
        </div>
        <button onClick={loadVendors} className="btn btn-secondary">Refresh</button>
      </div>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Vendor User ID</th>
              <th>Phone</th>
              <th>Roles</th>
              <th>Status</th>
              <th>Joined Date</th>
              <th style={{ textAlign: 'right' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={6} style={{ textAlign: 'center', padding: '2rem' }}>Loading vendors...</td></tr>
            ) : users.length === 0 ? (
              <tr>
                <td colSpan={6} style={{ textAlign: 'center', padding: '3rem', color: '#64748b' }}>
                  <Store size={32} style={{ margin: '0 auto 0.5rem', opacity: 0.5 }} />
                  <div>No registered vendors found.</div>
                </td>
              </tr>
            ) : (
              users.map(u => (
                <tr key={u.id}>
                  <td style={{ fontFamily: 'monospace', fontSize: '0.8rem', color: '#64748b' }}>{u.id}</td>
                  <td style={{ fontWeight: 600 }}>{u.phone}</td>
                  <td>
                    {u.roles?.map(r => (
                      <span key={r} className="badge badge-info" style={{ marginRight: '0.25rem' }}>{r}</span>
                    ))}
                  </td>
                  <td>
                    <span className={`badge ${u.status === 'ACTIVE' ? 'badge-success' : u.status === 'SUSPENDED' ? 'badge-danger' : 'badge-warning'}`}>
                      {u.status}
                    </span>
                  </td>
                  <td>{new Date(u.created_at).toLocaleDateString()}</td>
                  <td style={{ textAlign: 'right' }}>
                    {u.status === 'ACTIVE' ? (
                      <button
                        onClick={() => handleStatusChange(u.id, 'SUSPENDED')}
                        className="btn btn-danger"
                        style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                      >
                        <ShieldAlert size={14} /> Suspend
                      </button>
                    ) : (
                      <button
                        onClick={() => handleStatusChange(u.id, 'ACTIVE')}
                        className="btn btn-primary"
                        style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                      >
                        <Check size={14} /> Activate
                      </button>
                    )}
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

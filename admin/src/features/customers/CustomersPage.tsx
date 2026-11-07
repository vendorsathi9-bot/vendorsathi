import React, { useEffect, useState } from 'react';
import { ApiClient, User } from '../../core/api/apiClient';
import { Users, ShieldAlert, Check } from 'lucide-react';

export const CustomersPage: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  const loadCustomers = async () => {
    setLoading(true);
    try {
      const res = await ApiClient.listUsers(50);
      const customerUsers = (res.data || []).filter(u => u.roles?.includes('CUSTOMER'));
      setUsers(customerUsers);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCustomers();
  }, []);

  const handleStatusChange = async (userId: string, newStatus: 'ACTIVE' | 'SUSPENDED') => {
    const reason = prompt(`Reason for setting status to ${newStatus}:`);
    if (!reason) return;
    try {
      await ApiClient.updateUserStatus(userId, newStatus);
      await loadCustomers();
    } catch (err: any) {
      alert(`Action failed: ${err.message}`);
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a' }}>Customer Management</h1>
          <p style={{ color: '#64748b', fontSize: '0.9rem', marginTop: '0.25rem' }}>View registered buyers, order accounts, and access status.</p>
        </div>
        <button onClick={loadCustomers} className="btn btn-secondary">Refresh</button>
      </div>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Customer ID</th>
              <th>Phone</th>
              <th>Status</th>
              <th>Joined Date</th>
              <th style={{ textAlign: 'right' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '2rem' }}>Loading customers...</td></tr>
            ) : users.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ textAlign: 'center', padding: '3rem', color: '#64748b' }}>
                  <Users size={32} style={{ margin: '0 auto 0.5rem', opacity: 0.5 }} />
                  <div>No registered customers found.</div>
                </td>
              </tr>
            ) : (
              users.map(u => (
                <tr key={u.id}>
                  <td style={{ fontFamily: 'monospace', fontSize: '0.8rem', color: '#64748b' }}>{u.id}</td>
                  <td style={{ fontWeight: 600 }}>{u.phone}</td>
                  <td>
                    <span className={`badge ${u.status === 'ACTIVE' ? 'badge-success' : 'badge-danger'}`}>
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

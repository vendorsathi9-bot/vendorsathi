import React, { useEffect, useState } from 'react';
import { ApiClient, User } from '../../core/api/apiClient';
import { useLanguage } from '../../core/i18n/LanguageContext';
import { 
  Users, 
  ShieldAlert, 
  Check, 
  RefreshCw, 
  UserPlus, 
  XCircle, 
  Filter
} from 'lucide-react';

export const CustomersPage: React.FC = () => {
  const { t } = useLanguage();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [roleFilter, setRoleFilter] = useState<'ALL' | 'CUSTOMER' | 'VENDOR' | 'ADMIN'>('ALL');
  const [managingUserId, setManagingUserId] = useState<string | null>(null);
  const [selectedRole, setSelectedRole] = useState<'CUSTOMER' | 'VENDOR' | 'ADMIN'>('VENDOR');

  const loadUsers = async () => {
    setLoading(true);
    try {
      const res = await ApiClient.listUsers(100);
      setUsers(res.data || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadUsers();
  }, []);

  const handleStatusChange = async (userId: string, newStatus: 'ACTIVE' | 'SUSPENDED') => {
    if (!window.confirm(`Are you sure you want to change user status to ${newStatus}?`)) return;
    try {
      await ApiClient.updateUserStatus(userId, newStatus);
      await loadUsers();
    } catch (err: any) {
      alert(`Action failed: ${err.message}`);
    }
  };

  const handleAssignRole = async (userId: string) => {
    try {
      await ApiClient.assignUserRole(userId, selectedRole);
      setManagingUserId(null);
      await loadUsers();
    } catch (err: any) {
      alert(`Failed to assign role: ${err.message}`);
    }
  };

  const handleRevokeRole = async (userId: string, roleCode: string) => {
    if (!window.confirm(`Revoke role "${roleCode}" from this user?`)) return;
    try {
      await ApiClient.revokeUserRole(userId, roleCode);
      await loadUsers();
    } catch (err: any) {
      alert(`Failed to revoke role: ${err.message}`);
    }
  };

  const filteredUsers = users.filter(u => {
    if (roleFilter === 'ALL') return true;
    return u.roles?.includes(roleFilter);
  });

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 700, color: '#38200F' }}>{t('userTitle')}</h1>
          <p style={{ color: '#806B55', fontSize: '0.9rem', marginTop: '0.25rem' }}>{t('userSubtitle')}</p>
        </div>
        <button onClick={loadUsers} className="btn btn-secondary" disabled={loading}>
          <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
          {t('refresh')}
        </button>
      </div>

      {/* Role Filter Chips */}
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem', borderBottom: '1px solid #EAE1D1', paddingBottom: '0.75rem' }}>
        {(['ALL', 'CUSTOMER', 'VENDOR', 'ADMIN'] as const).map(tab => {
          const count = tab === 'ALL' ? users.length : users.filter(u => u.roles?.includes(tab)).length;
          const active = roleFilter === tab;
          return (
            <button
              key={tab}
              onClick={() => setRoleFilter(tab)}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.35rem',
                padding: '0.4rem 0.85rem',
                borderRadius: '0.375rem',
                fontSize: '0.85rem',
                fontWeight: active ? 600 : 500,
                backgroundColor: active ? '#F4B400' : '#FEF4CD',
                color: active ? '#38200F' : '#806B55',
                transition: 'all 0.15s ease-in-out',
              }}
            >
              <Filter size={13} />
              {tab === 'ALL' ? t('filterAll') : tab === 'CUSTOMER' ? 'Customers' : tab === 'VENDOR' ? 'Vendors' : 'Admins'}
              <span style={{
                background: active ? 'rgba(56,32,15,0.14)' : '#EAE1D1',
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

      {/* Table */}
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>{t('phone')}</th>
              <th>{t('roles')}</th>
              <th>{t('status')}</th>
              <th>{t('joinedDate')}</th>
              <th style={{ textAlign: 'right' }}>{t('actions')}</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={5} style={{ textAlign: 'center', padding: '3rem', color: '#806B55' }}>
                  <RefreshCw size={24} className="animate-spin" style={{ margin: '0 auto 0.75rem', display: 'block' }} />
                  {t('loading')}
                </td>
              </tr>
            ) : filteredUsers.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ textAlign: 'center', padding: '3.5rem', color: '#806B55' }}>
                  <Users size={36} style={{ margin: '0 auto 0.75rem', opacity: 0.4, display: 'block' }} />
                  <div style={{ fontWeight: 600, fontSize: '0.95rem' }}>No users match this filter.</div>
                </td>
              </tr>
            ) : (
              filteredUsers.map(u => (
                <tr key={u.id}>
                  <td>
                    <div style={{ fontWeight: 600, color: '#38200F' }}>{u.phone}</div>
                    <div style={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#806B55' }}>ID: {u.id.substring(0, 8)}...</div>
                  </td>
                  <td>
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.35rem', alignItems: 'center' }}>
                      {u.roles?.map(r => (
                        <span 
                          key={r} 
                          className="badge" 
                          style={{ 
                            background: r === 'ADMIN' ? '#fef3c7' : r === 'VENDOR' ? '#FEF4CD' : '#EAF6EE',
                            color: r === 'ADMIN' ? '#92400e' : r === 'VENDOR' ? '#1e40af' : '#2F7D4F',
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '0.25rem',
                          }}
                        >
                          {r}
                          {u.roles.length > 1 && (
                            <button
                              onClick={() => handleRevokeRole(u.id, r)}
                              title={`Revoke ${r} role`}
                              style={{ color: '#ef4444', display: 'inline-flex', alignItems: 'center' }}
                            >
                              <XCircle size={12} />
                            </button>
                          )}
                        </span>
                      ))}

                      {managingUserId === u.id ? (
                        <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}>
                          <select
                            value={selectedRole}
                            onChange={(e) => setSelectedRole(e.target.value as any)}
                            style={{ padding: '0.2rem 0.4rem', fontSize: '0.75rem', borderRadius: '0.25rem', border: '1px solid #EAE1D1' }}
                          >
                            <option value="VENDOR">VENDOR</option>
                            <option value="CUSTOMER">CUSTOMER</option>
                            <option value="ADMIN">ADMIN</option>
                          </select>
                          <button
                            onClick={() => handleAssignRole(u.id)}
                            className="btn btn-primary"
                            style={{ padding: '0.2rem 0.5rem', fontSize: '0.75rem' }}
                          >
                            Save
                          </button>
                          <button
                            onClick={() => setManagingUserId(null)}
                            className="btn btn-secondary"
                            style={{ padding: '0.2rem 0.4rem', fontSize: '0.75rem' }}
                          >
                            ✕
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={() => setManagingUserId(u.id)}
                          style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '0.2rem',
                            fontSize: '0.75rem',
                            color: '#38200F',
                            padding: '0.2rem 0.4rem',
                            background: '#FEF4CD',
                            borderRadius: '0.25rem',
                            fontWeight: 500,
                          }}
                        >
                          <UserPlus size={11} /> +Role
                        </button>
                      )}
                    </div>
                  </td>
                  <td>
                    <span className={`badge ${u.status === 'ACTIVE' ? 'badge-success' : 'badge-danger'}`}>
                      {u.status}
                    </span>
                  </td>
                  <td style={{ color: '#806B55', fontSize: '0.85rem' }}>
                    {new Date(u.created_at).toLocaleDateString()}
                  </td>
                  <td style={{ textAlign: 'right' }}>
                    {u.status === 'ACTIVE' ? (
                      <button
                        onClick={() => handleStatusChange(u.id, 'SUSPENDED')}
                        className="btn btn-danger"
                        style={{ padding: '0.35rem 0.65rem', fontSize: '0.75rem' }}
                      >
                        <ShieldAlert size={14} /> Suspend
                      </button>
                    ) : (
                      <button
                        onClick={() => handleStatusChange(u.id, 'ACTIVE')}
                        className="btn btn-primary"
                        style={{ padding: '0.35rem 0.65rem', fontSize: '0.75rem' }}
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

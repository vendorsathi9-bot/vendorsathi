import React, { useEffect, useState } from 'react';
import { ApiClient, AuditLog } from '../../core/api/apiClient';
import { Shield, Clock } from 'lucide-react';

export const AuditLogsPage: React.FC = () => {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);

  const loadAuditLogs = async () => {
    setLoading(true);
    try {
      const res = await ApiClient.listAuditLogs(50);
      setLogs(res.data || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAuditLogs();
  }, []);

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a' }}>Platform Audit Trail</h1>
          <p style={{ color: '#64748b', fontSize: '0.9rem', marginTop: '0.25rem' }}>Immutable ledger of privileged actions, status modifications, and admin overrides.</p>
        </div>
        <button onClick={loadAuditLogs} className="btn btn-secondary">Refresh</button>
      </div>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>Action</th>
              <th>Resource Type</th>
              <th>Actor ID</th>
              <th>Reason / Details</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={5} style={{ textAlign: 'center', padding: '2rem' }}>Loading audit logs...</td></tr>
            ) : logs.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ textAlign: 'center', padding: '3rem', color: '#64748b' }}>
                  <Shield size={32} style={{ margin: '0 auto 0.5rem', opacity: 0.5 }} />
                  <div>No audit records found.</div>
                </td>
              </tr>
            ) : (
              logs.map(log => (
                <tr key={log.id}>
                  <td style={{ whiteSpace: 'nowrap', color: '#64748b', display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                    <Clock size={14} /> {new Date(log.created_at).toLocaleString()}
                  </td>
                  <td>
                    <span className="badge badge-info">{log.action}</span>
                  </td>
                  <td style={{ fontWeight: 500 }}>{log.resource_type}</td>
                  <td style={{ fontFamily: 'monospace', fontSize: '0.8rem', color: '#64748b' }}>
                    {log.actor_user_id ? log.actor_user_id.substring(0, 8) + '...' : 'SYSTEM'}
                  </td>
                  <td style={{ color: '#475569' }}>{log.reason || '—'}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

import React, { useEffect, useState } from 'react';
import { ApiClient, DeliveryAssignment, DeliveryLocation, User } from '../../core/api/apiClient';
import { Truck, MapPin, RefreshCw, UserPlus, CheckCircle2, Clock, AlertCircle, Map as MapIcon } from 'lucide-react';
import { DeliveryMapModal } from './DeliveryMapModal';

export const DeliveryAgentsPage: React.FC = () => {
  const [assignments, setAssignments] = useState<DeliveryAssignment[]>([]);
  const [agents, setAgents] = useState<User[]>([]);
  const [allUsers, setAllUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTracking, setActiveTracking] = useState<{ [orderId: string]: DeliveryLocation }>({});
  const [trackingLoading, setTrackingLoading] = useState<{ [orderId: string]: boolean }>({});
  const [selectedAssignmentForMap, setSelectedAssignmentForMap] = useState<DeliveryAssignment | null>(null);

  // Assign modal state
  const [showAssignModal, setShowAssignModal] = useState(false);
  const [selectedOrderId, setSelectedOrderId] = useState('');
  const [selectedAgentId, setSelectedAgentId] = useState('');
  const [assignLoading, setAssignLoading] = useState(false);

  // New agent role modal state
  const [showAddAgentModal, setShowAddAgentModal] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState('');
  const [addAgentLoading, setAddAgentLoading] = useState(false);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [assigns, usersRes] = await Promise.all([
        ApiClient.listDeliveryAssignments(),
        ApiClient.listUsers(50),
      ]);

      setAssignments(assigns);
      const usersList: User[] = usersRes.data || [];
      setAllUsers(usersList);

      const deliveryAgents = usersList.filter((u: User) => u.roles.includes('DELIVERY_AGENT'));
      setAgents(deliveryAgents);
    } catch (err: any) {
      setError(err.message || 'Failed to load delivery data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleTrackLocation = async (orderId: string) => {
    try {
      setTrackingLoading(prev => ({ ...prev, [orderId]: true }));
      const loc = await ApiClient.getDeliveryLocation(orderId);
      setActiveTracking(prev => ({ ...prev, [orderId]: loc }));
    } catch (err: any) {
      alert(err.message || 'Location data not available yet for this order');
    } finally {
      setTrackingLoading(prev => ({ ...prev, [orderId]: false }));
    }
  };

  const handleAssignAgent = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedOrderId.trim() || !selectedAgentId.trim()) return;

    try {
      setAssignLoading(true);
      await ApiClient.assignDeliveryAgent(selectedOrderId.trim(), selectedAgentId);
      setShowAssignModal(false);
      setSelectedOrderId('');
      setSelectedAgentId('');
      await fetchData();
    } catch (err: any) {
      alert(err.message || 'Failed to assign delivery agent');
    } finally {
      setAssignLoading(false);
    }
  };

  const handleAddAgentRole = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedUserId.trim()) return;

    try {
      setAddAgentLoading(true);
      await ApiClient.assignUserRole(selectedUserId, 'DELIVERY_AGENT');
      setShowAddAgentModal(false);
      setSelectedUserId('');
      await fetchData();
    } catch (err: any) {
      alert(err.message || 'Failed to grant DELIVERY_AGENT role');
    } finally {
      setAddAgentLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'DELIVERED':
        return <span style={{ backgroundColor: '#dcfce7', color: '#166534', padding: '0.25rem 0.6rem', borderRadius: '9999px', fontSize: '0.75rem', fontWeight: 600 }}>Delivered</span>;
      case 'OUT_FOR_DELIVERY':
        return <span style={{ backgroundColor: '#fef3c7', color: '#92400e', padding: '0.25rem 0.6rem', borderRadius: '9999px', fontSize: '0.75rem', fontWeight: 600 }}>Out for Delivery</span>;
      case 'PICKED_UP':
        return <span style={{ backgroundColor: '#e0e7ff', color: '#3730a3', padding: '0.25rem 0.6rem', borderRadius: '9999px', fontSize: '0.75rem', fontWeight: 600 }}>Picked Up</span>;
      case 'ACCEPTED':
        return <span style={{ backgroundColor: '#f3e8ff', color: '#6b21a8', padding: '0.25rem 0.6rem', borderRadius: '9999px', fontSize: '0.75rem', fontWeight: 600 }}>Accepted</span>;
      default:
        return <span style={{ backgroundColor: '#f1f5f9', color: '#475569', padding: '0.25rem 0.6rem', borderRadius: '9999px', fontSize: '0.75rem', fontWeight: 600 }}>Assigned</span>;
    }
  };

  return (
    <div>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a', margin: 0 }}>Delivery & Agent Management</h1>
          <p style={{ color: '#64748b', marginTop: '0.25rem', marginBottom: 0 }}>
            Assign delivery agents to orders, monitor active delivery assignments, and track live GPS locations.
          </p>
        </div>
        <div style={{ display: 'flex', gap: '0.75rem' }}>
          <button
            onClick={() => setShowAddAgentModal(true)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              backgroundColor: '#f8fafc',
              border: '1px solid #cbd5e1',
              color: '#334155',
              padding: '0.625rem 1rem',
              borderRadius: '0.5rem',
              fontWeight: 600,
              fontSize: '0.875rem',
              cursor: 'pointer',
            }}
          >
            <UserPlus size={16} />
            Promote to Agent
          </button>
          <button
            onClick={() => setShowAssignModal(true)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              backgroundColor: '#2563eb',
              color: 'white',
              border: 'none',
              padding: '0.625rem 1rem',
              borderRadius: '0.5rem',
              fontWeight: 600,
              fontSize: '0.875rem',
              cursor: 'pointer',
            }}
          >
            <Truck size={16} />
            Assign Agent to Order
          </button>
          <button
            onClick={fetchData}
            title="Refresh"
            style={{
              display: 'flex',
              alignItems: 'center',
              backgroundColor: 'white',
              border: '1px solid #cbd5e1',
              padding: '0.625rem',
              borderRadius: '0.5rem',
              cursor: 'pointer',
            }}
          >
            <RefreshCw size={16} color="#64748b" />
          </button>
        </div>
      </div>

      {error && (
        <div style={{ backgroundColor: '#fef2f2', border: '1px solid #fecaca', color: '#991b1b', padding: '1rem', borderRadius: '0.5rem', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <AlertCircle size={18} />
          <span>{error}</span>
        </div>
      )}

      {/* Agents Overview Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1.25rem', marginBottom: '2rem' }}>
        <div style={{ backgroundColor: 'white', borderRadius: '0.75rem', padding: '1.25rem', border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{ fontSize: '0.875rem', fontWeight: 600, color: '#64748b' }}>Active Delivery Agents</span>
            <div style={{ padding: '0.5rem', backgroundColor: '#eff6ff', borderRadius: '0.5rem', color: '#2563eb' }}>
              <Truck size={20} />
            </div>
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a', marginTop: '0.5rem' }}>
            {agents.length}
          </div>
          <div style={{ fontSize: '0.75rem', color: '#94a3b8', marginTop: '0.25rem' }}>
            Registered personnel with DELIVERY_AGENT role
          </div>
        </div>

        <div style={{ backgroundColor: 'white', borderRadius: '0.75rem', padding: '1.25rem', border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{ fontSize: '0.875rem', fontWeight: 600, color: '#64748b' }}>Active Deliveries</span>
            <div style={{ padding: '0.5rem', backgroundColor: '#fef3c7', borderRadius: '0.5rem', color: '#d97706' }}>
              <Clock size={20} />
            </div>
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a', marginTop: '0.5rem' }}>
            {assignments.filter(a => a.status !== 'DELIVERED').length}
          </div>
          <div style={{ fontSize: '0.75rem', color: '#94a3b8', marginTop: '0.25rem' }}>
            In progress across Bangalore & Karnataka
          </div>
        </div>

        <div style={{ backgroundColor: 'white', borderRadius: '0.75rem', padding: '1.25rem', border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{ fontSize: '0.875rem', fontWeight: 600, color: '#64748b' }}>Completed Deliveries</span>
            <div style={{ padding: '0.5rem', backgroundColor: '#f0fdf4', borderRadius: '0.5rem', color: '#16a34a' }}>
              <CheckCircle2 size={20} />
            </div>
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a', marginTop: '0.5rem' }}>
            {assignments.filter(a => a.status === 'DELIVERED').length}
          </div>
          <div style={{ fontSize: '0.75rem', color: '#94a3b8', marginTop: '0.25rem' }}>
            Successfully fulfilled
          </div>
        </div>
      </div>

      {/* Delivery Assignments Table */}
      <div style={{ backgroundColor: 'white', borderRadius: '0.75rem', border: '1px solid #e2e8f0', overflow: 'hidden', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
        <div style={{ padding: '1.25rem', borderBottom: '1px solid #e2e8f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2 style={{ fontSize: '1.125rem', fontWeight: 700, color: '#0f172a', margin: 0 }}>Delivery Assignments</h2>
          <span style={{ fontSize: '0.875rem', color: '#64748b' }}>{assignments.length} total assignments</span>
        </div>

        {loading ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: '#64748b' }}>Loading assignments...</div>
        ) : assignments.length === 0 ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: '#64748b' }}>
            No delivery assignments recorded yet. Click "Assign Agent to Order" above to dispatch a delivery.
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.875rem' }}>
              <thead>
                <tr style={{ backgroundColor: '#f8fafc', borderBottom: '1px solid #e2e8f0', color: '#475569' }}>
                  <th style={{ padding: '0.75rem 1rem' }}>Order / ID</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Customer & Destination</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Vendor</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Status</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Amount</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Assigned Date</th>
                  <th style={{ padding: '0.75rem 1rem', textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {assignments.map(a => {
                  const loc = activeTracking[a.order_id];
                  const isLocLoading = trackingLoading[a.order_id];

                  return (
                    <React.Fragment key={a.id}>
                      <tr style={{ borderBottom: '1px solid #f1f5f9' }}>
                        <td style={{ padding: '1rem', verticalAlign: 'top' }}>
                          <span style={{ fontFamily: 'monospace', fontSize: '0.75rem', backgroundColor: '#f1f5f9', padding: '0.2rem 0.4rem', borderRadius: '0.25rem', color: '#0f172a' }}>
                            {a.order_id.slice(0, 8)}...
                          </span>
                        </td>
                        <td style={{ padding: '1rem', verticalAlign: 'top' }}>
                          <div style={{ fontWeight: 600, color: '#0f172a' }}>{a.customer_name || 'Customer'}</div>
                          <div style={{ fontSize: '0.75rem', color: '#64748b' }}>{a.customer_phone}</div>
                          <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.25rem' }}>
                            {a.delivery_address_line1} {a.delivery_city ? `, ${a.delivery_city}` : ''}
                          </div>
                        </td>
                        <td style={{ padding: '1rem', verticalAlign: 'top' }}>
                          <div style={{ fontWeight: 600, color: '#0f172a' }}>{a.vendor_name || 'Vendor'}</div>
                          <div style={{ fontSize: '0.75rem', color: '#64748b' }}>{a.vendor_phone}</div>
                        </td>
                        <td style={{ padding: '1rem', verticalAlign: 'top' }}>
                          {getStatusBadge(a.status)}
                        </td>
                        <td style={{ padding: '1rem', verticalAlign: 'top', fontWeight: 600, color: '#0f172a' }}>
                          ₹{((a.order_total_paise || 0) / 100).toFixed(2)}
                        </td>
                        <td style={{ padding: '1rem', verticalAlign: 'top', color: '#64748b', fontSize: '0.75rem' }}>
                          {new Date(a.assigned_at).toLocaleString()}
                        </td>
                        <td style={{ padding: '1rem', verticalAlign: 'top', textAlign: 'right' }}>
                          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.4rem', flexWrap: 'wrap' }}>
                            <button
                              onClick={() => setSelectedAssignmentForMap(a)}
                              title="Open interactive MapLibre / OSRM map"
                              style={{
                                display: 'inline-flex',
                                alignItems: 'center',
                                gap: '0.35rem',
                                backgroundColor: '#2563eb',
                                border: 'none',
                                color: 'white',
                                padding: '0.35rem 0.65rem',
                                borderRadius: '0.375rem',
                                fontSize: '0.75rem',
                                fontWeight: 600,
                                cursor: 'pointer',
                              }}
                            >
                              <MapIcon size={13} />
                              Map
                            </button>
                            <button
                              onClick={() => handleTrackLocation(a.order_id)}
                              disabled={isLocLoading}
                              title="Fetch quick GPS coordinates"
                              style={{
                                display: 'inline-flex',
                                alignItems: 'center',
                                gap: '0.35rem',
                                backgroundColor: '#f8fafc',
                                border: '1px solid #cbd5e1',
                                color: '#475569',
                                padding: '0.35rem 0.65rem',
                                borderRadius: '0.375rem',
                                fontSize: '0.75rem',
                                fontWeight: 600,
                                cursor: 'pointer',
                              }}
                            >
                              <MapPin size={13} />
                              {isLocLoading ? 'Tracking...' : 'GPS'}
                            </button>
                          </div>
                        </td>
                      </tr>
                      {loc && (
                        <tr style={{ backgroundColor: '#f0fdf4', borderBottom: '1px solid #dcfce7' }}>
                          <td colSpan={7} style={{ padding: '0.75rem 1.25rem', fontSize: '0.8rem', color: '#166534' }}>
                            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                <MapPin size={16} color="#16a34a" />
                                <strong>Live Agent Location:</strong> Lat {loc.latitude.toFixed(5)}, Lng {loc.longitude.toFixed(5)}
                                <span style={{ color: '#65a30d', marginLeft: '0.5rem' }}>
                                  (Updated: {new Date(loc.updated_at).toLocaleTimeString()})
                                </span>
                              </div>
                              <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
                                <button
                                  onClick={() => setSelectedAssignmentForMap(a)}
                                  style={{
                                    backgroundColor: '#16a34a',
                                    color: 'white',
                                    border: 'none',
                                    padding: '0.25rem 0.6rem',
                                    borderRadius: '0.25rem',
                                    fontSize: '0.75rem',
                                    fontWeight: 600,
                                    cursor: 'pointer',
                                  }}
                                >
                                  Open in MapLibre
                                </button>
                                <a
                                  href={`https://www.openstreetmap.org/?mlat=${loc.latitude}&mlon=${loc.longitude}#map=16/${loc.latitude}/${loc.longitude}`}
                                  target="_blank"
                                  rel="noreferrer"
                                  style={{ color: '#2563eb', textDecoration: 'underline', fontWeight: 600 }}
                                >
                                  View on OpenStreetMap ↗
                                </a>
                              </div>
                            </div>
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Assign Agent Modal */}
      {showAssignModal && (
        <div style={{
          position: 'fixed',
          top: 0, left: 0, right: 0, bottom: 0,
          backgroundColor: 'rgba(0,0,0,0.5)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
        }}>
          <div style={{ backgroundColor: 'white', borderRadius: '0.75rem', padding: '1.75rem', width: '100%', maxWidth: '480px', boxShadow: '0 20px 25px -5px rgba(0,0,0,0.1)' }}>
            <h3 style={{ margin: '0 0 1rem 0', fontSize: '1.25rem', fontWeight: 700, color: '#0f172a' }}>
              Assign Delivery Agent to Order
            </h3>
            <form onSubmit={handleAssignAgent}>
              <div style={{ marginBottom: '1rem' }}>
                <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: 600, color: '#334155', marginBottom: '0.25rem' }}>
                  Order UUID
                </label>
                <input
                  type="text"
                  placeholder="e.g. 550e8400-e29b-41d4-a716-446655440000"
                  value={selectedOrderId}
                  onChange={e => setSelectedOrderId(e.target.value)}
                  required
                  style={{ width: '100%', padding: '0.5rem 0.75rem', border: '1px solid #cbd5e1', borderRadius: '0.375rem', fontSize: '0.875rem' }}
                />
              </div>

              <div style={{ marginBottom: '1.5rem' }}>
                <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: 600, color: '#334155', marginBottom: '0.25rem' }}>
                  Select Delivery Agent
                </label>
                {agents.length === 0 ? (
                  <div style={{ fontSize: '0.8rem', color: '#dc2626' }}>
                    No delivery agents found. Use "Promote to Agent" to grant a user the DELIVERY_AGENT role first.
                  </div>
                ) : (
                  <select
                    value={selectedAgentId}
                    onChange={e => setSelectedAgentId(e.target.value)}
                    required
                    style={{ width: '100%', padding: '0.5rem 0.75rem', border: '1px solid #cbd5e1', borderRadius: '0.375rem', fontSize: '0.875rem', backgroundColor: 'white' }}
                  >
                    <option value="">-- Choose Agent --</option>
                    {agents.map(ag => (
                      <option key={ag.id} value={ag.id}>
                        {ag.phone} ({ag.name || 'Agent'})
                      </option>
                    ))}
                  </select>
                )}
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
                <button
                  type="button"
                  onClick={() => setShowAssignModal(false)}
                  style={{ padding: '0.5rem 1rem', border: '1px solid #cbd5e1', backgroundColor: 'white', borderRadius: '0.375rem', cursor: 'pointer', fontWeight: 500 }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={assignLoading || agents.length === 0}
                  style={{ padding: '0.5rem 1rem', backgroundColor: '#2563eb', color: 'white', border: 'none', borderRadius: '0.375rem', cursor: 'pointer', fontWeight: 600 }}
                >
                  {assignLoading ? 'Assigning...' : 'Assign Agent'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Promote to Agent Modal */}
      {showAddAgentModal && (
        <div style={{
          position: 'fixed',
          top: 0, left: 0, right: 0, bottom: 0,
          backgroundColor: 'rgba(0,0,0,0.5)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
        }}>
          <div style={{ backgroundColor: 'white', borderRadius: '0.75rem', padding: '1.75rem', width: '100%', maxWidth: '480px', boxShadow: '0 20px 25px -5px rgba(0,0,0,0.1)' }}>
            <h3 style={{ margin: '0 0 1rem 0', fontSize: '1.25rem', fontWeight: 700, color: '#0f172a' }}>
              Promote User to Delivery Agent
            </h3>
            <p style={{ fontSize: '0.85rem', color: '#64748b', marginBottom: '1.25rem' }}>
              Assigns the <code style={{ backgroundColor: '#f1f5f9', padding: '0.1rem 0.3rem', borderRadius: '0.2rem' }}>DELIVERY_AGENT</code> role to a registered user.
            </p>
            <form onSubmit={handleAddAgentRole}>
              <div style={{ marginBottom: '1.5rem' }}>
                <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: 600, color: '#334155', marginBottom: '0.25rem' }}>
                  Select User
                </label>
                <select
                  value={selectedUserId}
                  onChange={e => setSelectedUserId(e.target.value)}
                  required
                  style={{ width: '100%', padding: '0.5rem 0.75rem', border: '1px solid #cbd5e1', borderRadius: '0.375rem', fontSize: '0.875rem', backgroundColor: 'white' }}
                >
                  <option value="">-- Choose User --</option>
                  {allUsers.map(u => (
                    <option key={u.id} value={u.id}>
                      {u.phone} [{u.roles.join(', ')}]
                    </option>
                  ))}
                </select>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
                <button
                  type="button"
                  onClick={() => setShowAddAgentModal(false)}
                  style={{ padding: '0.5rem 1rem', border: '1px solid #cbd5e1', backgroundColor: 'white', borderRadius: '0.375rem', cursor: 'pointer', fontWeight: 500 }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={addAgentLoading}
                  style={{ padding: '0.5rem 1rem', backgroundColor: '#2563eb', color: 'white', border: 'none', borderRadius: '0.375rem', cursor: 'pointer', fontWeight: 600 }}
                >
                  {addAgentLoading ? 'Promoting...' : 'Promote User'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MapLibre Open-Source Map Modal */}
      {selectedAssignmentForMap && (
        <DeliveryMapModal
          assignment={selectedAssignmentForMap}
          onClose={() => setSelectedAssignmentForMap(null)}
        />
      )}
    </div>
  );
};

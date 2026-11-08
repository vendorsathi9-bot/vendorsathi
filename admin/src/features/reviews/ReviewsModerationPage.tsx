import React, { useEffect, useState } from 'react';
import { ApiClient, AdminReview } from '../../core/api/apiClient';
import { useLanguage } from '../../core/i18n/LanguageContext';
import { 
  Star, 
  EyeOff, 
  Eye, 
  RefreshCw, 
  MessageSquare, 
  Filter
} from 'lucide-react';

export const ReviewsModerationPage: React.FC = () => {
  const { t } = useLanguage();
  const [reviews, setReviews] = useState<AdminReview[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'ALL' | 'VISIBLE' | 'HIDDEN'>('ALL');
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);

  const loadReviews = async () => {
    setLoading(true);
    try {
      const data = await ApiClient.listReviews();
      setReviews(data || []);
    } catch (err) {
      console.error('Failed to load reviews', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadReviews();
  }, []);

  const handleModerate = async (reviewId: string, currentHidden: boolean) => {
    const actionText = currentHidden ? 'make this review visible' : 'hide this review from the public platform';
    if (!window.confirm(`Are you sure you want to ${actionText}?`)) return;

    setActionLoadingId(reviewId);
    try {
      await ApiClient.moderateReview(reviewId, !currentHidden);
      await loadReviews();
    } catch (err: any) {
      alert(`Moderation action failed: ${err.message}`);
    } finally {
      setActionLoadingId(null);
    }
  };

  const filteredReviews = reviews.filter(r => {
    if (filter === 'ALL') return true;
    if (filter === 'VISIBLE') return !r.is_hidden;
    if (filter === 'HIDDEN') return r.is_hidden;
    return true;
  });

  const renderStars = (rating: number) => {
    return (
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.15rem' }}>
        {[1, 2, 3, 4, 5].map(star => (
          <Star
            key={star}
            size={14}
            color={star <= rating ? '#eab308' : '#cbd5e1'}
            fill={star <= rating ? '#eab308' : 'none'}
          />
        ))}
        <span style={{ marginLeft: '0.35rem', fontWeight: 600, fontSize: '0.85rem' }}>{rating}/5</span>
      </div>
    );
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 700, color: '#0f172a' }}>{t('reviewTitle')}</h1>
          <p style={{ color: '#64748b', fontSize: '0.9rem', marginTop: '0.25rem' }}>{t('reviewSubtitle')}</p>
        </div>
        <button onClick={loadReviews} className="btn btn-secondary" disabled={loading}>
          <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
          {t('refresh')}
        </button>
      </div>

      {/* Filter Tabs */}
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem', borderBottom: '1px solid #e2e8f0', paddingBottom: '0.75rem' }}>
        {(['ALL', 'VISIBLE', 'HIDDEN'] as const).map(tab => {
          const count = tab === 'ALL' 
            ? reviews.length 
            : tab === 'VISIBLE' 
            ? reviews.filter(r => !r.is_hidden).length 
            : reviews.filter(r => r.is_hidden).length;
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
              {tab === 'ALL' ? t('filterAll') : tab === 'VISIBLE' ? t('visible') : t('hidden')}
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

      {/* Reviews Table */}
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>{t('rating')}</th>
              <th>{t('comment')}</th>
              <th>{t('customer')}</th>
              <th>{t('moderationStatus')}</th>
              <th>{t('date')}</th>
              <th style={{ textAlign: 'right' }}>{t('actions')}</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={6} style={{ textAlign: 'center', padding: '3rem', color: '#64748b' }}>
                  <RefreshCw size={24} className="animate-spin" style={{ margin: '0 auto 0.75rem', display: 'block' }} />
                  {t('loading')}
                </td>
              </tr>
            ) : filteredReviews.length === 0 ? (
              <tr>
                <td colSpan={6} style={{ textAlign: 'center', padding: '3.5rem', color: '#64748b' }}>
                  <MessageSquare size={36} style={{ margin: '0 auto 0.75rem', opacity: 0.4, display: 'block' }} />
                  <div style={{ fontWeight: 600, fontSize: '0.95rem' }}>{t('noReviewsFound')}</div>
                </td>
              </tr>
            ) : (
              filteredReviews.map(r => (
                <tr key={r.id}>
                  <td>{renderStars(r.rating)}</td>
                  <td style={{ maxWidth: '320px' }}>
                    <div style={{ color: '#0f172a', fontWeight: 500, fontStyle: r.comment ? 'normal' : 'italic' }}>
                      {r.comment ? `"${r.comment}"` : 'No written comment'}
                    </div>
                    <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.2rem', fontFamily: 'monospace' }}>
                      Order: #{r.order_id.substring(0, 8)}... | Vendor: #{r.vendor_id.substring(0, 8)}...
                    </div>
                  </td>
                  <td>
                    <div style={{ fontWeight: 600, color: '#0f172a' }}>{r.customer_name || 'Customer'}</div>
                    <div style={{ fontSize: '0.75rem', color: '#64748b', fontFamily: 'monospace' }}>
                      #{r.customer_id.substring(0, 8)}...
                    </div>
                  </td>
                  <td>
                    {r.is_hidden ? (
                      <span className="badge badge-danger" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}>
                        <EyeOff size={12} /> {t('hidden')}
                      </span>
                    ) : (
                      <span className="badge badge-success" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}>
                        <Eye size={12} /> {t('visible')}
                      </span>
                    )}
                  </td>
                  <td style={{ color: '#64748b', fontSize: '0.85rem' }}>
                    {new Date(r.created_at).toLocaleDateString()}
                  </td>
                  <td style={{ textAlign: 'right' }}>
                    <button
                      onClick={() => handleModerate(r.id, r.is_hidden)}
                      disabled={actionLoadingId === r.id}
                      className={r.is_hidden ? 'btn btn-primary' : 'btn btn-danger'}
                      style={{ padding: '0.35rem 0.65rem', fontSize: '0.75rem' }}
                    >
                      {r.is_hidden ? (
                        <>
                          <Eye size={14} /> {t('unhideReview')}
                        </>
                      ) : (
                        <>
                          <EyeOff size={14} /> {t('hideReview')}
                        </>
                      )}
                    </button>
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

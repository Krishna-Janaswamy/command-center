import React from 'react';
import { useNavigate } from 'react-router-dom';

const RequestDetail = ({ request, onClose, onDelete }) => {
  const navigate = useNavigate();

  const handleTest = () => {
    // Navigates to API Tester with state (would need App.jsx support or global store, 
    // but for now just navigates. State passing requires useLocation in APITesterTab)
    navigate('/service-virtualization/api', { state: { request } });
  };

  return (
    <div className="fade-in" style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <div style={{ display: 'flex', gap: '12px', alignItems: 'center', marginBottom: '8px' }}>
            <span className={`badge badge-${request.method === 'GET' ? 'info' : 'success'}`} style={{ fontSize: '1rem' }}>{request.method}</span>
            <h2 style={{ margin: 0, fontSize: '1.2rem', wordBreak: 'break-all' }}>{request.endpoint}</h2>
          </div>
          <div style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>{request.url}</div>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <button 
            className="btn-secondary" 
            title="Test in API Tester"
            style={{ padding: '6px 12px', fontSize: '0.85rem', borderRadius: '6px', background: 'rgba(59, 130, 246, 0.1)', color: '#60a5fa', border: '1px solid rgba(59, 130, 246, 0.2)' }} 
            onClick={handleTest}>
            🧪 Test
          </button>
          <button 
            className="btn-danger" 
            title="Delete Request"
            style={{ padding: '6px 12px', fontSize: '0.85rem', borderRadius: '6px', border: '1px solid rgba(239, 68, 68, 0.2)', background: 'rgba(239, 68, 68, 0.15)', color: '#fca5a5' }} 
            onClick={onDelete}>
            🗑️ Delete
          </button>
          <button 
            className="btn-secondary" 
            title="Close"
            style={{ padding: '6px 12px', fontSize: '0.85rem', borderRadius: '6px', border: '1px solid rgba(0,0,0,0.1)' }} 
            onClick={onClose}>
            ✕ Close
          </button>
        </div>
      </div>

      <div className="grid-4" style={{ borderTop: '1px solid var(--border-color)', borderBottom: '1px solid var(--border-color)', padding: '16px 0' }}>
        <div>
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Status</div>
          <div style={{ fontWeight: 600, color: (request.status >= 200 && request.status < 300) ? 'var(--success)' : 'var(--danger)' }}>
            {request.status}
          </div>
        </div>
        <div>
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Timestamp</div>
          <div>{new Date(request.timestamp && !request.timestamp.endsWith('Z') ? request.timestamp.replace(' ', 'T') + 'Z' : request.timestamp).toLocaleString()}</div>
        </div>
        <div>
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Category</div>
          <div><span className="badge badge-primary">{request.category}</span></div>
        </div>
        <div>
          <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Source</div>
          <div>
            {request.source ? (
              <span className={`badge ${request.source.includes('stub') ? 'badge-primary' : 'badge-warning'}`} style={{ textTransform: 'uppercase' }}>
                {request.source}
              </span>
            ) : (
              <span style={{ color: 'var(--text-muted)' }}>N/A</span>
            )}
          </div>
        </div>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
        <div>
          <h3 style={{ fontSize: '1rem', color: 'var(--text-muted)' }}>Request Headers</h3>
          <pre style={{ maxHeight: '150px' }}>{request.headers}</pre>
        </div>
        
        {request.body && (
          <div>
            <h3 style={{ fontSize: '1rem', color: 'var(--text-muted)' }}>Request Body</h3>
            <pre style={{ maxHeight: '200px' }}>{request.body}</pre>
          </div>
        )}

        <div>
          <h3 style={{ fontSize: '1rem', color: 'var(--text-muted)' }}>Response Headers</h3>
          <pre style={{ maxHeight: '150px' }}>{request.responseHeaders}</pre>
        </div>

        <div>
          <h3 style={{ fontSize: '1rem', color: 'var(--text-muted)' }}>Response Body</h3>
          <pre style={{ maxHeight: '300px' }}>{request.response || 'No Content'}</pre>
        </div>
      </div>
    </div>
  );
};

export default RequestDetail;

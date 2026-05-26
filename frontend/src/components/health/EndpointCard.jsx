import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const EndpointCard = ({ api, status, onCheck }) => {
  const [expanded, setExpanded] = useState(false);
  const navigate = useNavigate();

  const getStatusColor = (s) => {
    if (!s) return 'var(--text-muted)';
    return s.status === 'Up and Stable' ? 'var(--success)' : 'var(--danger)';
  };

  return (
    <div className="glass-panel" style={{ padding: '16px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', cursor: 'pointer' }} onClick={() => setExpanded(!expanded)}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <div style={{ width: '12px', height: '12px', borderRadius: '50%', background: getStatusColor(status) }} />
          <div>
            <div style={{ fontWeight: 600, fontSize: '1.1rem', color: 'var(--text-main)' }}>{api.name}</div>
            <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>{(api.baseUrl || '') + (api.endpoint || '')}</div>
          </div>
          <span className="badge badge-info" style={{ marginLeft: '12px' }}>{api.method}</span>
          <span className="badge badge-primary">{api.category}</span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          {status && (
            <>
              <span style={{ fontSize: '0.85rem', color: getStatusColor(status), fontWeight: 500, marginRight: '4px' }}>{status.status}</span>
              {status.statusCode && <span className="badge" style={{ background: 'rgba(255,255,255,0.1)' }}>{status.statusCode}</span>}
              {status.responseTime && <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>{status.responseTime}ms</span>}
            </>
          )}
          <button className="btn-secondary" style={{ padding: '4px 8px', fontSize: '0.8rem', borderRadius: '4px', background: 'rgba(59, 130, 246, 0.1)', color: '#60a5fa', border: '1px solid rgba(59, 130, 246, 0.2)' }} onClick={(e) => { e.stopPropagation(); navigate('/service-virtualization/api', { state: { request: api } }); }}>Test</button>
          <span style={{ color: 'var(--text-muted)', transform: expanded ? 'rotate(180deg)' : 'none', transition: 'transform 0.2s' }}>▼</span>
        </div>
      </div>
      
      {expanded && (
        <div style={{ marginTop: '16px', paddingTop: '16px', borderTop: '1px solid var(--border-color)', display: 'grid', gap: '12px', fontSize: '0.9rem' }}>
          <div className="grid-2">
            <div>
              <div style={{ color: 'var(--text-muted)', marginBottom: '4px', fontSize: '0.85rem' }}>Description</div>
              <div style={{ marginBottom: '12px' }}>{api.description || 'No description provided.'}</div>
              
              <div style={{ color: 'var(--text-muted)', marginBottom: '4px', fontSize: '0.85rem' }}>Created At</div>
              <div>{new Date(api.createdAt).toLocaleString()}</div>
            </div>
            <div>
              <div style={{ color: 'var(--text-muted)', marginBottom: '4px', fontSize: '0.85rem' }}>Health Check Config</div>
              <div style={{ display: 'flex', gap: '8px', marginBottom: '12px' }}>
                {api.healthCheckHeaders && api.healthCheckHeaders !== '{}' && <span className="badge badge-secondary">Custom Headers</span>}
                {api.healthCheckBody && <span className="badge badge-secondary">Custom Body</span>}
                {api.retryOn500 === 1 && <span className="badge badge-warning">Retry on 500</span>}
                {api.isCustom === 1 && <span className="badge badge-info">Custom Logic</span>}
                {(!api.healthCheckHeaders || api.healthCheckHeaders === '{}') && !api.healthCheckBody && !api.retryOn500 && !api.isCustom && <span style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>Standard Config</span>}
              </div>

              <div style={{ color: 'var(--text-muted)', marginBottom: '4px', fontSize: '0.85rem' }}>Environment</div>
              <div>{api.environment}</div>
            </div>
          </div>
          {status && status.error && (
            <div style={{ background: 'rgba(239, 68, 68, 0.1)', border: '1px solid var(--danger)', padding: '12px', borderRadius: '6px', color: '#f87171' }}>
              <strong>Error:</strong> {status.error}
            </div>
          )}
          {status && status.lastChecked && (
            <div style={{ color: 'var(--text-muted)', fontSize: '0.8rem', textAlign: 'right' }}>
              Last checked: {status.lastChecked.toLocaleString()}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default EndpointCard;

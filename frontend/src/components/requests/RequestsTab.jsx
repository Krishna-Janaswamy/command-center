import React, { useState, useEffect } from 'react';
import { requestApi } from '../../services/registryApi';
import RequestDetail from './RequestDetail';

const RequestsTab = () => {
  const [requests, setRequests] = useState([]);
  const [selectedRequest, setSelectedRequest] = useState(null);
  const [filter, setFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('All');

  const loadRequests = async () => {
    try {
      const res = await requestApi.getAll();
      setRequests(res.data);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    loadRequests();
    const interval = setInterval(loadRequests, 10000); // 10s auto-refresh
    return () => clearInterval(interval);
  }, []);

  const handleClear = async () => {
    if (window.confirm('Delete all recorded requests?')) {
      try {
        await requestApi.deleteAll();
        loadRequests();
      } catch (err) {
        console.error(err);
      }
    }
  };

  const filtered = requests.filter(r => {
    if (categoryFilter !== 'All' && r.category !== categoryFilter) return false;
    if (filter) {
      const f = filter.toLowerCase();
      return r.url.toLowerCase().includes(f) || 
             r.method.toLowerCase().includes(f) || 
             (r.baseUrl && r.baseUrl.toLowerCase().includes(f));
    }
    return true;
  });

  return (
    <div className="fade-in" style={{ display: 'flex', gap: '24px', height: 'calc(100vh - 120px)' }}>
      {/* Left Panel */}
      <div className="glass-panel" style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
          <h2>Requests ({filtered.length})</h2>
          <div style={{ display: 'flex', gap: '8px' }}>
            <button className="btn-secondary" style={{ padding: '6px 12px' }} onClick={loadRequests}>↻</button>
            <button className="btn-danger" style={{ padding: '6px 12px' }} onClick={handleClear}>Clear</button>
          </div>
        </div>
        <div style={{ display: 'flex', gap: '8px', marginBottom: '16px' }}>
          <input className="form-control" placeholder="Search URL..." value={filter} onChange={e => setFilter(e.target.value)} />
          <select className="form-control" style={{ width: '120px' }} value={categoryFilter} onChange={e => setCategoryFilter(e.target.value)}>
            <option>All</option><option>Claims</option><option>SBI</option><option>GW</option><option>Other</option>
          </select>
        </div>
        
        <div style={{ flex: 1, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {filtered.map(req => (
            <div 
              key={req.id} 
              onClick={() => setSelectedRequest(req)}
              style={{
                padding: '12px', background: selectedRequest?.id === req.id ? 'rgba(0,0,0,0.05)' : 'transparent',
                border: '1px solid var(--border-color)', borderRadius: '8px', cursor: 'pointer',
                borderColor: selectedRequest?.id === req.id ? 'var(--primary)' : 'var(--border-color)'
              }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                  <span className={`badge badge-${req.method === 'GET' ? 'info' : 'success'}`}>{req.method}</span>
                  <span className={`badge badge-${(req.status >= 200 && req.status < 300) ? 'success' : 'danger'}`}>{req.status}</span>
                  {req.source && (
                    <span className={`badge ${req.source.includes('stub') ? 'badge-primary' : 'badge-warning'}`} style={{ textTransform: 'uppercase', fontSize: '0.65rem' }}>
                      {req.source}
                    </span>
                  )}
                </div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                  {new Date(req.timestamp && !req.timestamp.endsWith('Z') ? req.timestamp.replace(' ', 'T') + 'Z' : req.timestamp).toLocaleTimeString()}
                </div>
              </div>
              <div style={{ fontSize: '0.85rem', color: 'var(--text-main)', wordBreak: 'break-all', marginBottom: '4px' }}>
                {req.endpoint}
              </div>
              <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>🌐 {req.baseUrl || 'N/A'}</div>
            </div>
          ))}
        </div>
      </div>

      {/* Right Panel */}
      <div className="glass-panel" style={{ flex: 1.5, display: 'flex', flexDirection: 'column', overflowY: 'auto' }}>
        {selectedRequest ? (
          <RequestDetail request={selectedRequest} onClose={() => setSelectedRequest(null)} onDelete={async () => {
            if(window.confirm('Delete this request?')) {
              await requestApi.delete(selectedRequest.id);
              setSelectedRequest(null);
              loadRequests();
            }
          }} />
        ) : (
          <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-muted)' }}>
            Select a request to view details.
          </div>
        )}
      </div>
    </div>
  );
};

export default RequestsTab;

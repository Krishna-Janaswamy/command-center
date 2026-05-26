import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { registryApi } from '../../services/registryApi';
import ApiFormModal from './ApiFormModal';

const ApiRegistry = () => {
  const [apis, setApis] = useState([]);
  const [showModal, setShowModal] = useState(false);
  const [editApi, setEditApi] = useState(null);
  const [expandedApi, setExpandedApi] = useState(null);
  const [responses, setResponses] = useState({});
  const [loadingResponse, setLoadingResponse] = useState({});
  const navigate = useNavigate();

  const toggleExpand = async (id) => {
    setExpandedApi(prev => prev === id ? null : id);
    if (expandedApi !== id && !responses[id]) {
       const api = apis.find(a => a.id === id);
       if(api) {
           setLoadingResponse(prev => ({ ...prev, [id]: true }));
           try {
               const { healthApi } = await import('../../services/registryApi');
               const res = await healthApi.check({
                  url: (api.baseUrl || '') + api.endpoint,
                  method: api.method,
                  headers: api.healthCheckHeaders && api.healthCheckHeaders !== '{}' ? JSON.parse(api.healthCheckHeaders) : {},
                  body: api.healthCheckBody
               });
               setResponses(prev => ({ ...prev, [id]: res.data.body }));
           } catch(err) {
               setResponses(prev => ({ ...prev, [id]: 'Error fetching response' }));
           } finally {
               setLoadingResponse(prev => ({ ...prev, [id]: false }));
           }
       }
    }
  };

  const loadApis = async () => {
    try {
      const res = await registryApi.getAll();
      setApis(res.data);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    loadApis();
  }, []);

  const handleDelete = async (id) => {
    if (window.confirm("Are you sure you want to delete this API?")) {
      try {
        await registryApi.delete(id);
        loadApis();
      } catch (err) {
        console.error("Delete failed", err);
      }
    }
  };

  return (
    <div className="fade-in">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <h1>API Registry</h1>
        <button className="btn" onClick={() => { setEditApi(null); setShowModal(true); }}>
          + Register API
        </button>
      </div>

      <div className="glass-panel">
        <table>
          <thead>
            <tr>
              <th style={{ width: '40px' }}></th>
              <th>Method</th>
              <th>Name</th>

              <th>Base URL</th>
              <th>Endpoint</th>
              <th>Category</th>
              <th>Environment</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {apis.map(api => (
              <React.Fragment key={api.id}>
                <tr style={{ cursor: 'pointer', background: expandedApi === api.id ? 'rgba(255,255,255,0.02)' : 'transparent' }} onClick={() => toggleExpand(api.id)}>
                  <td style={{ color: 'var(--text-muted)' }}>
                    {expandedApi === api.id ? '▼' : '▶'}
                  </td>
                  <td>
                    <span className={`badge badge-${api.method === 'GET' ? 'info' : api.method === 'POST' ? 'success' : 'warning'}`}>
                      {api.method}
                    </span>
                  </td>
                  <td style={{ fontWeight: 500 }}>{api.name}</td>

                  <td>{api.baseUrl}</td>
                  <td>{api.endpoint}</td>
                  <td><span className="badge badge-primary">{api.category}</span></td>
                  <td>{api.environment}</td>
                  <td onClick={(e) => e.stopPropagation()}>
                    <div style={{ display: 'flex', gap: '8px' }}>
                      <button 
                        className="btn-secondary" 
                        title="Test in API Tester"
                        style={{ padding: '6px 10px', fontSize: '0.85rem', borderRadius: '6px', background: 'rgba(59, 130, 246, 0.1)', color: '#60a5fa', border: '1px solid rgba(59, 130, 246, 0.2)' }} 
                        onClick={() => navigate('/service-virtualization/api', { state: { request: api } })}>
                        🧪 Test
                      </button>
                      <button 
                        className="btn-secondary" 
                        title="Edit API"
                        style={{ padding: '6px 10px', fontSize: '0.85rem', borderRadius: '6px', border: '1px solid rgba(255,255,255,0.1)' }} 
                        onClick={() => { setEditApi(api); setShowModal(true); }}>
                        ✏️ Edit
                      </button>
                      <button 
                        className="btn-danger" 
                        title="Delete API"
                        style={{ padding: '6px 10px', fontSize: '0.85rem', borderRadius: '6px', border: '1px solid rgba(239, 68, 68, 0.2)', background: 'rgba(239, 68, 68, 0.15)', color: '#fca5a5' }} 
                        onClick={() => handleDelete(api.id)}>
                        🗑️ Delete
                      </button>
                    </div>
                  </td>
                </tr>
                {expandedApi === api.id && (
                  <tr style={{ background: 'rgba(0,0,0,0.2)' }}>
                    <td colSpan="8" style={{ padding: '24px' }}>
                      <div style={{ display: 'flex', gap: '32px', marginBottom: '16px' }}>
                        <div>
                          <strong style={{ color: 'var(--text-muted)', fontSize: '0.85rem', display: 'block', marginBottom: '4px' }}>Created At</strong>
                          <div style={{ fontSize: '0.9rem' }}>{new Date(api.createdAt).toLocaleString()}</div>
                        </div>
                        <div>
                          <strong style={{ color: 'var(--text-muted)', fontSize: '0.85rem', display: 'block', marginBottom: '4px' }}>Environment</strong>
                          <div style={{ fontSize: '0.9rem' }}>{api.environment}</div>
                        </div>
                      </div>
                      
                      <div>
                        <strong style={{ color: 'var(--text-muted)', fontSize: '0.85rem', display: 'block', marginBottom: '4px' }}>Response Payload</strong>
                        {loadingResponse[api.id] ? (
                          <div style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>Loading response...</div>
                        ) : (
                          <pre style={{ maxHeight: '200px', overflowY: 'auto', overflowX: 'auto', whiteSpace: 'pre-wrap', wordBreak: 'break-word', background: 'rgba(0,0,0,0.2)', padding: '12px', borderRadius: '6px' }}>
                            {(() => {
                              try {
                                return JSON.stringify(JSON.parse(responses[api.id]), null, 2);
                              } catch(e) {
                                return responses[api.id] || 'No response data.';
                              }
                            })()}
                          </pre>
                        )}
                      </div>
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}
            {apis.length === 0 && (
              <tr>
                <td colSpan="8" style={{ textAlign: 'center', color: 'var(--text-muted)' }}>No APIs registered yet.</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {showModal && (
        <ApiFormModal 
          api={editApi} 
          onClose={() => setShowModal(false)} 
          onSave={() => { setShowModal(false); loadApis(); }} 
        />
      )}
    </div>
  );
};

export default ApiRegistry;

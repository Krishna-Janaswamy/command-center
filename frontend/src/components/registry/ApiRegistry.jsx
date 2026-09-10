import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { registryApi } from '../../services/registryApi';
import { parseJsonObject } from '../../utils/parseJsonObject';
import ApiFormModal from './ApiFormModal';

const ApiRegistry = () => {
  const [apis, setApis] = useState([]);
  const [showModal, setShowModal] = useState(false);
  const [editApi, setEditApi] = useState(null);
  const [expandedApi, setExpandedApi] = useState(null);
  const [responses, setResponses] = useState({});
  const [loadingResponse, setLoadingResponse] = useState({});
  const [filter, setFilter] = useState('');
  const navigate = useNavigate();

  const toggleExpand = async (id) => {
    setExpandedApi(prev => (prev === id ? null : id));
    if (expandedApi !== id && !responses[id]) {
      const api = apis.find(a => a.id === id);
      if (!api) return;
      setLoadingResponse(prev => ({ ...prev, [id]: true }));
      try {
        const { healthApi } = await import('../../services/registryApi');
        const res = await healthApi.check({
          url: (api.baseUrl || '') + api.endpoint,
          method: api.method,
          headers: parseJsonObject(api.healthCheckHeaders),
          body: api.healthCheckBody
        });
        setResponses(prev => ({ ...prev, [id]: res.data.body }));
      } catch (err) {
        setResponses(prev => ({ ...prev, [id]: 'Error fetching response' }));
      } finally {
        setLoadingResponse(prev => ({ ...prev, [id]: false }));
      }
    }
  };

  const loadApis = async () => {
    try {
      const res = await registryApi.getAll();
      setApis(Array.isArray(res.data) ? res.data : []);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    loadApis();
  }, []);

  const handleDelete = async (id) => {
    if (!window.confirm('Are you sure you want to delete this API?')) return;
    try {
      await registryApi.delete(id);
      loadApis();
    } catch (err) {
      console.error('Delete failed', err);
    }
  };

  const filteredApis = apis.filter(api => {
    if (!filter) return true;
    const f = filter.toLowerCase();
    return (api.name && api.name.toLowerCase().includes(f)) ||
      (api.endpoint && api.endpoint.toLowerCase().includes(f)) ||
      (api.baseUrl && api.baseUrl.toLowerCase().includes(f));
  });

  const formatResponse = (value) => {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch (e) {
      return value || 'No response data.';
    }
  };

  return (
    <div className="fade-in registry-page">
      <div className="page-toolbar">
        <h1>API Registry</h1>
        <div className="page-toolbar-actions">
          <input
            className="form-control page-toolbar-search"
            placeholder="Search APIs..."
            value={filter}
            onChange={e => setFilter(e.target.value)}
          />
          <button type="button" className="btn" onClick={() => { setEditApi(null); setShowModal(true); }}>
            + Register API
          </button>
        </div>
      </div>

      <div className="registry-list">
        {filteredApis.map(api => {
          const open = expandedApi === api.id;
          const fullUrl = `${api.baseUrl || ''}${api.endpoint || ''}` || 'No URL configured';
          return (
            <article key={api.id} className={`glass-panel registry-item${open ? ' is-expanded' : ''}`}>
              <div className="registry-row">
                <button type="button" className="registry-row-main" onClick={() => toggleExpand(api.id)}>
                  <span className="registry-chevron" aria-hidden="true">{open ? '▼' : '▶'}</span>
                  <span className={`badge badge-${api.method === 'GET' ? 'info' : api.method === 'POST' ? 'success' : 'warning'}`}>
                    {api.method || 'ANY'}
                  </span>
                  <span className="registry-item-name" title={api.name}>{api.name || 'Untitled API'}</span>
                  <span className="registry-item-url" title={fullUrl}>{fullUrl}</span>
                  {api.category && <span className="badge badge-primary registry-col-category">{api.category}</span>}
                  {api.environment && <span className="registry-env registry-col-env">{api.environment}</span>}
                </button>

                <div className="registry-item-actions">
                  <button
                    type="button"
                    className="btn-secondary registry-action-btn registry-action-test"
                    onClick={() => navigate('/service-virtualization/api', { state: { request: api } })}
                  >
                    Test
                  </button>
                  <button
                    type="button"
                    className="btn-secondary registry-action-btn"
                    onClick={() => { setEditApi(api); setShowModal(true); }}
                  >
                    Edit
                  </button>
                  <button
                    type="button"
                    className="btn-danger registry-action-btn registry-action-delete"
                    onClick={() => handleDelete(api.id)}
                  >
                    Delete
                  </button>
                </div>
              </div>

              {open && (
                <div className="registry-expanded">
                  <div className="registry-expanded-meta">
                    <div>
                      <strong>Created At</strong>
                      <div>{api.createdAt ? new Date(api.createdAt).toLocaleString() : '—'}</div>
                    </div>
                    <div>
                      <strong>Base URL</strong>
                      <div className="registry-wrap">{api.baseUrl || '—'}</div>
                    </div>
                    <div>
                      <strong>Endpoint</strong>
                      <div className="registry-wrap">{api.endpoint || '—'}</div>
                    </div>
                  </div>

                  <div>
                    <strong className="registry-response-label">Response Payload</strong>
                    {loadingResponse[api.id] ? (
                      <div className="registry-muted">Loading response...</div>
                    ) : (
                      <pre className="registry-response-pre">{formatResponse(responses[api.id])}</pre>
                    )}
                  </div>
                </div>
              )}
            </article>
          );
        })}

        {filteredApis.length === 0 && (
          <div className="glass-panel registry-empty">No APIs found matching your filters.</div>
        )}
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

import React, { useState, useEffect } from 'react';
import { registryApi, healthApi } from '../../services/registryApi';
import { parseJsonObject } from '../../utils/parseJsonObject';
import EndpointCard from './EndpointCard';

const ApiHealth = () => {
  const [apis, setApis] = useState([]);
  const [healthStatus, setHealthStatus] = useState({});
  const [filter, setFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('All');
  const [lastRefreshed, setLastRefreshed] = useState(null);

  useEffect(() => {
    registryApi.getAll().then(res => {
      setApis(res.data);
      checkAll(res.data);
    });
  }, []);

  const checkAll = async (apiList) => {
    const promises = apiList
      .filter(api => api.baseUrl || api.endpoint)
      .map(api => checkHealth(api));
    await Promise.all(promises);
    setLastRefreshed(new Date());
  };

  const checkHealth = async (api) => {
    try {
      const start = Date.now();
      const res = await healthApi.check({
        url: (api.baseUrl || '') + (api.endpoint && api.endpoint !== '/' ? api.endpoint : ''),
        method: api.method,
        headers: parseJsonObject(api.healthCheckHeaders),
        body: api.healthCheckBody,
        params: parseJsonObject(api.healthCheckParams)
      });
      const time = Date.now() - start;
      
      setHealthStatus(prev => ({
        ...prev,
        [api.id]: {
          status: (res.data.statusCode >= 200 && res.data.statusCode < 300) ? 'Up and Stable' : 'DOWN',
          statusCode: res.data.statusCode,
          responseTime: res.data.responseTime || time,
          body: res.data.body,
          error: res.data.error,
          lastChecked: new Date()
        }
      }));
    } catch (err) {
      setHealthStatus(prev => ({
        ...prev,
        [api.id]: { status: 'DOWN', error: err.message, lastChecked: new Date() }
      }));
    }
  };

  const filteredApis = apis.filter(api => {
    if (categoryFilter !== 'All' && api.category !== categoryFilter) return false;
    if (filter) {
      const f = filter.toLowerCase();
      return (api.name && api.name.toLowerCase().includes(f)) || 
             (api.endpoint && api.endpoint.toLowerCase().includes(f)) ||
             (api.baseUrl && api.baseUrl.toLowerCase().includes(f));
    }
    return true;
  });

  return (
    <div className="fade-in">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <div>
          <h1 style={{ margin: 0 }}>API Health Dashboard</h1>
          {lastRefreshed && (
            <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '6px', fontWeight: 500 }}>
              Last refreshed: {lastRefreshed.toLocaleTimeString()}
            </div>
          )}
        </div>
        <button className="btn" onClick={() => checkAll(apis)}>Refresh All</button>
      </div>

      <div className="glass-panel" style={{ padding: '16px', marginBottom: '24px', display: 'flex', gap: '16px' }}>
        <input 
          className="form-control" 
          placeholder="Search by API name or endpoint..." 
          style={{ flex: 1 }}
          value={filter} 
          onChange={e => setFilter(e.target.value)} 
        />
        <select 
          className="form-control" 
          style={{ width: '200px' }} 
          value={categoryFilter} 
          onChange={e => setCategoryFilter(e.target.value)}
        >
          <option>All</option><option>Claims</option><option>SBI</option><option>GW</option><option>Other</option>
        </select>
      </div>

      <div style={{ display: 'grid', gap: '16px' }}>
        {filteredApis.map(api => (
          <EndpointCard key={api.id} api={api} status={healthStatus[api.id]} onCheck={() => checkHealth(api)} />
        ))}
        {filteredApis.length === 0 && <div className="glass-panel" style={{ textAlign: 'center', color: 'var(--text-muted)' }}>No APIs found matching your filters.</div>}
      </div>
    </div>
  );
};

export default ApiHealth;

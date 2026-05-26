import React, { useState, useEffect, useMemo } from 'react';
import { requestApi } from '../../services/registryApi';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
} from 'chart.js';
import { Line, Pie } from 'react-chartjs-2';

ChartJS.register(
  CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, ArcElement
);

const AnalyticsTab = () => {
  const [requests, setRequests] = useState([]);
  const [selectedBaseUrl, setSelectedBaseUrl] = useState('All');
  const [searchQuery, setSearchQuery] = useState('');
  
  useEffect(() => {
    requestApi.getAll().then(res => setRequests(res.data)).catch(console.error);
  }, []);

  const stats = useMemo(() => {
    if (requests.length === 0) return { total: 0, success: 0, error: 0, avgTime: 0, endpoints: [] };

    let success = 0;
    let error = 0;
    const endpointMap = {};

    requests.forEach(r => {
      if (r.status >= 200 && r.status < 400) success++;
      else error++;
    });

    return {
      total: requests.length,
      success,
      error,
      avgTime: 45 // Placeholder
    };
  }, [requests]);

  const baseUrls = useMemo(() => {
    const urls = new Set();
    requests.forEach(r => {
      if (r.baseUrl) urls.add(r.baseUrl.replace(/^https?:\/\//, ''));
    });
    return ['All', ...Array.from(urls)];
  }, [requests]);

  const filteredEndpointsData = useMemo(() => {
    const endpointMap = {};
    let filteredTotal = 0;

    requests.forEach(r => {
      const bUrl = r.baseUrl ? r.baseUrl.replace(/^https?:\/\//, '') : '';
      if (selectedBaseUrl !== 'All' && bUrl !== selectedBaseUrl) return;
      
      const displayName = selectedBaseUrl === 'All' && bUrl ? `${bUrl}${r.endpoint}` : r.endpoint;
      
      if (searchQuery && !displayName.toLowerCase().includes(searchQuery.toLowerCase())) return;

      filteredTotal++;
      if (!endpointMap[displayName]) {
        endpointMap[displayName] = 0;
      }
      endpointMap[displayName]++;
    });

    const endpoints = Object.entries(endpointMap)
      .sort((a, b) => b[1] - a[1])
      .map(([name, count]) => ({ name, count }))
      .slice(0, 20); // show top 20 or more if searching

    return { endpoints, total: filteredTotal || 1 }; // prevent division by zero
  }, [requests, selectedBaseUrl, searchQuery]);


  const pieData = {
    labels: ['Success (2xx/3xx)', 'Error (4xx/5xx)'],
    datasets: [
      {
        data: [stats.success, stats.error],
        backgroundColor: ['rgba(16, 185, 129, 0.6)', 'rgba(239, 68, 68, 0.6)'],
        borderColor: ['rgba(16, 185, 129, 1)', 'rgba(239, 68, 68, 1)'],
        borderWidth: 1,
      },
    ],
  };

  const lineData = useMemo(() => {
    const days = [];
    const dateStrings = [];
    for (let i = 6; i >= 0; i--) {
      const d = new Date();
      d.setDate(d.getDate() - i);
      days.push(d.toLocaleDateString('en-US', { weekday: 'short' }));
      dateStrings.push(d.toLocaleDateString());
    }

    const counts = [0, 0, 0, 0, 0, 0, 0];
    requests.forEach(r => {
      let ts = r.timestamp;
      if (ts && !ts.endsWith('Z') && ts.includes(' ')) ts = ts.replace(' ', 'T') + 'Z';
      else if (ts && !ts.endsWith('Z')) ts += 'Z';
      
      const rDate = new Date(ts);
      if (!isNaN(rDate.getTime())) {
        const localDateString = rDate.toLocaleDateString();
        const idx = dateStrings.indexOf(localDateString);
        if (idx !== -1) counts[idx]++;
      }
    });

    return {
      labels: days,
      datasets: [
        {
          label: 'Requests per day',
          data: counts,
          borderColor: 'rgba(139, 92, 246, 1)',
          backgroundColor: 'rgba(139, 92, 246, 0.2)',
          fill: true,
          tension: 0.4
        }
      ]
    };
  }, [requests]);

  return (
    <div className="fade-in">
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '24px' }}>
        <div style={{ background: 'linear-gradient(135deg, var(--primary), #a78bfa)', width: '40px', height: '40px', borderRadius: '10px', display: 'flex', alignItems: 'center', justifyContent: 'center', boxShadow: '0 0 15px rgba(139, 92, 246, 0.4)' }}>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M3 3v18h18"/><path d="m19 9-5 5-4-4-3 3"/></svg>
        </div>
        <h1 style={{ margin: 0, fontWeight: 700, letterSpacing: '-0.5px' }} className="text-gradient gradient-primary">Analytics Dashboard</h1>
      </div>
      
      <div className="grid-3" style={{ marginBottom: '32px' }}>
        <div className="glass-panel glass-panel-hover">
          <div style={{ color: 'var(--text-muted)', textTransform: 'uppercase', fontSize: '0.8rem', letterSpacing: '1px', marginBottom: '8px', fontWeight: 600 }}>Total Requests</div>
          <div style={{ fontSize: '3rem', fontWeight: 700 }} className="text-gradient gradient-primary">{stats.total}</div>
        </div>
        <div className="glass-panel glass-panel-hover">
          <div style={{ color: 'var(--text-muted)', textTransform: 'uppercase', fontSize: '0.8rem', letterSpacing: '1px', marginBottom: '8px', fontWeight: 600 }}>Success Rate</div>
          <div style={{ fontSize: '3rem', fontWeight: 700 }} className="text-gradient gradient-success">
            {stats.total ? Math.round((stats.success / stats.total) * 100) : 0}%
          </div>
        </div>
        <div className="glass-panel glass-panel-hover">
          <div style={{ color: 'var(--text-muted)', textTransform: 'uppercase', fontSize: '0.8rem', letterSpacing: '1px', marginBottom: '8px', fontWeight: 600 }}>Avg Response Time</div>
          <div style={{ fontSize: '3rem', fontWeight: 700 }} className="text-gradient gradient-info">{stats.avgTime}ms</div>
        </div>
      </div>

      <div className="grid-2" style={{ marginBottom: '32px' }}>
        <div className="glass-panel" style={{ height: '350px' }}>
          <h3>Request Volume</h3>
          <div style={{ height: '280px' }}>
            <Line data={lineData} options={{ maintainAspectRatio: false, color: '#94a3b8', scales: { x: { ticks: { color: '#94a3b8' } }, y: { ticks: { color: '#94a3b8' } } } }} />
          </div>
        </div>
        
        <div className="glass-panel" style={{ height: '350px' }}>
          <h3>Success vs Error</h3>
          <div style={{ height: '280px', display: 'flex', justifyContent: 'center' }}>
            <Pie data={pieData} options={{ maintainAspectRatio: false, color: '#94a3b8' }} />
          </div>
        </div>
      </div>

      <div className="glass-panel" style={{ maxHeight: '400px', overflowY: 'auto' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px', position: 'sticky', top: 0, zIndex: 10, background: 'var(--panel-bg)', paddingBottom: '12px' }}>
          <h3 style={{ margin: 0 }}>Top Endpoints</h3>
          <div style={{ display: 'flex', gap: '12px' }}>
            <input 
              type="text" 
              placeholder="Search endpoints..." 
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
              style={{
                background: 'rgba(255,255,255,0.5)', 
                border: '1px solid var(--border-color)', 
                color: 'var(--text-main)', 
                padding: '6px 12px', 
                borderRadius: '6px',
                fontSize: '0.85rem',
                outline: 'none',
                minWidth: '200px'
              }}
            />
            <select 
              value={selectedBaseUrl} 
              onChange={e => setSelectedBaseUrl(e.target.value)}
              style={{ 
                background: 'rgba(255,255,255,0.5)', 
                border: '1px solid var(--border-color)', 
                color: 'var(--text-main)', 
                padding: '6px 12px', 
                borderRadius: '6px',
                fontSize: '0.85rem',
                outline: 'none',
                minWidth: '150px'
              }}
            >
              {baseUrls.map(url => (
                <option key={url} value={url}>{url}</option>
              ))}
            </select>
          </div>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          {filteredEndpointsData.endpoints.map(ep => (
            <div key={ep.name} className="endpoint-item" style={{ background: 'rgba(0,0,0,0.02)', padding: '12px 16px', borderRadius: '8px', border: '1px solid rgba(0,0,0,0.05)' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '12px', fontSize: '0.95rem' }}>
                <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '80%', fontWeight: 500, color: 'var(--text-main)' }} title={ep.name}>{ep.name}</span>
                <span style={{ background: 'rgba(0,0,0,0.05)', color: '#8b5cf6', padding: '2px 10px', borderRadius: '12px', fontSize: '0.8rem', fontWeight: 600 }}>{ep.count} hits</span>
              </div>
              <div className="progress-bar-bg" style={{ height: '6px' }}>
                <div className="progress-bar-fill" style={{ width: `${(ep.count / filteredEndpointsData.total) * 100}%` }} />
              </div>
            </div>
          ))}
          {filteredEndpointsData.endpoints.length === 0 && (
            <div style={{ color: 'var(--text-muted)', textAlign: 'center', marginTop: '40px', fontStyle: 'italic', gridColumn: '1 / -1' }}>
              No requests recorded yet.
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default AnalyticsTab;

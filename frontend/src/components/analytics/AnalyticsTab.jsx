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

      if (!endpointMap[r.endpoint]) {
        endpointMap[r.endpoint] = 0;
      }
      endpointMap[r.endpoint]++;
    });

    const endpoints = Object.entries(endpointMap)
      .sort((a, b) => b[1] - a[1])
      .map(([name, count]) => ({ name, count }));

    return {
      total: requests.length,
      success,
      error,
      avgTime: 45, // Placeholder, would need responseTime in requests table to be real
      endpoints: endpoints.slice(0, 5)
    };
  }, [requests]);

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

  const lineData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      {
        label: 'Requests per day',
        data: [12, 19, 3, 5, 2, 3, stats.total], // Mocked time series for visual
        borderColor: 'rgba(139, 92, 246, 1)',
        backgroundColor: 'rgba(139, 92, 246, 0.2)',
        fill: true,
        tension: 0.4
      }
    ]
  };

  return (
    <div className="fade-in">
      <h1>Traffic Analytics</h1>
      
      <div className="grid-3" style={{ marginBottom: '24px' }}>
        <div className="glass-panel">
          <div style={{ color: 'var(--text-muted)' }}>Total Requests</div>
          <div style={{ fontSize: '2.5rem', fontWeight: 600, color: 'var(--primary)' }}>{stats.total}</div>
        </div>
        <div className="glass-panel">
          <div style={{ color: 'var(--text-muted)' }}>Success Rate</div>
          <div style={{ fontSize: '2.5rem', fontWeight: 600, color: 'var(--success)' }}>
            {stats.total ? Math.round((stats.success / stats.total) * 100) : 0}%
          </div>
        </div>
        <div className="glass-panel">
          <div style={{ color: 'var(--text-muted)' }}>Avg Response Time</div>
          <div style={{ fontSize: '2.5rem', fontWeight: 600, color: 'var(--info)' }}>{stats.avgTime}ms</div>
        </div>
      </div>

      <div className="grid-2">
        <div className="glass-panel" style={{ height: '350px' }}>
          <h3>Request Volume</h3>
          <div style={{ height: '280px' }}>
            <Line data={lineData} options={{ maintainAspectRatio: false, color: '#94a3b8', scales: { x: { ticks: { color: '#94a3b8' } }, y: { ticks: { color: '#94a3b8' } } } }} />
          </div>
        </div>
        
        <div className="grid-2">
          <div className="glass-panel" style={{ height: '350px' }}>
            <h3>Success vs Error</h3>
            <div style={{ height: '280px', display: 'flex', justifyContent: 'center' }}>
              <Pie data={pieData} options={{ maintainAspectRatio: false, color: '#94a3b8' }} />
            </div>
          </div>
          
          <div className="glass-panel" style={{ height: '350px', overflowY: 'auto' }}>
            <h3>Top Endpoints</h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px', marginTop: '16px' }}>
              {stats.endpoints.map(ep => (
                <div key={ep.name}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px', fontSize: '0.9rem' }}>
                    <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '70%' }}>{ep.name}</span>
                    <span style={{ color: 'var(--text-muted)' }}>{ep.count}</span>
                  </div>
                  <div style={{ width: '100%', height: '6px', background: 'rgba(255,255,255,0.1)', borderRadius: '3px', overflow: 'hidden' }}>
                    <div style={{ width: `${(ep.count / stats.total) * 100}%`, height: '100%', background: 'var(--primary)' }} />
                  </div>
                </div>
              ))}
              {stats.endpoints.length === 0 && <div style={{ color: 'var(--text-muted)' }}>No data available</div>}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AnalyticsTab;

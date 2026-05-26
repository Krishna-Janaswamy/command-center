import React, { useState, useEffect } from 'react';
import { NavLink, Link } from 'react-router-dom';
import { settingsApi } from '../services/registryApi';

const Sidebar = ({ user }) => {
  const [healthOpen, setHealthOpen] = useState(true);
  const [virtOpen, setVirtOpen] = useState(true);
  const [toggle, setToggle] = useState(true);

  useEffect(() => {
    settingsApi.getAll().then(res => {
      setToggle(res.data.useToggle === 'true' || res.data.useToggle === 'on');
    }).catch(console.error);
  }, []);

  const handleToggle = async () => {
    const newValue = !toggle;
    try {
      await settingsApi.update('useToggle', newValue ? 'on' : 'off');
      setToggle(newValue);
    } catch (err) {
      console.error(err);
    }
  };
  return (
    <aside className="sidebar">
      <Link to="/" style={{ textDecoration: 'none', color: 'inherit' }}>
        <div className="brand" style={{ cursor: 'pointer' }}>
          <span style={{ fontSize: '1.5rem' }}>⬢</span>
          Command Center
        </div>
      </Link>

      <div 
        onClick={() => setHealthOpen(!healthOpen)}
        style={{ padding: '0 16px', margin: '16px 0 8px', color: 'var(--text-muted)', fontSize: '0.75rem', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em', cursor: 'pointer', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
      >
        <span>🩺 Health Monitoring</span>
        <span>{healthOpen ? '▼' : '▶'}</span>
      </div>
      {healthOpen && (
        <nav>
          <NavLink to="/api-registry" className={({isActive}) => isActive ? "active" : ""}>📚 API Registry</NavLink>
          <NavLink to="/api-health" className={({isActive}) => isActive ? "active" : ""}>💓 API Health</NavLink>
        </nav>
      )}

      <div 
        onClick={() => setVirtOpen(!virtOpen)}
        style={{ padding: '0 16px', margin: '24px 0 8px', color: 'var(--text-muted)', fontSize: '0.75rem', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em', cursor: 'pointer', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
      >
        <span>🧪 Service Virtualization</span>
        <span>{virtOpen ? '▼' : '▶'}</span>
      </div>
      {virtOpen && (
        <>
          <div style={{ marginBottom: '8px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '8px', background: 'rgba(0,0,0,0.2)', padding: '12px 16px', borderBottom: '1px solid var(--border-color)', borderTop: '1px solid var(--border-color)' }}>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>Live API:</span>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <button 
                onClick={handleToggle}
                style={{
                  width: '40px', height: '22px', borderRadius: '11px', border: 'none',
                  background: toggle ? 'var(--success)' : 'var(--text-muted)',
                  position: 'relative', cursor: 'pointer', transition: 'background 0.2s'
                }}
              >
                <div style={{
                  width: '18px', height: '18px', borderRadius: '50%', background: 'white',
                  position: 'absolute', top: '2px', left: toggle ? '20px' : '2px',
                  transition: 'left 0.2s', boxShadow: '0 1px 3px rgba(0,0,0,0.3)'
                }} />
              </button>
              <span style={{ fontSize: '0.8rem', fontWeight: 600, color: toggle ? 'var(--success)' : 'var(--text-muted)' }}>
                {toggle ? 'LIVE' : 'STUBBED'}
              </span>
            </div>
          </div>
          <nav>
            <NavLink to="/service-virtualization/api" className={({isActive}) => isActive ? "active" : ""}>🔌 API Request</NavLink>
            <NavLink to="/service-virtualization/stubs" className={({isActive}) => isActive ? "active" : ""}>🎭 Stubs</NavLink>
            <NavLink to="/service-virtualization/requests" className={({isActive}) => isActive ? "active" : ""}>📋 Recorded Requests</NavLink>
            <NavLink to="/service-virtualization/analytics" className={({isActive}) => isActive ? "active" : ""}>📊 Analytics</NavLink>
            <NavLink to="/service-virtualization/settings" className={({isActive}) => isActive ? "active" : ""}>⚙️ Settings</NavLink>
          </nav>
        </>
      )}
    </aside>
  );
};

export default Sidebar;

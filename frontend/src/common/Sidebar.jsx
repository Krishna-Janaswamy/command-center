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

  const handleToggle = () => {
    setToggle(prev => {
      const newValue = !prev;
      settingsApi.update('useToggle', newValue ? 'on' : 'off').catch(console.error);
      return newValue;
    });
  };
  return (
    <aside className="sidebar">
      <Link to="/" style={{ textDecoration: 'none', color: 'inherit' }}>
        <div className="brand" style={{ cursor: 'pointer' }}>
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
          <nav style={{ marginTop: '8px' }}>
            <NavLink to="/service-virtualization/api" className={({isActive}) => isActive ? "active" : ""}>🔌 API Request</NavLink>
            <NavLink to="/service-virtualization/stubs" className={({isActive}) => isActive ? "active" : ""}>🎭 Stubs</NavLink>
            <NavLink to="/service-virtualization/requests" className={({isActive}) => isActive ? "active" : ""}>📋 Recorded Requests</NavLink>
            <NavLink to="/service-virtualization/analytics" className={({isActive}) => isActive ? "active" : ""}>📊 Analytics</NavLink>
            <NavLink to="/service-virtualization/settings" className={({isActive}) => isActive ? "active" : ""}>⚙️ Settings</NavLink>
          </nav>

          <div style={{ margin: '24px 16px', padding: '16px', background: 'rgba(255,255,255,0.6)', borderRadius: '12px', border: '1px solid var(--border-color)', display: 'flex', flexDirection: 'column', gap: '12px', boxShadow: '0 2px 4px rgba(0,0,0,0.02)' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)' }}>Request Mode</span>
              <div 
                onClick={handleToggle}
                style={{
                  width: '50px', height: '28px', borderRadius: '14px',
                  background: toggle ? 'var(--success)' : 'var(--primary)',
                  position: 'relative', cursor: 'pointer', transition: 'all 0.3s cubic-bezier(0.4, 0.0, 0.2, 1)',
                  boxShadow: 'inset 0 1px 3px rgba(0,0,0,0.2)',
                  display: 'flex', alignItems: 'center'
                }}
              >
                <div style={{
                  width: '24px', height: '24px', borderRadius: '50%', background: 'white',
                  position: 'absolute', top: '2px', left: toggle ? '24px' : '2px',
                  transition: 'all 0.3s cubic-bezier(0.4, 0.0, 0.2, 1)', boxShadow: '0 2px 4px rgba(0,0,0,0.2)'
                }} />
              </div>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem', fontWeight: 600 }}>
              <span style={{ color: 'var(--primary)', opacity: !toggle ? 1 : 0, transition: 'opacity 0.3s' }}>STUBBED</span>
              <span style={{ color: 'var(--success)', opacity: toggle ? 1 : 0, transition: 'opacity 0.3s' }}>LIVE</span>
            </div>
          </div>
        </>
      )}
    </aside>
  );
};

export default Sidebar;

import React, { useState } from 'react';
import { NavLink, Link } from 'react-router-dom';

const Sidebar = ({ user }) => {
  const [healthOpen, setHealthOpen] = useState(true);
  const [virtOpen, setVirtOpen] = useState(true);
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
        <nav>
          <NavLink to="/service-virtualization/api" className={({isActive}) => isActive ? "active" : ""}>🔌 API Request</NavLink>
          <NavLink to="/service-virtualization/stubs" className={({isActive}) => isActive ? "active" : ""}>🎭 Stubs</NavLink>
          <NavLink to="/service-virtualization/requests" className={({isActive}) => isActive ? "active" : ""}>📋 Recorded Requests</NavLink>
          <NavLink to="/service-virtualization/analytics" className={({isActive}) => isActive ? "active" : ""}>📊 Analytics</NavLink>
          <NavLink to="/service-virtualization/settings" className={({isActive}) => isActive ? "active" : ""}>⚙️ Settings</NavLink>
        </nav>
      )}
    </aside>
  );
};

export default Sidebar;

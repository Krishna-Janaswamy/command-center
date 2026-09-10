import React, { useState, useEffect } from 'react';
import { NavLink, Link } from 'react-router-dom';
import { settingsApi } from '../services/registryApi';

const Sidebar = ({ user, onNavigate }) => {
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
      <Link to="/" style={{ textDecoration: 'none', color: 'inherit' }} onClick={onNavigate}>
        <div className="brand" style={{ cursor: 'pointer' }}>
          Command Center
        </div>
      </Link>

      <div className="sidebar-scroll">
        <button
          type="button"
          className="sidebar-section-toggle"
          onClick={() => setHealthOpen(!healthOpen)}
        >
          <span>Health Monitoring</span>
          <span>{healthOpen ? '▼' : '▶'}</span>
        </button>
        {healthOpen && (
          <nav>
            <NavLink to="/api-registry" className={({ isActive }) => isActive ? 'active' : ''} onClick={onNavigate}>API Registry</NavLink>
            <NavLink to="/api-health" className={({ isActive }) => isActive ? 'active' : ''} onClick={onNavigate}>API Health</NavLink>
          </nav>
        )}

        <button
          type="button"
          className="sidebar-section-toggle"
          onClick={() => setVirtOpen(!virtOpen)}
        >
          <span>Service Virtualization</span>
          <span>{virtOpen ? '▼' : '▶'}</span>
        </button>
        {virtOpen && (
          <>
            <nav>
              <NavLink to="/service-virtualization/api" className={({ isActive }) => isActive ? 'active' : ''} onClick={onNavigate}>API Request</NavLink>
              <NavLink to="/service-virtualization/stubs" className={({ isActive }) => isActive ? 'active' : ''} onClick={onNavigate}>Stubs</NavLink>
              <NavLink to="/service-virtualization/requests" className={({ isActive }) => isActive ? 'active' : ''} onClick={onNavigate}>Recorded Requests</NavLink>
              <NavLink to="/service-virtualization/analytics" className={({ isActive }) => isActive ? 'active' : ''} onClick={onNavigate}>Analytics</NavLink>
              <NavLink to="/service-virtualization/settings" className={({ isActive }) => isActive ? 'active' : ''} onClick={onNavigate}>Settings</NavLink>
            </nav>

            <div className="sidebar-mode-card">
              <div className="sidebar-mode-row">
                <span>Request Mode</span>
                <button
                  type="button"
                  className={`mode-switch${toggle ? ' is-live' : ''}`}
                  aria-label="Toggle request mode"
                  onClick={handleToggle}
                >
                  <span className="mode-switch-knob" />
                </button>
              </div>
              <div className="sidebar-mode-labels">
                <span className={!toggle ? 'is-active' : ''}>STUBBED</span>
                <span className={toggle ? 'is-active' : ''}>LIVE</span>
              </div>
            </div>
          </>
        )}
      </div>
    </aside>
  );
};

export default Sidebar;

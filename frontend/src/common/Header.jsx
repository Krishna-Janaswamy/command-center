import React, { useState, useEffect } from 'react';
import { settingsApi } from '../services/registryApi';

const Header = ({ user, onLogout }) => {
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
    <header>
      <div style={{ display: 'flex', alignItems: 'center', gap: '24px' }}>
        <h2 style={{ margin: 0, fontSize: '1.25rem' }}>Workspace</h2>
        
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', background: 'rgba(0,0,0,0.2)', padding: '6px 12px', borderRadius: '20px', border: '1px solid var(--border-color)' }}>
          <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Real Time API</span>
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
      
      <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
        <div style={{ textAlign: 'right' }}>
          <div style={{ fontWeight: 500 }}>{user?.sub || user?.username}</div>
          <div style={{ fontSize: '0.8rem', color: 'var(--info)' }}>{user?.role}</div>
        </div>
        <button className="btn btn-secondary" onClick={onLogout}>Logout</button>
      </div>
    </header>
  );
};

export default Header;

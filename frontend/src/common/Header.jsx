import React, { useState, useEffect } from 'react';


const Header = ({ user, onLogout }) => {


  return (
    <header>
      <div style={{ display: 'flex', alignItems: 'center', gap: '24px' }}>
        <h2 style={{ margin: 0, fontSize: '1.25rem' }}>Control Panel</h2>
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

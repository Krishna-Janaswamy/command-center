import React from 'react';

const Header = ({ user, onLogout, onMenuClick }) => {
  return (
    <header className="app-header">
      <div className="app-header-left">
        <button
          type="button"
          className="menu-toggle"
          aria-label="Open navigation"
          onClick={onMenuClick}
        >
          <span />
          <span />
          <span />
        </button>
        <h2>Control Panel</h2>
      </div>

      <div className="app-header-right">
        <div className="app-header-user">
          <div className="app-header-username">{user?.sub || user?.username}</div>
          <div className="app-header-role">{user?.role}</div>
        </div>
        <button className="btn btn-secondary" onClick={onLogout}>Logout</button>
      </div>
    </header>
  );
};

export default Header;

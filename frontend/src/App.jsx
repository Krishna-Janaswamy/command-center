import React, { useState, useEffect } from 'react';
import { Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom';
import Sidebar from './common/Sidebar';
import Header from './common/Header';
import LoginPage from './components/auth/LoginPage';
import RegisterPage from './components/auth/RegisterPage';
import { authApi } from './services/registryApi';
import Dashboard from './common/Dashboard';
import './index.css';

// Lazy loading placeholders (will implement real components later)
import ApiRegistry from './components/registry/ApiRegistry';
import ApiHealth from './components/health/ApiHealth';
import APITesterTab from './components/tester/APITesterTab';
import StubsTab from './components/stubs/StubsTab';
import RequestsTab from './components/requests/RequestsTab';
import AnalyticsTab from './components/analytics/AnalyticsTab';
import SettingsTab from './components/settings/SettingsTab';

function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const verifyUser = async () => {
      const token = localStorage.getItem('token');
      if (token) {
        try {
          const res = await authApi.verify();
          setUser(res.data);
        } catch (err) {
          localStorage.removeItem('token');
        }
      }
      setLoading(false);
    };
    verifyUser();
  }, []);

  const handleLogout = async () => {
    await authApi.logout();
    setUser(null);
    navigate('/login');
  };

  if (loading) {
    return <div className="auth-page">Loading...</div>;
  }

  if (!user) {
    return (
      <Routes>
        <Route path="/login" element={<LoginPage onLogin={setUser} />} />
        <Route path="/register" element={<RegisterPage onRegister={setUser} />} />
        <Route path="*" element={<Navigate to="/login" state={{ from: location }} replace />} />
      </Routes>
    );
  }

  return (
    <div className="app-shell fade-in">
      <Sidebar user={user} />
      <main>
        <Header user={user} onLogout={handleLogout} />
        <div className="content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/api-registry" element={<ApiRegistry />} />
            <Route path="/api-health" element={<ApiHealth />} />
            <Route path="/service-virtualization/api" element={<APITesterTab />} />
            <Route path="/service-virtualization/stubs" element={<StubsTab />} />
            <Route path="/service-virtualization/requests" element={<RequestsTab />} />
            <Route path="/service-virtualization/analytics" element={<AnalyticsTab />} />
            <Route path="/service-virtualization/settings" element={<SettingsTab />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </div>
      </main>
    </div>
  );
}

export default App;

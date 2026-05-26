import React from 'react';
import { useNavigate } from 'react-router-dom';

const Dashboard = () => {
  const navigate = useNavigate();

  return (
    <div className="fade-in" style={{ padding: '12px 0' }}>
      <div style={{ 
        background: 'linear-gradient(135deg, rgba(139, 92, 246, 0.15) 0%, rgba(59, 130, 246, 0.05) 100%)',
        borderRadius: '20px',
        padding: '56px 48px',
        marginBottom: '48px',
        border: '1px solid rgba(255, 255, 255, 0.05)',
        boxShadow: '0 20px 40px -20px rgba(0,0,0,0.5)',
        position: 'relative',
        overflow: 'hidden'
      }}>
        <div style={{ position: 'relative', zIndex: 1 }}>
          <h1 style={{ fontSize: '3.2rem', marginBottom: '16px', fontWeight: 700, letterSpacing: '-1px', background: 'linear-gradient(to right, #ffffff, #a78bfa)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent' }}>
            Command Center
          </h1>
          <p style={{ color: 'var(--text-muted)', fontSize: '1.15rem', maxWidth: '650px', lineHeight: '1.6', marginBottom: '36px' }}>
            The ultimate hub for API orchestration. Manage your external registries, monitor real-time health, and seamlessly virtualize network traffic with intelligent stubs.
          </p>
          <div style={{ display: 'flex', gap: '16px' }}>
            <button className="btn" style={{ padding: '14px 28px', fontSize: '1rem', borderRadius: '8px' }} onClick={() => navigate('/api-registry')}>
              Register an API
            </button>
            <button className="btn-secondary" style={{ padding: '14px 28px', fontSize: '1rem', borderRadius: '8px' }} onClick={() => navigate('/service-virtualization/api')}>
              Test Proxy
            </button>
          </div>
        </div>
        {/* Decorative background element */}
        <div style={{ 
          position: 'absolute', top: '-40%', right: '-5%', width: '700px', height: '700px', 
          background: 'radial-gradient(circle, rgba(139, 92, 246, 0.1) 0%, transparent 60%)', 
          borderRadius: '50%', pointerEvents: 'none' 
        }} />
      </div>

      <div className="grid-3">
        <div className="glass-panel" style={{ transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)', cursor: 'pointer' }} onMouseOver={e => { e.currentTarget.style.transform = 'translateY(-6px)'; e.currentTarget.style.boxShadow = '0 20px 25px -5px rgba(0, 0, 0, 0.2)'; e.currentTarget.style.borderColor = 'rgba(139, 92, 246, 0.4)'; }} onMouseOut={e => { e.currentTarget.style.transform = 'translateY(0)'; e.currentTarget.style.boxShadow = 'var(--glass-shadow)'; e.currentTarget.style.borderColor = 'var(--border-color)'; }} onClick={() => navigate('/api-registry')}>
          <div style={{ fontSize: '2.5rem', marginBottom: '20px' }}>📋</div>
          <h3 style={{ marginBottom: '12px', fontSize: '1.3rem' }}>API Registry</h3>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.95rem', lineHeight: '1.6' }}>
            Centralize your external dependencies. Configure strict health checks, environment parameters, and intelligent retry logic.
          </p>
        </div>

        <div className="glass-panel" style={{ transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)', cursor: 'pointer' }} onMouseOver={e => { e.currentTarget.style.transform = 'translateY(-6px)'; e.currentTarget.style.boxShadow = '0 20px 25px -5px rgba(0, 0, 0, 0.2)'; e.currentTarget.style.borderColor = 'rgba(139, 92, 246, 0.4)'; }} onMouseOut={e => { e.currentTarget.style.transform = 'translateY(0)'; e.currentTarget.style.boxShadow = 'var(--glass-shadow)'; e.currentTarget.style.borderColor = 'var(--border-color)'; }} onClick={() => navigate('/service-virtualization/stubs')}>
          <div style={{ fontSize: '2.5rem', marginBottom: '20px' }}>🧪</div>
          <h3 style={{ marginBottom: '12px', fontSize: '1.3rem' }}>Smart Virtualization</h3>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.95rem', lineHeight: '1.6' }}>
            Bypass brittle environments by mocking responses. Utilize regex matching, simulate network delays, and version your stubs.
          </p>
        </div>

        <div className="glass-panel" style={{ transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)', cursor: 'pointer' }} onMouseOver={e => { e.currentTarget.style.transform = 'translateY(-6px)'; e.currentTarget.style.boxShadow = '0 20px 25px -5px rgba(0, 0, 0, 0.2)'; e.currentTarget.style.borderColor = 'rgba(139, 92, 246, 0.4)'; }} onMouseOut={e => { e.currentTarget.style.transform = 'translateY(0)'; e.currentTarget.style.boxShadow = 'var(--glass-shadow)'; e.currentTarget.style.borderColor = 'var(--border-color)'; }} onClick={() => navigate('/service-virtualization/requests')}>
          <div style={{ fontSize: '2.5rem', marginBottom: '20px' }}>📡</div>
          <h3 style={{ marginBottom: '12px', fontSize: '1.3rem' }}>Request Capture</h3>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.95rem', lineHeight: '1.6' }}>
            Auto-record all requests. Convert real requests into reusable, version-controlled stubs without writing a single line of JSON.
          </p>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;

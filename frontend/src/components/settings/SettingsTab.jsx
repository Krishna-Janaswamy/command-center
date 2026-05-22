import React, { useState, useEffect } from 'react';
import { settingsApi } from '../../services/registryApi';

const SettingsTab = () => {
  const [settings, setSettings] = useState({
    recordingMode: 'true',
    playbackMode: 'true',
    useToggle: 'true'
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    settingsApi.getAll().then(res => {
      setSettings(res.data);
      setLoading(false);
    });
  }, []);

  const handleChange = (key, value) => {
    setSettings(prev => ({ ...prev, [key]: value }));
  };

  const handleSave = async (key) => {
    setSaving(true);
    try {
      await settingsApi.update(key, settings[key]);
    } catch (err) {
      console.error(err);
      alert('Failed to update setting');
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <div className="fade-in">Loading settings...</div>;

  const toggleToBool = (val) => val === 'true' || val === 'on';

  return (
    <div className="fade-in">
      <h1>Global Settings</h1>
      
      <div className="grid-2">
        <div className="glass-panel">
          <h3>Operation Modes</h3>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', marginBottom: '24px' }}>
            Control the global behavior of the service virtualization proxy.
          </p>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <div style={{ fontWeight: 500 }}>Real Time Proxy</div>
                <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>Bypass stubs globally and forward all traffic.</div>
              </div>
              <label style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
                <input 
                  type="checkbox" 
                  checked={toggleToBool(settings.useToggle)} 
                  onChange={(e) => {
                    const val = e.target.checked ? 'on' : 'off';
                    handleChange('useToggle', val);
                    handleSave('useToggle');
                  }}
                  style={{ width: '20px', height: '20px', cursor: 'pointer' }}
                />
              </label>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <div style={{ fontWeight: 500 }}>Auto-Record Live Traffic</div>
                <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>Automatically create stubs for successful (2xx) proxy responses.</div>
              </div>
              <label style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
                <input 
                  type="checkbox" 
                  checked={toggleToBool(settings.recordingMode)} 
                  onChange={(e) => {
                    const val = e.target.checked ? 'true' : 'false';
                    handleChange('recordingMode', val);
                    handleSave('recordingMode');
                  }}
                  style={{ width: '20px', height: '20px', cursor: 'pointer' }}
                />
              </label>
            </div>
          </div>
        </div>

        <div className="glass-panel" style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
          <h3>How It Works</h3>
          
          <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
            <div style={{ background: 'var(--primary)', color: 'white', width: '32px', height: '32px', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 'bold', flexShrink: 0, boxShadow: '0 0 10px rgba(139, 92, 246, 0.5)' }}>1</div>
            <div>
              <div style={{ fontWeight: 600, color: 'var(--text-main)', marginBottom: '6px', fontSize: '1.05rem' }}>Authentication</div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.9rem', lineHeight: '1.5' }}>
                All APIs are secured via custom JWT tokens. Include <code style={{ background: 'rgba(0,0,0,0.3)', border: '1px solid var(--border-color)', padding: '2px 6px', borderRadius: '4px', color: 'var(--info)' }}>Authorization: Bearer &lt;token&gt;</code> in your requests.
              </div>
            </div>
          </div>
          
          <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
            <div style={{ background: 'var(--info)', color: 'white', width: '32px', height: '32px', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 'bold', flexShrink: 0, boxShadow: '0 0 10px rgba(59, 130, 246, 0.5)' }}>2</div>
            <div>
              <div style={{ fontWeight: 600, color: 'var(--text-main)', marginBottom: '6px', fontSize: '1.05rem' }}>Targeting Domains</div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.9rem', lineHeight: '1.5' }}>
                To proxy a request to a live domain (e.g. api.example.com), specify the domain in the <code style={{ background: 'rgba(0,0,0,0.3)', border: '1px solid var(--border-color)', padding: '2px 6px', borderRadius: '4px', color: 'var(--info)' }}>X-Target-Host</code> header.
              </div>
            </div>
          </div>

          <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
            <div style={{ background: 'var(--success)', color: 'white', width: '32px', height: '32px', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 'bold', flexShrink: 0, boxShadow: '0 0 10px rgba(16, 185, 129, 0.5)' }}>3</div>
            <div>
              <div style={{ fontWeight: 600, color: 'var(--text-main)', marginBottom: '6px', fontSize: '1.05rem' }}>Mock vs Live</div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.9rem', lineHeight: '1.5' }}>
                Requests sent to <code style={{ background: 'rgba(0,0,0,0.3)', border: '1px solid var(--border-color)', padding: '2px 6px', borderRadius: '4px' }}>/api/proxy-request</code> are evaluated against saved Stubs. If no active stub matches, the proxy falls back to the live API and <span style={{ color: 'var(--success)', fontWeight: 500 }}>auto-records</span> the response.
              </div>
            </div>
          </div>

          <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
            <div style={{ background: 'var(--warning)', color: '#fff', width: '32px', height: '32px', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 'bold', flexShrink: 0, boxShadow: '0 0 10px rgba(245, 158, 11, 0.5)' }}>4</div>
            <div>
              <div style={{ fontWeight: 600, color: 'var(--text-main)', marginBottom: '6px', fontSize: '1.05rem' }}>Header Overrides</div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.9rem', lineHeight: '1.5' }}>
                Send <code style={{ background: 'rgba(0,0,0,0.3)', border: '1px solid var(--border-color)', padding: '2px 6px', borderRadius: '4px', color: 'var(--warning)' }}>X-Use-Toggle: on</code> to forcefully bypass all mocks and hit the live API directly for a specific request.
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SettingsTab;

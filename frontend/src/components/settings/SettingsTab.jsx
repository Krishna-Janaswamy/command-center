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
            Control the global behavior of the service virtualization.
          </p>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <div style={{ fontWeight: 500 }}>Live API</div>
                <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>API hits real time server.</div>
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
                <div style={{ fontWeight: 500 }}>Auto-Record Live Request</div>
                <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>Automatically create stubs.</div>
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


      </div>
    </div>
  );
};

export default SettingsTab;

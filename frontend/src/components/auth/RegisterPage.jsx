import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { authApi } from '../../services/registryApi';

const RegisterPage = ({ onRegister }) => {
  const [formData, setFormData] = useState({ username: '', password: '', email: '', adGroup: 'QED_DEFAULT_USER' });
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const res = await authApi.register(formData);
      localStorage.setItem('token', res.data.token);
      // server now returns the created user to avoid extra verify round-trip
      const user = res.data && res.data.user ? res.data.user : null;
      if (user) {
        onRegister(user);
      } else {
        // fallback to verify if user not returned
        const userRes = await authApi.verify();
        onRegister(userRes.data);
      }
      navigate('/', { replace: true });
    } catch (err) {
      setError(err.response?.data?.error || 'Registration failed');
    }
  };

  return (
    <div className="auth-page">
      <div className="auth-card glass-panel fade-in">
        <h2 style={{ textAlign: 'center', marginBottom: '24px', color: 'var(--primary)' }}>Create Account</h2>
        {error && <div className="badge badge-danger" style={{ display: 'block', marginBottom: '16px', textAlign: 'center' }}>{error}</div>}
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Username</label>
            <input type="text" className="form-control" value={formData.username} onChange={e => setFormData({ ...formData, username: e.target.value })} required />
          </div>
          <div className="form-group">
            <label>Password</label>
            <input type="password" className="form-control" value={formData.password} onChange={e => setFormData({ ...formData, password: e.target.value })} required />
          </div>
          <div className="form-group">
            <label>Email (optional)</label>
            <input type="email" className="form-control" value={formData.email} onChange={e => setFormData({ ...formData, email: e.target.value })} />
          </div>
          <button type="submit" className="btn" style={{ width: '100%', justifyContent: 'center', marginTop: '12px' }}>Register</button>
        </form>
        <p style={{ textAlign: 'center', marginTop: '24px', fontSize: '0.9rem', color: 'var(--text-muted)' }}>
          Already have an account? <Link to="/login" style={{ color: 'var(--info)', textDecoration: 'none' }}>Login</Link>
        </p>
      </div>
    </div>
  );
};

export default RegisterPage;

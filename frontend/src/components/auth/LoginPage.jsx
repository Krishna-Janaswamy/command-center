import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { authApi } from '../../services/registryApi';

const LoginPage = ({ onLogin }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const res = await authApi.login({ username, password });
      localStorage.setItem('token', res.data.token);
      // server may return user in response to avoid extra verify round-trip
      const user = res.data && res.data.user ? res.data.user : null;
      if (user) {
        onLogin(user);
      } else {
        const userRes = await authApi.verify();
        onLogin(userRes.data);
      }
      navigate('/', { replace: true });
    } catch (err) {
      setError(err.response?.data?.error || 'Login failed');
    }
  };

  return (
    <div className="auth-page">
      <div className="auth-card glass-panel fade-in">
        <h2 style={{ textAlign: 'center', marginBottom: '24px', color: 'var(--primary)' }}>Command Center</h2>
        {error && <div className="badge badge-danger" style={{ display: 'block', marginBottom: '16px', textAlign: 'center' }}>{error}</div>}
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Username</label>
            <input type="text" className="form-control" value={username} onChange={e => setUsername(e.target.value)} required />
          </div>
          <div className="form-group">
            <label>Password</label>
            <input type="password" className="form-control" value={password} onChange={e => setPassword(e.target.value)} required />
          </div>
          <button type="submit" className="btn" style={{ width: '100%', justifyContent: 'center', marginTop: '12px' }}>Login</button>
        </form>
        <p style={{ textAlign: 'center', marginTop: '24px', fontSize: '0.9rem', color: 'var(--text-muted)' }}>
          Don't have an account? <Link to="/register" style={{ color: 'var(--info)', textDecoration: 'none' }}>Register</Link>
        </p>
      </div>
    </div>
  );
};

export default LoginPage;

import axios from 'axios';

const api = axios.create({
  baseURL: '/api'
});

// Auth requests go directly to Go backend on port 3001 (login/register)
const authApiDirect = axios.create({
  baseURL: 'http://localhost:3001/api'
});

// Mirror same auth interceptor so authApiDirect sends stored token
authApiDirect.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

authApiDirect.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      const url = error.config && error.config.url ? error.config.url : '';
      // Don't auto-redirect on verify-token (it is used to check session)
      if (!url.endsWith('/auth/verify-token')) {
        localStorage.removeItem('token');
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      const url = error.config && error.config.url ? error.config.url : '';
      if (!url.endsWith('/auth/verify-token')) {
        localStorage.removeItem('token');
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export const authApi = {
  login: (credentials) => authApiDirect.post('/auth/login', credentials),
  register: (data) => authApiDirect.post('/auth/register', data),
  verify: () => authApiDirect.get('/auth/verify-token'),
  logout: () => {
    localStorage.removeItem('token');
    return api.post('/auth/logout');
  }
};

export const registryApi = {
  getAll: (env, category) => api.get('/registry', { params: { env, category } }),
  getById: (id) => api.get(`/registry/${id}`),
  create: (data) => api.post('/registry', data),
  update: (id, data) => api.patch(`/registry/${id}`, data),
  delete: (id) => api.delete(`/registry/${id}`),
};

export const healthApi = {
  check: (data) => api.post('/health-check', data)
};

export const settingsApi = {
  getAll: () => api.get('/settings'),
  update: (key, value) => api.put(`/settings/${key}`, { value })
};

export const requestApi = {
  getAll: () => api.get('/recorded-requests'),
  delete: (id) => api.delete(`/requests/${id}`),
  deleteAll: () => api.delete('/requests')
};

export const stubApi = {
  getAll: () => api.get('/stubs'),
  create: (data) => api.post('/stubs', data),
  update: (id, data) => api.put(`/stubs/${id}`, data),
  delete: (id) => api.delete(`/stubs/${id}`),
  deleteAll: () => api.delete('/stubs'),
  toggle: (id) => api.post(`/stubs/${id}/toggle`),
  getVersions: (id) => api.get(`/stubs/${id}/versions`),
  createVersion: (id, data) => api.post(`/stubs/${id}/versions`, data),
  updateVersion: (stubId, versionId, data) => api.put(`/stubs/${stubId}/versions/${versionId}`, data),
  deleteVersion: (stubId, versionId) => api.delete(`/stubs/${stubId}/versions/${versionId}`),
  activateVersion: (stubId, versionId) => api.post(`/stubs/${stubId}/versions/${versionId}/activate`),
};

export default api;

import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 8080,
    proxy: {
      // Proxy frontend API calls to the Go Lambda-compatible local server.
      '/api': 'http://localhost:3001',
      '/virtualize': 'http://localhost:3001'
    }
  }
});

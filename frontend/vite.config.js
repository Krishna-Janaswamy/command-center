import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 8080,
    proxy: {
      // Proxy non-auth API calls to Spring backend on port 3002
      '/api': 'http://localhost:3002',
      '/virtualize': 'http://localhost:3002'
    }
  }
});

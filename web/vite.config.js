import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  publicDir: 'static',
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:10086',
        changeOrigin: true,
        // rewrite: (path) => path.replace(/^\/api/, '')
      },
      '/index': {
        target: 'http://127.0.0.1:10086',
        changeOrigin: true
      },

      '/ws': {
        target: 'ws://127.0.0.1:10086',
        ws: true,
        changeOrigin: true
      },
      '/img': {
        target: 'http://127.0.0.1:10086',
        changeOrigin: true
      }
    }
  }
}) 
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

const baseURL = "http://127.0.0.1:8082"

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
        target: baseURL,
        changeOrigin: true,
        // rewrite: (path) => path.replace(/^\/api/, '')
      },
      '/index': {
        target: baseURL,
        changeOrigin: true
      },

      '/ws': {
        target: baseURL,
        ws: true,
        changeOrigin: true
      },
      '/img': {
        target: baseURL,
        changeOrigin: true
      }
    }
  }
}) 
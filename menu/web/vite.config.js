import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
// import { visualizer } from 'rollup-plugin-visualizer'
// import styleImport from 'vite-plugin-style-import'

const baseURL = "http://localhost:8083"

export default defineConfig({
  plugins: [
    vue(),
    // visualizer(), // 打包分析（如未安装可注释）
    // styleImport({
    //   libs: [
    //     {
    //       libraryName: 'ant-design-vue',
    //       esModule: true,
    //       resolveStyle: (name) => {
    //         return `ant-design-vue/es/${name}/style/index`
    //       }
    //     }
    //   ]
    // })
  ],
  publicDir: 'static',
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    host: '0.0.0.0', // 新增此行，允许外部访问
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
  },
  build: {
    // minify: 'terser', // 删除或注释此行，使用默认压缩
    terserOptions: {
      compress: {
        drop_console: true, // 移除所有 console
        drop_debugger: true
      }
    },
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            return 'vendor'
          }
        }
      }
    }
  },
})
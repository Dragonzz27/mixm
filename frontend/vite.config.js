import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 5155,
    host: true,
    allowedHosts: ['mixm-edu.yoobit.cn', 'https://mixm-edu.yoobit.cn/'],
    proxy: {
      '/api': {
        target: 'http://localhost:18002',
        changeOrigin: true
      }
    }
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@import "@/styles/variables.scss";`
      }
    }
  }
})

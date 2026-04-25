import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'
import solid from 'vite-plugin-solid'

const supervisorOrigin = 'http://127.0.0.1:6700'

export default defineConfig({
  plugins: [solid(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      '/api': { target: supervisorOrigin, changeOrigin: true, ws: true },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})

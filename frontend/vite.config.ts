import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': process.env.AURA_GATEWAY_URL ?? 'http://aura-platform.home',
      '/s/': process.env.AURA_GATEWAY_URL ?? 'http://aura-platform.home',
    },
  },
  test: { environment: 'jsdom', setupFiles: './src/test/setup.ts' },
})

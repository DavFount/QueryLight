import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  use: {
    baseURL: 'http://127.0.0.1:5179',
    channel: 'msedge',
    viewport: { width: 1280, height: 820 },
    trace: 'retain-on-failure',
  },
  webServer: {
    command: 'npm run dev -- --host 127.0.0.1 --port 5179 --strictPort',
    url: 'http://127.0.0.1:5179',
    reuseExistingServer: false,
  },
})

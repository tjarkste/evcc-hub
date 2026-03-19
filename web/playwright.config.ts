import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',
  timeout: 30000,
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:7071',
    headless: true,
  },
  webServer: process.env.CI ? undefined : {
    command: 'npm run dev',
    url: 'http://localhost:7071',
    reuseExistingServer: true,
    timeout: 30000,
  },
})

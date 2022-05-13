const { devices } = require('@playwright/test')

const config = {
  testDir: './integrationtest',
  globalSetup: './integrationtest/global-setup.ts',
  use: {
    trace: 'on-first-retry',
    baseURL: 'http://localhost:3040',
    storageState: 'storageState.json',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
  ],
  webServer: {
    command: 'make cy-wide',
    url: 'http://localhost:3040/health',
    reuseExistingServer: true,
  },
}

module.exports = config

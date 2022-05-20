import { devices } from '@playwright/test'

const config = {
  testDir: './integrationtest',
  globalSetup: './integrationtest/global-setup.ts',
  use: {
    trace: 'on-first-retry',
    baseURL: 'http://localhost:3040',
    viewport: { width: 1440, height: 900 },
    timezoneId: 'America/Chicago',
    launchOptions: {
      // slowMo: 1000,
    },
  },
  projects: [
    {
      name: 'chromium-wide',
      use: {
        ...devices['Desktop Chrome'],
        viewportSize: { width: 1440, height: 900 },
      },
    },
    {
      name: 'chromium-mobile',
      use: {
        ...devices['Pixel 5'],
        viewportSize: { width: 375, height: 667 },
      },
    },
    {
      name: 'firefox-wide',
      use: { ...devices['Desktop Firefox'] },
      viewportSize: { width: 1440, height: 900 },
    },
  ],
  webServer: {
    command: 'make cy-wide',
    url: 'http://localhost:3040/health',
    reuseExistingServer: true,
  },
}

module.exports = config

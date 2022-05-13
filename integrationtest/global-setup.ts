import { chromium, FullConfig } from '@playwright/test'
import { login } from './login'

export default async function globalSetup(config: FullConfig): Promise<void> {
  const { storageState } = config.projects[0].use
  const browser = await chromium.launch()
  const page = await browser.newPage({
    baseURL: config.projects[0].use.baseURL,
  })
  await page.goto('./')
  await login(page)
  await page.context().storageState({
    path: storageState as string,
  })
  await browser.close()
}

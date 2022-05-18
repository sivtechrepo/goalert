import { chromium, FullConfig } from '@playwright/test'
import { adminSession, userSession, login } from './login'
import { adminUser, normalUser } from './util'

export default async function globalSetup(config: FullConfig): Promise<void> {
  const browser = await chromium.launch()

  async function createSession(
    path: string,
    user: string,
    pass: string,
  ): Promise<void> {
    const page = await browser.newPage({
      baseURL: config.projects[0].use.baseURL,
    })
    await page.goto('./')
    await login(page, user, pass)
    await page.context().storageState({ path })
    if (user !== 'test.admin') {
      return
    }

    // click on admin
    await page.locator('nav a', { hasText: 'Admin' }).click()
    // click on SMTP header
    await page.click('h2 >> "SMTP"')
    await page.check('input[name="SMTP.Enable"]')
    await page.fill('input[name="SMTP.From"]', 'goalert-test@localhost')
    await page.fill('input[name="SMTP.Address"]', 'localhost:1025')
    await page.check('input[name="SMTP.DisableTLS"]')
    // save and confirm
    if (await page.locator('button >> "Save"').isEnabled()) {
      await page.click('button >> "Save"')
      await page.click('button >> "Confirm"')
    }
  }

  await Promise.all([
    createSession(adminSession, adminUser.name, adminUser.pass),
    createSession(userSession, normalUser.name, normalUser.pass),
  ])

  await browser.close()
}

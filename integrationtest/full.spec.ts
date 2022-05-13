import { test, expect, Page } from '@playwright/test'
import {
  login,
  logout,
  newTestUser,
  deleteUser,
  createService,
  deleteService,
} from './util'

import Chance from 'chance'
const c = new Chance()

test('everything', async ({ page, browser }) => {
  await page.goto('./')
  const u = await newTestUser(page)

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

  await logout(page)
  await login(page, u.name, u.pass)

  // manage profile
  await page.locator('[aria-label="Manage Profile"]').click()
  await page.click('button >> "Manage Profile"')

  await page.hover('button[aria-label="Add Items"]')
  await page.click('[aria-label="Add Contact Method"]')

  await page.fill('input[name=name]', 'test_contact_method')
  const addr = c.email()
  await page.fill('input[name=value]', addr)
  await page.click('button >> "Submit"')

  // get verification code in new browser page
  const mailPage = await browser.newPage()
  await mailPage.goto(`http://localhost:8025`)
  await mailPage.click(`div >> "${addr}"`)
  const codeSrc = (await mailPage
    .locator('#preview-plain')
    .textContent()) as string

  const m = codeSrc.match(/\d{6}/)
  if (!m) {
    throw new Error('could not find verification code')
  }
  const code = m[0]

  await page.fill('input[name=code]', code)
  await page.click('button >> "Submit"')

  const svc = await createService(page)
  await page.locator('nav a', { hasText: 'Services' }).click()
  await page.fill('input[name=search]', svc.name)
  await page.locator('li', { hasText: svc.name }).click()
  await page.locator('li', { hasText: 'Alerts' }).first().click()

  const summary = c.word({ length: 20 })
  await page.locator('[aria-label="Create Alert"]').click()
  await page.fill('input[name=summary]', summary)
  await page.click('button >> "Next"')
  await page.click('button >> "Submit"')
  const [alertTab] = await Promise.all([
    page.waitForEvent('popup'),
    page
      .locator('ul[aria-label="Successfully created alerts"] li a')
      .first()
      .click(),
  ])
  await expect(
    await alertTab.locator('[data-cy="alert-logs"] li', {
      hasText: `Notification sent to ${u.name} (Email)`,
    }),
  ).toBeVisible({ timeout: 10000 })
  await alertTab.close()
  await page.click('button >> "Done"')

  await mailPage.locator('li a', { hasText: 'Inbox' }).click()
  await mailPage.locator(`div.msglist-message`, { hasText: summary }).click()

  const [mailAlertTab] = await Promise.all([
    mailPage.waitForEvent('popup'),
    mailPage
      .frameLocator('#preview-html')
      .locator('a >> "Open Alert Details"')
      .click(),
  ])
  await login(mailAlertTab, u.name, u.pass)
  await expect(
    await mailAlertTab.locator('h2', { hasText: summary }),
  ).toBeVisible()
  await mailAlertTab.close()
  await mailPage.close()

  await deleteService(page, svc.name)
  await logout(page)
  await login(page)
  await deleteUser(page, u.name)
})

import { test, expect } from '@playwright/test'
import { normalUser } from './login'

test('basic auth', async ({ page }) => {
  await page.goto('./')
  await page.fill('input[name=username]', normalUser.name)
  await page.fill('input[name=password]', 'invalid')
  await page.click('button >> "Login"')

  await expect(await page.locator('#app')).toContainText(
    'unknown username/password',
  )

  await page.fill('input[name=username]', normalUser.name)
  await page.fill('input[name=password]', normalUser.pass)
  await page.click('button >> "Login"')

  await expect(await page.locator('#app')).toContainText(
    'Showing active alerts',
  )

  await page.locator('[aria-label="Manage Profile"]').click()
  await page.locator('button', { hasText: 'Logout' }).click()

  await expect(await page.locator('button >> "Login"')).toBeVisible()
})

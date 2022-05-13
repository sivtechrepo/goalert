import { test, expect } from '@playwright/test'

test('basic auth', async ({ page }) => {
  await page.goto('./')
  await page.fill('input[name=username]', 'admin')
  await page.fill('input[name=password]', 'admin123_invalid')
  await page.click('button >> "Login"')

  await expect(await page.locator('#app')).toContainText(
    'unknown username/password',
  )

  await page.fill('input[name=username]', 'admin')
  await page.fill('input[name=password]', 'admin123')
  await page.click('button >> "Login"')

  await expect(await page.locator('#app')).toContainText(
    'Showing active alerts',
  )

  await page.locator('[aria-label="Manage Profile"]').click()
  await page.locator('button', { hasText: 'Logout' }).click()

  await expect(await page.locator('button >> "Login"')).toBeVisible()
})

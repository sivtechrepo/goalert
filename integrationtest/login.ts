import { Page } from '@playwright/test'

export async function login(
  page: Page,
  user = 'admin',
  pass = 'admin123',
): Promise<void> {
  await page.fill('input[name=username]', user)
  await page.fill('input[name=password]', pass)
  await page.click('button[type=submit] >> "Login"')
}

export async function logout(page: Page): Promise<void> {
  // click logout from manage profile
  await page.locator('[aria-label="Manage Profile"]').click()
  await page.locator('button >> "Logout"').click()
}

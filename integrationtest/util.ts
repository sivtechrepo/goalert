import { expect, Page } from '@playwright/test'
import Chance from 'chance'
const c = new Chance()

export * from './login'

export type User = {
  id: string
  name: string
  pass: string
}

export type Service = {
  id: string
  name: string
}

export async function gotoCategory(
  page: Page,
  category: string,
): Promise<void> {
  await page.locator('nav a', { hasText: category }).click()
}

export async function gotoDetails(
  page: Page,
  category: string,
  name: string,
): Promise<void> {
  await gotoCategory(page, category)
  await page.fill('input[name=search]', name)
  await page.locator('li', { hasText: name }).click()
}
export async function simpleDelete(
  page: Page,
  category: string,
  name: string,
): Promise<void> {
  await gotoDetails(page, category, name)
  await page.locator('[aria-label="Delete"]').click()
  await page.locator('button >> "Confirm"').click()
}

export async function simpleCreate(
  page: Page,
  category: string,
  details: Record<string, string>,
): Promise<string> {
  await gotoCategory(page, category)
  await page
    .locator('[aria-label="Create ' + category.replace(/s$/, '') + '"]')
    .click()
  for (const key in details) {
    await page.locator('input[name="' + key + '"]').fill(details[key])
  }
  await page.locator('button >> "Submit"').click()

  return page.url().split('/').pop() as string
}

export async function deleteService(page: Page, name: string): Promise<void> {
  await simpleDelete(page, 'Services', name)
}

export async function createService(page: Page): Promise<Service> {
  const name = 'zz_test_svc ' + c.word({ length: 12 })
  return {
    name,
    id: await simpleCreate(page, 'Services', { name }),
  }
}

export async function newTestUser(page: Page): Promise<User> {
  const name = 'zz.test.usr.' + c.word({ length: 12 })
  const pass = c.word({ length: 12 })

  return {
    name,
    pass,
    id: await simpleCreate(page, 'Users', {
      username: name,
      password: pass,
      password2: pass,
    }),
  }
}

export async function deleteUser(page: Page, name: string): Promise<void> {
  await simpleDelete(page, 'Users', name)
}

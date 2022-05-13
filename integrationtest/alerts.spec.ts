import { test, expect } from '@playwright/test'

import Chance from 'chance'
const c = new Chance()

const svcPrefix = 'SVC ' + c.word({ length: 12 })

const svc1Name = svcPrefix + ' ' + c.word({ length: 12 })
const svc2Name = svcPrefix + ' ' + c.word({ length: 12 })

let svc1URL: string, svc2URL: string

test.beforeAll(async ({ browser }) => {
  const page = await browser.newPage()
  async function createService(name: string): Promise<string> {
    await page.goto('/services')
    await page.locator('[aria-label="Create Service"]').click()
    await page.locator('input[name=name]').fill(name)
    await page.locator('button[type=submit]').click()
    await expect(await page.locator('h2', { hasText: name })).toBeVisible()
    const url = page.url()
    return url
  }

  const svc1 = createService(svc1Name)
  svc1URL = await svc1
  const svc2 = createService(svc2Name)
  svc2URL = await svc2
  await page.close()
})
test.afterAll(async ({ browser }) => {
  async function deleteService(url: string): Promise<void> {
    const page = await browser.newPage()
    await page.goto(url)
    await page.locator('[aria-label="Delete"]').click()
    await page.locator('button >> "Confirm"').click()
    await page.close()
  }

  await Promise.all([deleteService(svc1URL), deleteService(svc2URL)])
})

test('create alerts from /alerts', async ({ page }) => {
  await page.goto('./alerts')
  await page.locator('[aria-label="Create Alert"]').click()

  await page.locator('input[name=summary]').fill('Test Alert')
  await page.locator('button >> "Next"').click()

  await page.locator('input[name=serviceSearch]').fill(svc1Name)

  await page.locator('[role=button]', { hasText: svc1Name }).click()
  await page.locator('button >> "Next"').click()
  await page.locator('button >> "Submit"').click()

  await expect(await page.locator('[role=dialog]')).toContainText(
    'Successfully created 1 alert',
  )
})

test('click alert check', async ({ page }) => {
  await page.goto('./alerts')
  const el = await page.locator('ul input[type=checkbox]')
  for (let i = 0; i < 100; i++) {
    await el.click()
  }
})

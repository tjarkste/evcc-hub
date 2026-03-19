import { test, expect } from '@playwright/test'

test('can switch to register form', async ({ page }) => {
  await page.goto('/')
  await page.waitForURL(/#\/login/)
  // App needs time to mount Vue components (backend connection attempt)
  await page.waitForSelector('a:has-text("Registrieren")', { timeout: 10000 })
  // Use force:true because the offline-indicator backdrop overlays the page when no backend is running
  await page.click('a:has-text("Registrieren")', { force: true })
  await expect(page.locator('[data-test="register-btn"]')).toBeVisible()
})

test('can switch back to login form', async ({ page }) => {
  await page.goto('/')
  await page.waitForURL(/#\/login/)
  await page.waitForSelector('a:has-text("Registrieren")', { timeout: 10000 })
  await page.click('a:has-text("Registrieren")', { force: true })
  await page.click('a:has-text("Anmelden")', { force: true })
  await expect(page.locator('[data-test="login-btn"]')).toBeVisible()
})

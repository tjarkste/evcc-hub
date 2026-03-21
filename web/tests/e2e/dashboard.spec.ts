import { test, expect } from '@playwright/test'

// Register a real user via the backend API and return auth data
async function registerAndAuth(baseURL: string) {
  const email = `dashboard-test-${Date.now()}@example.com`
  const resp = await fetch(`${baseURL}/api/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password: 'testpassword123' }),
  })
  return resp.json()
}

async function setTestAuth(page: any, baseURL: string) {
  const auth = await registerAndAuth(baseURL)
  await page.addInitScript((authData: any) => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify(authData))
  }, auth)
}

test('unauthenticated user is redirected to login', async ({ page }) => {
  await page.goto('/')
  await expect(page).toHaveURL(/#\/login/)
})

test('login page shows email and password fields', async ({ page }) => {
  await page.goto('/')
  await page.waitForURL(/#\/login/)
  await expect(page.locator('[data-test="email"]')).toBeVisible({ timeout: 10000 })
  await expect(page.locator('[data-test="password"]')).toBeVisible()
  await expect(page.locator('[data-test="login-btn"]')).toBeVisible()
})

test('authenticated user sees dashboard', async ({ page, baseURL }) => {
  await setTestAuth(page, baseURL!)
  await page.goto('/')
  await expect(page).not.toHaveURL(/#\/login/)
})

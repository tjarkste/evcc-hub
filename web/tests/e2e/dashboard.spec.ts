import { test, expect } from '@playwright/test'

// Hilfsfunktion: Setzt Auth-Daten direkt in localStorage (umgeht echten Login-Flow im Test)
async function setTestAuth(page: any) {
  await page.addInitScript(() => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify({
      token: 'test-token',
      mqttUsername: 'user_test_user_1',
      mqttPassword: 'test-password',
      topicPrefix: 'user/test_user_1/evcc',
    }))
  })
}

test('unauthenticated user is redirected to login', async ({ page }) => {
  // Kein Auth in localStorage — Router leitet zu /#/login weiter
  await page.goto('/')
  // Hash-based routing: URL enthält #/login
  await expect(page).toHaveURL(/#\/login/)
})

test('login page shows email and password fields', async ({ page }) => {
  await page.goto('/')
  await page.waitForURL(/#\/login/)
  // App needs time to mount Vue components (backend connection attempt)
  await expect(page.locator('[data-test="email"]')).toBeVisible({ timeout: 10000 })
  await expect(page.locator('[data-test="password"]')).toBeVisible()
  await expect(page.locator('[data-test="login-btn"]')).toBeVisible()
})

test('authenticated user sees dashboard', async ({ page }) => {
  await setTestAuth(page)
  await page.goto('/')
  // Sollte NICHT auf /login redirecten
  await expect(page).not.toHaveURL(/#\/login/)
})

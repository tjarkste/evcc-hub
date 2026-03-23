import { test, expect } from '@playwright/test'

// Registration response matches AuthData in services/auth.ts:
// { token, refreshToken, mqttUsername, mqttPassword, userId, defaultSite? }
// Backend auto-creates one "My Home" site on register (Phase 1 behaviour).
async function registerUser(baseURL: string) {
  const email = `phase4-${Date.now()}@example.com`
  const resp = await fetch(`${baseURL}/api/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password: 'testpassword123' }),
  })
  return resp.json() as Promise<{
    token: string
    refreshToken: string
    mqttUsername: string
    mqttPassword: string
    userId: string
    defaultSite?: { id: string; name: string; mqttUsername: string; mqttPassword: string; topicPrefix: string }
  }>
}

async function createSite(baseURL: string, token: string, name: string) {
  const resp = await fetch(`${baseURL}/api/sites`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ name }),
  })
  return resp.json() as Promise<{ id: string; name: string; topicPrefix: string }>
}

// ── Onboarding ────────────────────────────────────────────────────────────────

test('onboarding shows three numbered steps after register', async ({ page, baseURL }) => {
  await page.goto('/#/login')
  await page.click('text=Noch kein Konto?')
  await page.fill('[data-test="email"]', `onboard-${Date.now()}@example.com`)
  await page.fill('[data-test="password"]', 'testpassword123')
  await page.click('[data-test="register-btn"]')

  await expect(page.locator('text=MQTT-Konfiguration hinzufügen')).toBeVisible({ timeout: 5000 })
  await expect(page.locator('text=evcc neu starten')).toBeVisible()
  await expect(page.locator('[data-test="mqtt-config"]')).toBeVisible()
  await expect(page.locator('[data-test="to-dashboard-btn"]')).toBeVisible()
})

test('copy button in onboarding copies MQTT config', async ({ page, baseURL }) => {
  await page.goto('/#/login')
  await page.click('text=Noch kein Konto?')
  await page.fill('[data-test="email"]', `onboard2-${Date.now()}@example.com`)
  await page.fill('[data-test="password"]', 'testpassword123')
  await page.click('[data-test="register-btn"]')

  await expect(page.locator('[data-test="copy-config-btn"]')).toBeVisible({ timeout: 5000 })
  await page.click('[data-test="copy-config-btn"]')
  await expect(page.locator('text=✓ Kopiert!')).toBeVisible({ timeout: 2000 })
})

// ── Site Overview ─────────────────────────────────────────────────────────────

test('single-site user lands on dashboard, not overview', async ({ page, baseURL }) => {
  const auth = await registerUser(baseURL!)

  await page.addInitScript((authData: any) => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify(authData))
  }, auth)

  await page.goto('/#/')
  await page.waitForTimeout(1500)
  expect(page.url()).not.toContain('/overview')
})

test('multi-site user is redirected to /overview on load', async ({ page, baseURL }) => {
  const auth = await registerUser(baseURL!)
  await createSite(baseURL!, auth.token, 'Ferienhaus')

  await page.addInitScript((authData: any) => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify(authData))
  }, auth)

  await page.goto('/#/')
  // Two site cards: "My Home" (auto-created on register) + "Ferienhaus"
  await expect(page.locator('[data-testid="site-card"]')).toHaveCount(2, { timeout: 5000 })
})

test('multi-site overview shows all site names', async ({ page, baseURL }) => {
  const auth = await registerUser(baseURL!)
  await createSite(baseURL!, auth.token, 'Ferienhaus')

  await page.addInitScript((authData: any) => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify(authData))
  }, auth)

  await page.goto('/#/')
  await expect(page.locator('text=My Home')).toBeVisible({ timeout: 5000 })
  await expect(page.locator('text=Ferienhaus')).toBeVisible()
})

test('clicking Anzeigen on overview card navigates to dashboard', async ({ page, baseURL }) => {
  const auth = await registerUser(baseURL!)
  const secondSite = await createSite(baseURL!, auth.token, 'Ferienhaus')

  await page.addInitScript((authData: any) => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify(authData))
  }, auth)

  await page.goto('/#/')
  await expect(page.locator('[data-testid="site-card"]')).toHaveCount(2, { timeout: 5000 })

  await page.click(`[data-testid="view-site-${secondSite.id}"]`)
  await page.waitForTimeout(500)
  expect(page.url()).toContain('#/')
  expect(page.url()).not.toContain('/overview')
})

// ── Site Switcher ─────────────────────────────────────────────────────────────

test('site switcher is hidden for single-site user', async ({ page, baseURL }) => {
  const auth = await registerUser(baseURL!)

  await page.addInitScript((authData: any) => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify(authData))
  }, auth)

  await page.goto('/#/')
  await page.waitForTimeout(1500)
  await expect(page.locator('[data-testid="site-switcher"]')).not.toBeVisible()
})

test('site switcher is visible for multi-site user on dashboard', async ({ page, baseURL }) => {
  const auth = await registerUser(baseURL!)
  await createSite(baseURL!, auth.token, 'Ferienhaus')

  await page.addInitScript((authData: any) => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify(authData))
  }, auth)

  await page.goto('/#/')
  // Wait for overview, then navigate to dashboard
  await page.waitForSelector('[data-testid="site-card"]', { timeout: 5000 })
  await page.click('[data-testid^="view-site-"]')

  await expect(page.locator('[data-testid="site-switcher"]')).toBeVisible({ timeout: 3000 })
})

test('site switcher dropdown lists all sites', async ({ page, baseURL }) => {
  const auth = await registerUser(baseURL!)
  await createSite(baseURL!, auth.token, 'Ferienhaus')

  await page.addInitScript((authData: any) => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify(authData))
  }, auth)

  await page.goto('/#/')
  await page.waitForSelector('[data-testid="site-card"]', { timeout: 5000 })
  await page.click('[data-testid^="view-site-"]')

  await page.click('[data-testid="site-switcher-toggle"]')
  await expect(page.locator('text=My Home')).toBeVisible()
  await expect(page.locator('text=Ferienhaus')).toBeVisible()
})

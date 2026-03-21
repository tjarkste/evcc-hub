import { test, expect } from '@playwright/test'

// Helper: register a user and get auth data via API
async function registerTestUser(baseURL: string) {
  const email = `test-${Date.now()}@example.com`
  const resp = await fetch(`${baseURL}/api/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password: 'testpassword123' }),
  })
  return resp.json()
}

test('can list sites after login', async ({ page, baseURL }) => {
  const auth = await registerTestUser(baseURL!)

  await page.addInitScript((authData: any) => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify(authData))
  }, auth)

  await page.goto('/#/sites')
  // Should see at least the default "My Home" site
  await expect(page.locator('text=My Home')).toBeVisible({ timeout: 10000 })
})

test('can create a new site', async ({ page, baseURL }) => {
  const auth = await registerTestUser(baseURL!)

  await page.addInitScript((authData: any) => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify(authData))
  }, auth)

  await page.goto('/#/sites')
  await page.waitForSelector('text=My Home', { timeout: 10000 })

  await page.fill('[data-test="new-site-name"]', 'Ferienhaus')
  await page.click('[data-test="create-site-btn"]')

  // Should show the new site with MQTT config
  await expect(page.locator('text=Ferienhaus')).toBeVisible({ timeout: 5000 })
  await expect(page.locator('text=wurde erstellt')).toBeVisible()
})

test('can delete a site', async ({ page, baseURL }) => {
  const auth = await registerTestUser(baseURL!)

  // Create a second site via API so we have 2 (can't delete last one)
  await fetch(`${baseURL}/api/sites`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${auth.token}`,
    },
    body: JSON.stringify({ name: 'To Delete' }),
  })

  await page.addInitScript((authData: any) => {
    localStorage.setItem('evcc-cloud-auth', JSON.stringify(authData))
  }, auth)

  await page.goto('/#/sites')
  await page.waitForSelector('text=To Delete', { timeout: 10000 })

  // Delete the "To Delete" site (click its delete button)
  const deleteButtons = page.locator('[data-test="delete-site-btn"]')
  await deleteButtons.last().click()

  // Should no longer show "To Delete"
  await expect(page.locator('text=To Delete')).not.toBeVisible({ timeout: 5000 })
})

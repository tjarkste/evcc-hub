// web/tests/e2e/reliability.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Health endpoint', () => {
  test('returns ok with database check', async ({ request }) => {
    const resp = await request.get('/health')
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    expect(body.status).toBe('ok')
    expect(body.checks.database).toBe('ok')
  })
})

test.describe('Error response format', () => {
  test('login error includes code field', async ({ request }) => {
    const resp = await request.post('/api/auth/login', {
      data: { email: 'nonexistent@test.com', password: 'wrongpassword' },
    })
    expect(resp.status()).toBe(401)
    const body = await resp.json()
    expect(body.error).toBeDefined()
    expect(body.code).toBe('invalid_credentials')
  })

  test('register validation error includes code field', async ({ request }) => {
    const resp = await request.post('/api/auth/register', {
      data: { email: 'bad', password: '123' },
    })
    expect(resp.status()).toBe(400)
    const body = await resp.json()
    expect(body.error).toBeDefined()
    expect(body.code).toBe('invalid_input')
  })
})

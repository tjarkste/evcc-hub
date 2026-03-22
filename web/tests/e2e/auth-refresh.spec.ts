import { test, expect } from '@playwright/test'

test.describe('Auth: Refresh Token & Logout', () => {
  let email: string
  let authData: { token: string; refreshToken: string; userId: string; mqttUsername: string; mqttPassword: string }

  test.beforeEach(async ({ request, baseURL }) => {
    email = `refresh-${Date.now()}@test.de`
    const resp = await request.post(`${baseURL}/api/auth/register`, {
      data: { email, password: 'testpass123' },
    })
    expect(resp.ok()).toBeTruthy()
    authData = await resp.json()
    expect(authData.token, 'register should return access token').toBeTruthy()
    expect(authData.refreshToken, 'register should return refresh token').toBeTruthy()
  })

  test('login returns refresh token', async ({ request, baseURL }) => {
    const resp = await request.post(`${baseURL}/api/auth/login`, {
      data: { email, password: 'testpass123' },
    })
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.refreshToken, 'login should return refresh token').toBeTruthy()
  })

  test('refresh returns new tokens', async ({ request, baseURL }) => {
    const resp = await request.post(`${baseURL}/api/auth/refresh`, {
      data: { refreshToken: authData.refreshToken },
    })
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.token, 'should return new access token').toBeTruthy()
    expect(data.refreshToken, 'should return new refresh token').toBeTruthy()
    expect(data.refreshToken).not.toBe(authData.refreshToken)
  })

  test('old refresh token rejected after rotation', async ({ request, baseURL }) => {
    // Use the token once
    await request.post(`${baseURL}/api/auth/refresh`, {
      data: { refreshToken: authData.refreshToken },
    })
    // Second use with the same token should fail
    const resp = await request.post(`${baseURL}/api/auth/refresh`, {
      data: { refreshToken: authData.refreshToken },
    })
    expect(resp.status()).toBe(401)
  })

  test('logout invalidates refresh token', async ({ request, baseURL }) => {
    const logoutResp = await request.post(`${baseURL}/api/auth/logout`, {
      data: { refreshToken: authData.refreshToken },
    })
    expect(logoutResp.ok()).toBeTruthy()

    const refreshResp = await request.post(`${baseURL}/api/auth/refresh`, {
      data: { refreshToken: authData.refreshToken },
    })
    expect(refreshResp.status()).toBe(401)
  })

  test('invalid refresh token returns 401', async ({ request, baseURL }) => {
    const resp = await request.post(`${baseURL}/api/auth/refresh`, {
      data: { refreshToken: 'completely-invalid-token-xyz' },
    })
    expect(resp.status()).toBe(401)
  })
})

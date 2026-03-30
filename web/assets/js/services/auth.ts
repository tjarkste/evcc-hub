// web/assets/js/services/auth.ts
interface AuthData {
  token: string
  refreshToken: string
  mqttUsername: string
  mqttPassword: string
  userId: string
  defaultSite?: {
    id: string
    name: string
    mqttUsername: string
    mqttPassword: string
    topicPrefix: string
  }
}

const AUTH_KEY = 'evcc-cloud-auth'
let refreshTimer: ReturnType<typeof setTimeout> | null = null

export async function login(email: string, password: string): Promise<AuthData> {
  const resp = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  if (!resp.ok) throw new Error('Login fehlgeschlagen')
  const data: AuthData = await resp.json()
  localStorage.setItem(AUTH_KEY, JSON.stringify(data))
  scheduleTokenRefresh()
  return data
}

export async function register(email: string, password: string): Promise<AuthData> {
  const resp = await fetch('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  if (!resp.ok) throw new Error('Registrierung fehlgeschlagen')
  const data: AuthData = await resp.json()
  localStorage.setItem(AUTH_KEY, JSON.stringify(data))
  scheduleTokenRefresh()
  return data
}

export async function refreshAccessToken(): Promise<AuthData | null> {
  const auth = getStoredAuth()
  if (!auth?.refreshToken) return null

  try {
    const resp = await fetch('/api/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken: auth.refreshToken }),
    })
    if (!resp.ok) {
      logout()
      return null
    }
    const refreshData = await resp.json()
    // Merge new tokens with existing auth data (keep MQTT creds etc.)
    const updated: AuthData = {
      ...auth,
      token: refreshData.token,
      refreshToken: refreshData.refreshToken,
    }
    localStorage.setItem(AUTH_KEY, JSON.stringify(updated))
    scheduleTokenRefresh()
    return updated
  } catch {
    return null
  }
}

// Schedule a refresh 1 minute before the 15-minute access token expires (at 14 min).
export function scheduleTokenRefresh(): void {
  if (refreshTimer) clearTimeout(refreshTimer)
  refreshTimer = setTimeout(() => {
    refreshAccessToken()
  }, 14 * 60 * 1000)
}

export function stopTokenRefresh(): void {
  if (refreshTimer) {
    clearTimeout(refreshTimer)
    refreshTimer = null
  }
}

export function getStoredAuth(): AuthData | null {
  try {
    const raw = localStorage.getItem(AUTH_KEY)
    if (!raw) return null
    return JSON.parse(raw)
  } catch {
    return null
  }
}

export function getAuthToken(): string | null {
  const auth = getStoredAuth()
  return auth?.token ?? null
}

export async function logout(): Promise<void> {
  const auth = getStoredAuth()
  if (auth?.refreshToken) {
    fetch('/api/auth/logout', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken: auth.refreshToken }),
    }).catch(() => {})
  }
  stopTokenRefresh()
  localStorage.removeItem(AUTH_KEY)
  localStorage.removeItem('evcc-cloud-selected-site')
  localStorage.removeItem('evcc-cloud-state-cache-v2')
  localStorage.removeItem('evcc-cloud-cached-topic-prefix')

  // Disconnect the MQTT client so the next user doesn't inherit this session's
  // live data stream. Import lazily to avoid a circular dependency at module load.
  const { disconnectMqtt } = await import('./mqtt')
  disconnectMqtt()

  // Clear in-memory store state so the next login starts clean.
  const { default: store } = await import('../store')
  store.reset()
  store.state.lastDataAt = null
}

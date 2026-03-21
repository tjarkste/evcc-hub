// web/assets/js/services/auth.ts
interface AuthData {
  token: string
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

export async function login(email: string, password: string): Promise<AuthData> {
  const resp = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  if (!resp.ok) throw new Error('Login fehlgeschlagen')
  const data: AuthData = await resp.json()
  localStorage.setItem(AUTH_KEY, JSON.stringify(data))
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
  return data
}

export function getStoredAuth(): AuthData | null {
  const stored = localStorage.getItem(AUTH_KEY)
  return stored ? JSON.parse(stored) : null
}

export function getAuthToken(): string | null {
  const auth = getStoredAuth()
  return auth?.token ?? null
}

export function logout() {
  localStorage.removeItem(AUTH_KEY)
  localStorage.removeItem('evcc-cloud-selected-site')
}

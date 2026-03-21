// web/assets/js/services/sites.ts
import { getAuthToken } from './auth'

export interface Site {
  id: string
  userId: string
  name: string
  mqttUsername: string
  mqttPassword?: string
  topicPrefix: string
  timezone: string | null
  createdAt: string
  updatedAt: string
}

const SELECTED_SITE_KEY = 'evcc-cloud-selected-site'

function authHeaders(): HeadersInit {
  const token = getAuthToken()
  return {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`,
  }
}

export async function fetchSites(): Promise<Site[]> {
  const resp = await fetch('/api/sites', { headers: authHeaders() })
  if (!resp.ok) throw new Error('Could not fetch sites')
  const data = await resp.json()
  return data.sites
}

export async function createSite(name: string, timezone?: string): Promise<Site> {
  const resp = await fetch('/api/sites', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ name, timezone }),
  })
  if (!resp.ok) throw new Error('Could not create site')
  const data = await resp.json()
  return data.site
}

export async function updateSite(id: string, name: string): Promise<Site> {
  const resp = await fetch(`/api/sites/${id}`, {
    method: 'PUT',
    headers: authHeaders(),
    body: JSON.stringify({ name }),
  })
  if (!resp.ok) throw new Error('Could not update site')
  const data = await resp.json()
  return data.site
}

export async function deleteSite(id: string): Promise<void> {
  const resp = await fetch(`/api/sites/${id}`, {
    method: 'DELETE',
    headers: authHeaders(),
  })
  if (!resp.ok) throw new Error('Could not delete site')
}

export function getSelectedSiteId(): string | null {
  return localStorage.getItem(SELECTED_SITE_KEY)
}

export function setSelectedSiteId(siteId: string): void {
  localStorage.setItem(SELECTED_SITE_KEY, siteId)
}

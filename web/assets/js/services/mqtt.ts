// web/assets/js/services/mqtt.ts
import mqtt from 'mqtt'
import type { MqttClient } from 'mqtt'
import store from '../store'
import { ConnectionState } from '../types/evcc'
import { saveStateCache } from './stateCache'

interface MqttConfig {
  brokerUrl: string
  username: string
  password: string
}

let client: MqttClient | null = null
let currentTopicPrefix: string = ''
let reconnectAttempt = 0
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let savedConfig: MqttConfig | null = null

const MAX_BACKOFF_MS = 30_000
const BASE_DELAY_MS = 1_000

/**
 * Calculate reconnect delay with exponential backoff and jitter.
 * Formula: min(base * 2^attempt, cap) + random jitter (0–50% of delay)
 */
export function calculateBackoff(attempt: number): number {
  const exponential = Math.min(BASE_DELAY_MS * Math.pow(2, attempt), MAX_BACKOFF_MS)
  const jitter = Math.random() * exponential * 0.5
  return exponential + jitter
}

function setConnectionState(state: ConnectionState): void {
  store.state.connectionState = state
  store.state.offline = state !== ConnectionState.CONNECTED
}

function scheduleReconnect(): void {
  if (reconnectTimer) clearTimeout(reconnectTimer)
  if (!savedConfig) return

  setConnectionState(ConnectionState.RECONNECTING)
  const delay = calculateBackoff(reconnectAttempt)
  console.log(`MQTT reconnecting in ${Math.round(delay)}ms (attempt ${reconnectAttempt + 1})`)

  reconnectTimer = setTimeout(() => {
    reconnectAttempt++
    if (client) {
      client.reconnect()
    }
  }, delay)
}

export function connectMqtt(config: MqttConfig): MqttClient {
  savedConfig = config
  reconnectAttempt = 0

  setConnectionState(ConnectionState.RECONNECTING)

  client = mqtt.connect(config.brokerUrl, {
    username: config.username,
    password: config.password,
    protocolVersion: 4,
    reconnectPeriod: 0, // disable built-in reconnect — we manage it
  })

  client.on('connect', () => {
    console.log('MQTT connected')
    reconnectAttempt = 0
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    setConnectionState(ConnectionState.CONNECTED)

    // Re-subscribe to active site
    if (currentTopicPrefix) {
      client!.subscribe(`${currentTopicPrefix}/#`)
    }
  })

  client.on('message', (topic, payload) => {
    store.state.lastDataAt = Date.now()
    const storeUpdate = mqttToStoreUpdate(topic, payload.toString(), currentTopicPrefix)
    if (storeUpdate) {
      store.update(storeUpdate)
      saveStateCache(store.state)
    }
  })

  client.on('offline', () => {
    if (savedConfig) {
      scheduleReconnect()
    }
  })

  client.on('close', () => {
    if (savedConfig) {
      scheduleReconnect()
    }
  })

  client.on('error', (err) => {
    console.warn('MQTT error:', err.message)
  })

  return client
}

const CACHED_TOPIC_KEY = 'evcc-cloud-cached-topic-prefix'

export function subscribeSite(topicPrefix: string): void {
  if (!client) return

  if (currentTopicPrefix) {
    client.unsubscribe(`${currentTopicPrefix}/#`)
  }

  store.reset()
  currentTopicPrefix = topicPrefix
  localStorage.setItem(CACHED_TOPIC_KEY, topicPrefix)
  client.subscribe(`${topicPrefix}/#`)
}

export function getCachedTopicPrefix(): string | null {
  return localStorage.getItem(CACHED_TOPIC_KEY)
}

export function disconnectMqtt() {
  savedConfig = null
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  client?.end()
  client = null
  currentTopicPrefix = ''
  setConnectionState(ConnectionState.OFFLINE)
}

export function publishCommand(topicSuffix: string, value: string) {
  if (!client) throw new Error('MQTT not connected')
  client.publish(`${currentTopicPrefix}/${topicSuffix}`, value)
}

export function getTopicPrefix(): string {
  return currentTopicPrefix
}

/**
 * Konvertiert eine MQTT-Message in ein Store-Update-Objekt.
 */
export function mqttToStoreUpdate(
  topic: string,
  payload: string,
  prefix: string
): Record<string, unknown> | null {
  if (!prefix || !topic.startsWith(prefix + '/')) return null
  const relative = topic.slice(prefix.length + 1)
  const parts = relative.split('/')

  if (parts[parts.length - 1] === 'set') return null

  let storeKey: string

  if (parts[0] === 'loadpoints' && parts.length >= 3) {
    const mqttIndex = parseInt(parts[1])
    const storeIndex = mqttIndex - 1
    const field = parts.slice(2).join('.')
    storeKey = `loadpoints.${storeIndex}.${field}`
  } else if (parts[0] === 'site' && parts.length >= 2) {
    storeKey = parts.slice(1).join('.')
  } else {
    storeKey = parts.join('.')
  }

  try {
    return { [storeKey]: parsePayload(payload) }
  } catch (e) {
    console.warn(`Malformed MQTT payload on ${topic}:`, payload, e)
    return null
  }
}

export function parsePayload(raw: string): unknown {
  if (raw === 'true') return true
  if (raw === 'false') return false
  if (raw === '' || raw === 'nil' || raw === 'null') return null
  const num = Number(raw)
  if (!isNaN(num)) return num
  if (raw.startsWith('[') || raw.startsWith('{')) {
    try { return JSON.parse(raw) } catch {}
  }
  return raw
}

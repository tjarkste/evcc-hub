// web/assets/js/services/mqtt.ts
import mqtt from 'mqtt'
import type { MqttClient } from 'mqtt'
import store from '../store'

interface MqttConfig {
  brokerUrl: string
  username: string
  password: string
}

let client: MqttClient | null = null
let currentTopicPrefix: string = ''

export function connectMqtt(config: MqttConfig): MqttClient {
  client = mqtt.connect(config.brokerUrl, {
    username: config.username,
    password: config.password,
    protocolVersion: 4,
    reconnectPeriod: 2500,
  })

  client.on('connect', () => {
    store.state.offline = false
    // Subscribe to active site if set
    if (currentTopicPrefix) {
      client!.subscribe(`${currentTopicPrefix}/#`)
    }
  })

  client.on('message', (topic, payload) => {
    const storeUpdate = mqttToStoreUpdate(topic, payload.toString(), currentTopicPrefix)
    if (storeUpdate) {
      store.update(storeUpdate)
    }
  })

  client.on('offline', () => {
    store.state.offline = true
  })

  client.on('close', () => {
    store.state.offline = true
  })

  return client
}

export function subscribeSite(topicPrefix: string): void {
  if (!client) return

  // Unsubscribe from old site
  if (currentTopicPrefix) {
    client.unsubscribe(`${currentTopicPrefix}/#`)
  }

  // Reset store state for new site
  store.reset()

  // Subscribe to new site
  currentTopicPrefix = topicPrefix
  client.subscribe(`${topicPrefix}/#`)
}

export function disconnectMqtt() {
  client?.end()
  client = null
  currentTopicPrefix = ''
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
 *
 * MQTT-Topic:  user/abc/site/s1/evcc/site/pvPower       -> Store-Key: "pvPower"
 * MQTT-Topic:  user/abc/site/s1/evcc/loadpoints/1/mode   -> Store-Key: "loadpoints.0.mode"
 *
 * Kritisch: Loadpoint-Index 1-basiert (MQTT) -> 0-basiert (Store-Array)
 */
export function mqttToStoreUpdate(
  topic: string,
  payload: string,
  prefix: string
): Record<string, unknown> | null {
  if (!prefix || !topic.startsWith(prefix + '/')) return null
  const relative = topic.slice(prefix.length + 1)
  const parts = relative.split('/')

  // /set Topics ignorieren (sind Befehle, keine State-Updates)
  if (parts[parts.length - 1] === 'set') return null

  let storeKey: string

  if (parts[0] === 'loadpoints' && parts.length >= 3) {
    // loadpoints/1/mode -> loadpoints.0.mode (Index 1->0)
    const mqttIndex = parseInt(parts[1])
    const storeIndex = mqttIndex - 1
    const field = parts.slice(2).join('.')
    storeKey = `loadpoints.${storeIndex}.${field}`
  } else if (parts[0] === 'site' && parts.length >= 2) {
    // site/pvPower -> pvPower (evcc store is flat for site-level data)
    storeKey = parts.slice(1).join('.')
  } else {
    // startupCompleted -> startupCompleted
    storeKey = parts.join('.')
  }

  return { [storeKey]: parsePayload(payload) }
}

export function parsePayload(raw: string): unknown {
  if (raw === 'true') return true
  if (raw === 'false') return false
  if (raw === '' || raw === 'nil' || raw === 'null') return null
  const num = Number(raw)
  return isNaN(num) ? raw : num
}

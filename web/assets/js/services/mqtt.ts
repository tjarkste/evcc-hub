// web/assets/js/services/mqtt.ts
import mqtt from 'mqtt'
import type { MqttClient } from 'mqtt'
import store from '../store'

interface MqttConfig {
  brokerUrl: string      // wss://mqtt.evcc-cloud.de/mqtt
  username: string       // MQTT-Username
  password: string       // MQTT-Passwort
  topicPrefix: string    // user/abc123/evcc
}

let client: MqttClient | null = null
let currentTopicPrefix: string = ''

export function connectMqtt(config: MqttConfig): MqttClient {
  currentTopicPrefix = config.topicPrefix
  client = mqtt.connect(config.brokerUrl, {
    username: config.username,
    password: config.password,
    protocolVersion: 4,
    reconnectPeriod: 2500,
  })

  client.on('connect', () => {
    store.state.offline = false
    client!.subscribe(`${config.topicPrefix}/#`)
  })

  client.on('message', (topic, payload) => {
    const storeUpdate = mqttToStoreUpdate(topic, payload.toString(), config.topicPrefix)
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
 * MQTT-Topic:  user/abc/evcc/site/pvPower       → Store-Key: "site.pvPower"
 * MQTT-Topic:  user/abc/evcc/loadpoints/1/mode   → Store-Key: "loadpoints.0.mode"
 *
 * Kritisch: Loadpoint-Index 1-basiert (MQTT) → 0-basiert (Store-Array)
 */
export function mqttToStoreUpdate(
  topic: string,
  payload: string,
  prefix: string
): Record<string, unknown> | null {
  const relative = topic.replace(`${prefix}/`, '')
  const parts = relative.split('/')

  // /set Topics ignorieren (sind Befehle, keine State-Updates)
  if (parts[parts.length - 1] === 'set') return null

  let storeKey: string

  if (parts[0] === 'loadpoints' && parts.length >= 3) {
    // loadpoints/1/mode → loadpoints.0.mode (Index 1→0)
    const mqttIndex = parseInt(parts[1])
    const storeIndex = mqttIndex - 1
    const field = parts.slice(2).join('.')
    storeKey = `loadpoints.${storeIndex}.${field}`
  } else if (parts[0] === 'site' && parts.length >= 2) {
    // site/pvPower → pvPower (evcc store is flat for site-level data)
    storeKey = parts.slice(1).join('.')
  } else {
    // startupCompleted → startupCompleted
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

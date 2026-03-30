// web/assets/js/api.ts — Ersetzt axios durch MQTT-Publish
// Das Interface bleibt kompatibel, damit keine Komponenten geändert werden müssen.

import { publishCommand } from './services/mqtt'

interface MqttMapping {
  topic: string
  payload: string
}

/**
 * Konvertiert REST-API-Pfade in MQTT-Topics.
 *
 * POST loadpoints/1/mode/pv          → loadpoints/1/mode/set, "pv"
 * POST loadpoints/1/limitsoc/80      → loadpoints/1/limitSoc/set, "80"
 * POST loadpoints/1/minsoc/20        → loadpoints/1/minSoc/set, "20"
 * POST loadpoints/1/phases/3         → loadpoints/1/phasesConfigured/set, "3"
 * POST site/batterymode/hold         → site/batteryMode/set, "hold"
 * POST prioritysoc/10                → site/prioritySoc/set, "10"
 * POST buffersoc/80                  → site/bufferSoc/set, "80"
 * POST bufferstartsoc/50             → site/bufferStartSoc/set, "50"
 * POST batterydischargecontrol/true  → site/batteryDischargeControl/set, "true"
 */
export function restPathToMqttTopic(path: string): MqttMapping | null {
  // loadpoints/{id}/mode/{value}
  let match = path.match(/^loadpoints\/(\d+)\/mode\/(.+)$/)
  if (match) return { topic: `loadpoints/${match[1]}/mode/set`, payload: match[2] }

  // loadpoints/{id}/limitsoc/{value}
  match = path.match(/^loadpoints\/(\d+)\/limitsoc\/(.+)$/i)
  if (match) return { topic: `loadpoints/${match[1]}/limitSoc/set`, payload: match[2] }

  // loadpoints/{id}/minsoc/{value}
  match = path.match(/^loadpoints\/(\d+)\/minsoc\/(.+)$/i)
  if (match) return { topic: `loadpoints/${match[1]}/minSoc/set`, payload: match[2] }

  // loadpoints/{id}/phases/{value}
  match = path.match(/^loadpoints\/(\d+)\/phases\/(.+)$/)
  if (match) return { topic: `loadpoints/${match[1]}/phasesConfigured/set`, payload: match[2] }

  // site/batterymode/{value}
  match = path.match(/^site\/batterymode\/(.+)$/i)
  if (match) return { topic: `site/batteryMode/set`, payload: match[1] }

  // site/smartcostlimit/{value}
  match = path.match(/^site\/smartcostlimit\/(.+)$/i)
  if (match) return { topic: `site/smartCostLimit/set`, payload: match[1] }

  // prioritysoc/{value}
  match = path.match(/^prioritysoc\/(.+)$/i)
  if (match) return { topic: `site/prioritySoc/set`, payload: match[1] }

  // buffersoc/{value}
  match = path.match(/^buffersoc\/(.+)$/i)
  if (match) return { topic: `site/bufferSoc/set`, payload: match[1] }

  // bufferstartsoc/{value}
  match = path.match(/^bufferstartsoc\/(.+)$/i)
  if (match) return { topic: `site/bufferStartSoc/set`, payload: match[1] }

  // batterydischargecontrol/{value}
  match = path.match(/^batterydischargecontrol\/(.+)$/i)
  if (match) return { topic: `site/batteryDischargeControl/set`, payload: match[1] }

  return null
}

// Kompatibilitäts-Wrapper: Fängt die gleichen Pfade ab wie die REST-API
const api = {
  post(path: string, _data?: unknown): Promise<{ data: unknown }> {
    const mqttTopic = restPathToMqttTopic(path)
    if (!mqttTopic) {
      console.warn(`Kein MQTT-Äquivalent für: ${path}`)
      return Promise.resolve({ data: {} })
    }
    publishCommand(mqttTopic.topic, mqttTopic.payload)
    return Promise.resolve({ data: {} })
  },
  // GET-Requests: Nicht benötigt (kein Config, kein Log im Cloud-MVP)
  get(_path: string): Promise<{ data: unknown }> {
    return Promise.resolve({ data: {} })
  },
}

export default api

// Named Exports für Kompatibilität mit Komponenten, die baseApi, i18n,
// allowClientError oder downloadFile importieren.
// Im Cloud-MVP werden diese Funktionen nicht aktiv genutzt,
// aber fehlende Exporte würden den Build brechen.

export const baseApi = {
  get(_path: string): Promise<{ data: unknown }> {
    return Promise.resolve({ data: {} })
  },
  post(_path: string, _data?: unknown): Promise<{ data: unknown }> {
    return Promise.resolve({ data: {} })
  },
}

export const i18n = {
  get(_path: string): Promise<{ data: unknown }> {
    return Promise.resolve({ data: {} })
  },
}

export const allowClientError = {
  validateStatus(status: number) {
    return status >= 200 && status < 500
  },
}

export function downloadFile(_res: { status: number; data: unknown; headers: Record<string, string> }): void {
  // Im Cloud-MVP nicht implementiert
  console.warn('downloadFile ist im Cloud-MVP nicht verfügbar')
}

// scripts/simulator.js
// Simuliert einen evcc-Installaton via MQTT.
// Publiziert feste Zustandswerte und reagiert auf /set-Befehle.
const mqtt = require('mqtt')

const broker = process.env.MQTT_BROKER || 'mqtt://localhost:1883'
const prefix = process.env.MQTT_TOPIC || 'user/test_user_1/evcc'
const interval = parseInt(process.env.INTERVAL || '2000')

// Feste Werte — deterministisch, Tests können exakt prüfen
const state = {
  'site/pvPower': 6200,
  'site/gridPower': -1800,
  'site/batteryPower': 1500,
  'site/batterySoc': 72,
  'site/homePower': 2900,
  'site/batteryMode': 'normal',
  'loadpoints/1/chargePower': 7400,
  'loadpoints/1/vehicleSoc': 64,
  'loadpoints/1/mode': 'pv',
  'loadpoints/1/charging': 'true',
  'loadpoints/1/connected': 'true',
  'loadpoints/1/phasesActive': 3,
  'loadpoints/1/chargedEnergy': 12300,
  'loadpoints/1/vehicleRange': 210,
  'loadpoints/1/title': 'Garage',
  'loadpoints/1/enabled': 'true',
  'loadpoints/1/chargeDuration': 8100,
}

const client = mqtt.connect(broker)

client.on('connect', () => {
  console.log(`Simulator connected to ${broker}, publishing on ${prefix}`)
  client.subscribe(`${prefix}/+/+/set`)
  client.subscribe(`${prefix}/+/set`)

  setInterval(() => {
    for (const [topic, value] of Object.entries(state)) {
      client.publish(`${prefix}/${topic}`, String(value), { retain: true })
    }
  }, interval)
})

client.on('message', (topic, payload) => {
  if (!topic.endsWith('/set')) return
  const stateTopic = topic.replace(`${prefix}/`, '').replace('/set', '')
  state[stateTopic] = payload.toString()
  client.publish(`${prefix}/${stateTopic}`, payload, { retain: true })
  console.log(`Set: ${stateTopic} = ${payload}`)
})

client.on('error', (err) => {
  console.error('MQTT error:', err.message)
})

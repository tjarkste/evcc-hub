import { describe, test, expect } from 'vitest'
import { mqttToStoreUpdate, parsePayload, calculateBackoff } from './mqtt'

describe('mqttToStoreUpdate', () => {
  test('site/pvPower maps to flat pvPower', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/site/pvPower', '6200', 'user/abc/evcc'
    )
    expect(result).toEqual({ pvPower: 6200 })
  })

  test('site/homePower maps to flat homePower', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/site/homePower', '1200', 'user/abc/evcc'
    )
    expect(result).toEqual({ homePower: 1200 })
  })

  test('site/gridPower remaps to grid.power (nested)', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/site/gridPower', '-800', 'user/abc/evcc'
    )
    expect(result).toEqual({ 'grid.power': -800 })
  })

  test('site/batteryPower remaps to battery.power (nested)', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/site/batteryPower', '500', 'user/abc/evcc'
    )
    expect(result).toEqual({ 'battery.power': 500 })
  })

  test('site/batterySoc remaps to battery.soc (nested)', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/site/batterySoc', '72', 'user/abc/evcc'
    )
    expect(result).toEqual({ 'battery.soc': 72 })
  })

  test('site/batteryCapacity remaps to battery.capacity (nested)', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/site/batteryCapacity', '10000', 'user/abc/evcc'
    )
    expect(result).toEqual({ 'battery.capacity': 10000 })
  })

  test('site/batteryMode stays flat (not remapped)', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/site/batteryMode', 'normal', 'user/abc/evcc'
    )
    expect(result).toEqual({ batteryMode: 'normal' })
  })

  test('loadpoint index shifts from 1-based to 0-based', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/loadpoints/1/mode', 'pv', 'user/abc/evcc'
    )
    expect(result).toEqual({ 'loadpoints.0.mode': 'pv' })
  })

  test('loadpoint 2 maps to index 1', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/loadpoints/2/chargePower', '3600', 'user/abc/evcc'
    )
    expect(result).toEqual({ 'loadpoints.1.chargePower': 3600 })
  })

  test('pv/1/power shifts to pv.0.power (0-indexed)', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/pv/1/power', '5000', 'user/abc/evcc'
    )
    expect(result).toEqual({ 'pv.0.power': 5000 })
  })

  test('pv/2/power shifts to pv.1.power', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/pv/2/power', '2500', 'user/abc/evcc'
    )
    expect(result).toEqual({ 'pv.1.power': 2500 })
  })

  test('battery/1/power maps to battery.devices.0.power', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/battery/1/power', '1000', 'user/abc/evcc'
    )
    expect(result).toEqual({ 'battery.devices.0.power': 1000 })
  })

  test('battery/1/soc maps to battery.devices.0.soc', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/battery/1/soc', '80', 'user/abc/evcc'
    )
    expect(result).toEqual({ 'battery.devices.0.soc': 80 })
  })

  test('aux/1/power shifts to aux.0.power', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/aux/1/power', '300', 'user/abc/evcc'
    )
    expect(result).toEqual({ 'aux.0.power': 300 })
  })

  test('/set topics are ignored', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/loadpoints/1/mode/set', 'pv', 'user/abc/evcc'
    )
    expect(result).toBeNull()
  })

  test('pv/1 count topic (no subfield) is ignored', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/pv/1', '1', 'user/abc/evcc'
    )
    expect(result).toBeNull()
  })

  test('loadpoints/1 count topic (no subfield) is ignored', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/loadpoints/1', '1', 'user/abc/evcc'
    )
    expect(result).toBeNull()
  })

  test('battery/1 count topic (no subfield) is ignored', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/battery/1', '1', 'user/abc/evcc'
    )
    expect(result).toBeNull()
  })
})

describe('calculateBackoff', () => {
  test('first attempt is ~1000ms', () => {
    const delay = calculateBackoff(0)
    expect(delay).toBeGreaterThanOrEqual(1000)
    expect(delay).toBeLessThanOrEqual(1500)
  })

  test('second attempt is ~2000ms', () => {
    const delay = calculateBackoff(1)
    expect(delay).toBeGreaterThanOrEqual(2000)
    expect(delay).toBeLessThanOrEqual(3000)
  })

  test('caps at 30000ms', () => {
    const delay = calculateBackoff(10)
    expect(delay).toBeGreaterThanOrEqual(30000)
    expect(delay).toBeLessThanOrEqual(45000)
  })

  test('third attempt is ~4000ms', () => {
    const delay = calculateBackoff(2)
    expect(delay).toBeGreaterThanOrEqual(4000)
    expect(delay).toBeLessThanOrEqual(6000)
  })
})

describe('malformed payload handling', () => {
  test('returns null and does not throw for topics with malformed prefix', () => {
    const result = mqttToStoreUpdate('completely/wrong/topic', '123', 'user/abc/evcc')
    expect(result).toBeNull()
  })
})

describe('parsePayload', () => {
  test('boolean true', () => {
    expect(parsePayload('true')).toBe(true)
  })

  test('boolean false', () => {
    expect(parsePayload('false')).toBe(false)
  })

  test('number positive', () => {
    expect(parsePayload('6200')).toBe(6200)
  })

  test('number negative', () => {
    expect(parsePayload('-1800')).toBe(-1800)
  })

  test('string stays string', () => {
    expect(parsePayload('pv')).toBe('pv')
  })

  test('empty string returns null', () => {
    expect(parsePayload('')).toBeNull()
  })

  test('null string returns null', () => {
    expect(parsePayload('null')).toBeNull()
  })
})

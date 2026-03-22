import { describe, test, expect } from 'vitest'
import { mqttToStoreUpdate, parsePayload, calculateBackoff } from './mqtt'

describe('mqttToStoreUpdate', () => {
  test('site topic maps correctly', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/site/pvPower', '6200', 'user/abc/evcc'
    )
    expect(result).toEqual({ pvPower: 6200 })
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

  test('/set topics are ignored', () => {
    const result = mqttToStoreUpdate(
      'user/abc/evcc/loadpoints/1/mode/set', 'pv', 'user/abc/evcc'
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

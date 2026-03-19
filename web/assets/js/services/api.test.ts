import { describe, test, expect } from 'vitest'
import { restPathToMqttTopic } from '../api'

describe('restPathToMqttTopic', () => {
  test('mode change', () => {
    expect(restPathToMqttTopic('loadpoints/1/mode/pv')).toEqual({
      topic: 'loadpoints/1/mode/set',
      payload: 'pv',
    })
  })

  test('limitsoc', () => {
    expect(restPathToMqttTopic('loadpoints/1/limitsoc/80')).toEqual({
      topic: 'loadpoints/1/limitSoc/set',
      payload: '80',
    })
  })

  test('minsoc', () => {
    expect(restPathToMqttTopic('loadpoints/1/minsoc/20')).toEqual({
      topic: 'loadpoints/1/minSoc/set',
      payload: '20',
    })
  })

  test('phases', () => {
    expect(restPathToMqttTopic('loadpoints/1/phases/3')).toEqual({
      topic: 'loadpoints/1/phasesConfigured/set',
      payload: '3',
    })
  })

  test('battery mode', () => {
    expect(restPathToMqttTopic('site/batterymode/hold')).toEqual({
      topic: 'site/batteryMode/set',
      payload: 'hold',
    })
  })

  test('unsupported path returns null', () => {
    expect(restPathToMqttTopic('loadpoints/1/vehicle/mycar')).toBeNull()
  })

  test('unknown path returns null', () => {
    expect(restPathToMqttTopic('sessions')).toBeNull()
  })
})

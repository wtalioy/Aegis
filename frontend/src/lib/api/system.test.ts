import { beforeEach, describe, expect, it, vi } from 'vitest'

const requestJSON = vi.fn()

vi.mock('../http', () => ({
  requestJSON
}))

describe('lib/api/system', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.useFakeTimers()
    requestJSON.mockReset()
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  it('calls the v1 system stats and alerts endpoints', async () => {
    requestJSON
      .mockResolvedValueOnce({ processCount: 1 })
      .mockResolvedValueOnce([{ id: 'alert-1' }])

    const mod = await import('./system')

    await expect(mod.getSystemStats()).resolves.toEqual({ processCount: 1 })
    await expect(mod.getAlerts()).resolves.toEqual([{ id: 'alert-1' }])
    expect(requestJSON).toHaveBeenNthCalledWith(1, '/api/v1/system/stats')
    expect(requestJSON).toHaveBeenNthCalledWith(2, '/api/v1/system/alerts')
  })

  it('shares one polling interval across subscribers and fans out alert updates', async () => {
    requestJSON.mockResolvedValue([{ id: 'alert-1' }])
    const setIntervalSpy = vi.spyOn(window, 'setInterval')
    const clearIntervalSpy = vi.spyOn(window, 'clearInterval')

    const mod = await import('./system')
    const listenerA = vi.fn()
    const listenerB = vi.fn()

    const unsubscribeA = mod.subscribeToAlerts(listenerA)
    const unsubscribeB = mod.subscribeToAlerts(listenerB)

    expect(setIntervalSpy).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(2000)

    expect(requestJSON).toHaveBeenCalledTimes(1)
    expect(listenerA).toHaveBeenCalledWith([{ id: 'alert-1' }])
    expect(listenerB).toHaveBeenCalledWith([{ id: 'alert-1' }])

    unsubscribeA()
    expect(clearIntervalSpy).not.toHaveBeenCalled()

    unsubscribeB()
    expect(clearIntervalSpy).toHaveBeenCalledTimes(1)
  })

  it('keeps polling after a failed fetch', async () => {
    requestJSON
      .mockRejectedValueOnce(new Error('boom'))
      .mockResolvedValueOnce([{ id: 'alert-2' }])

    const mod = await import('./system')
    const listener = vi.fn()
    const unsubscribe = mod.subscribeToAlerts(listener)

    await vi.advanceTimersByTimeAsync(2000)
    expect(listener).not.toHaveBeenCalled()
    expect(console.error).toHaveBeenCalled()

    await vi.advanceTimersByTimeAsync(2000)
    expect(listener).toHaveBeenCalledWith([{ id: 'alert-2' }])

    unsubscribe()
  })
})

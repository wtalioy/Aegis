import { describe, expect, it, vi, beforeEach } from 'vitest'

import { requestJSON } from './http'

describe('lib/http requestJSON', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  it('returns null for empty successful responses', async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      text: vi.fn().mockResolvedValue('')
    } as unknown as Response)

    await expect(requestJSON('/api/test')).resolves.toBeNull()
  })

  it('returns parsed JSON for successful responses', async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      text: vi.fn().mockResolvedValue('{"ok":true}')
    } as unknown as Response)

    await expect(requestJSON<{ ok: boolean }>('/api/test')).resolves.toEqual({ ok: true })
  })

  it('prefers error, then message, then HTTP status text on failures', async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce({
        ok: false,
        status: 400,
        text: vi.fn().mockResolvedValue('{"error":"broken"}')
      } as unknown as Response)
      .mockResolvedValueOnce({
        ok: false,
        status: 403,
        text: vi.fn().mockResolvedValue('{"message":"denied"}')
      } as unknown as Response)
      .mockResolvedValueOnce({
        ok: false,
        status: 503,
        text: vi.fn().mockResolvedValue('')
      } as unknown as Response)

    await expect(requestJSON('/api/test')).rejects.toThrow('broken')
    await expect(requestJSON('/api/test')).rejects.toThrow('denied')
    await expect(requestJSON('/api/test')).rejects.toThrow('HTTP 503')
  })
})

import { beforeEach, describe, expect, it, vi } from 'vitest'

const { queryEvents } = vi.hoisted(() => ({
  queryEvents: vi.fn()
}))

vi.mock('./useAI', () => ({
  useAI: () => ({
    explainEvent: vi.fn(),
    analyzeContext: vi.fn()
  })
}))

vi.mock('../lib/api', () => ({
  queryEvents
}))

import { useInvestigation } from './useInvestigation'

const execEvent = {
  id: 'exec-1',
  type: 'exec' as const,
  timestamp: 300,
  pid: 10,
  cgroupId: '1',
  processName: 'bash',
  parentComm: 'init',
  filename: '/bin/bash',
  commandLine: 'bash',
  blocked: false
}

const fileEvent = {
  id: 'file-1',
  type: 'file' as const,
  timestamp: 200,
  pid: 11,
  cgroupId: '1',
  processName: 'cat',
  filename: '/tmp/one',
  flags: 0,
  blocked: false
}

const connectEvent = {
  id: 'net-1',
  type: 'connect' as const,
  timestamp: 100,
  pid: 12,
  cgroupId: '1',
  processName: 'curl',
  family: 2,
  port: 443,
  addr: '10.0.0.5',
  blocked: false
}

describe('useInvestigation', () => {
  beforeEach(() => {
    queryEvents.mockReset()
  })

  it('tracks pagination state and hasMore fallback for full pages', async () => {
    const investigation = useInvestigation()
    queryEvents.mockResolvedValue({
      events: [execEvent, fileEvent],
      total: 2,
      page: 1,
      limit: 2,
      totalPages: 0,
      typeCounts: { exec: 1, file: 1, connect: 0 }
    })

    const result = await investigation.searchEvents({ page: 1, limit: 2 })

    expect(result?.events).toHaveLength(2)
    expect(investigation.state.value.events).toHaveLength(2)
    expect(investigation.typeCounts.value).toEqual({ exec: 1, file: 1, connect: 0 })
    expect(investigation.hasMore.value).toBe(true)
  })

  it('dedupes load-more results against existing events', async () => {
    const investigation = useInvestigation()
    queryEvents
      .mockResolvedValueOnce({
        events: [execEvent, fileEvent],
        total: 3,
        page: 1,
        limit: 2,
        totalPages: 2,
        typeCounts: { exec: 1, file: 1, connect: 1 }
      })
      .mockResolvedValueOnce({
        events: [fileEvent, connectEvent],
        total: 3,
        page: 2,
        limit: 2,
        totalPages: 2,
        typeCounts: { exec: 1, file: 1, connect: 1 }
      })

    await investigation.searchEvents({ page: 1, limit: 2 })
    const pageTwo = await investigation.loadMoreEvents()

    expect(pageTwo?.events).toEqual([fileEvent, connectEvent])
    expect(investigation.state.value.events.map((event) => event.id)).toEqual(['exec-1', 'file-1', 'net-1'])
    expect(investigation.hasMore.value).toBe(false)
  })

  it('refreshes by replacing page one, deduping older pages, and capping history', async () => {
    const investigation = useInvestigation()
    queryEvents
      .mockResolvedValueOnce({
        events: [execEvent, fileEvent],
        total: 2,
        page: 1,
        limit: 2,
        totalPages: 1,
        typeCounts: { exec: 1, file: 1, connect: 0 }
      })
      .mockResolvedValueOnce({
      events: [
        { ...execEvent, id: 'new-1', timestamp: 5000 },
        { ...fileEvent, id: 'new-2', timestamp: 4000 }
      ],
      total: 2052,
      page: 1,
      limit: 2,
      totalPages: 1026,
      typeCounts: { exec: 1, file: 1, connect: 0 }
      })

    await investigation.searchEvents({ page: 1, limit: 2 })
    investigation.state.value.events = [
      { ...execEvent, id: 'page1-old', timestamp: 3000 },
      { ...fileEvent, id: 'page1-old-2', timestamp: 2000 },
      ...Array.from({ length: 2050 }, (_, index) => ({
        ...connectEvent,
        id: `older-${index}`,
        timestamp: 1500 - index
      }))
    ]

    await investigation.refreshEvents()

    expect(investigation.state.value.events[0].id).toBe('new-1')
    expect(investigation.state.value.events[1].id).toBe('new-2')
    expect(investigation.state.value.events).toHaveLength(2000)
    expect(new Set(investigation.state.value.events.map((event) => event.id)).size).toBe(2000)
  })

  it('returns null and stores an error when search fails', async () => {
    const investigation = useInvestigation()
    queryEvents.mockRejectedValue(new Error('search failed'))

    await expect(investigation.searchEvents({ page: 1, limit: 10 })).resolves.toBeNull()
    expect(investigation.error.value).toBe('search failed')
  })
})

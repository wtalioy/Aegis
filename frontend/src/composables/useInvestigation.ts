// Investigation Composable - Phase 4
import { ref, computed } from 'vue'
import { useAI } from './useAI'
import type { QueryRequest, QueryResponse, EventType } from '../types/events'

export interface InvestigationState {
  query: string
  filters: {
    types: EventType[]
    processes: string[]
    pids: number[]
    timeWindow?: {
      start: Date
      end: Date
    }
  }
  events: any[]
  selectedEvent: any | null
  aiContext: any | null
}

const API_BASE = '/api'

export function useInvestigation() {
  const { explainEvent, analyzeContext } = useAI()

  const state = ref<InvestigationState>({
    query: '',
    filters: {
      types: [] as EventType[],
      processes: [],
      pids: []
    },
    events: [],
    selectedEvent: null,
    aiContext: null
  })

  const loading = ref(false)
  const loadingMore = ref(false)
  const error = ref<string | null>(null)

  // pagination state
  const currentPage = ref(1)
  const currentLimit = ref(50)
  const total = ref(0)
  const totalPages = ref(0)
  const lastBatchCount = ref(0)
  const typeCounts = ref({ exec: 0, file: 0, connect: 0 })
  const hasMore = computed(() => {
    if (totalPages.value > 0) {
      return currentPage.value < totalPages.value
    }
    // Fallback when backend doesn't return totals: 
    // If last page was full (exactly equal to limit), assume more pages might exist
    // But if we got fewer events than requested, we've reached the end
    const limit = currentLimit.value || 50
    return lastBatchCount.value >= limit && lastBatchCount.value > 0
  })
  const lastQuery = ref<QueryRequest | null>(null)

  const hasFilters = computed(() => {
    return state.value.filters.types.length > 0 ||
      state.value.filters.processes.length > 0 ||
      state.value.filters.pids.length > 0
  })

  // Shared normalization function: converts backend flat structure to frontend header structure
  const normalizeEvent = (ev: any, index: number): any => {
    const timestamp = ev.timestamp || ev.header?.timestamp || Date.now()
    const pid = ev.pid || ev.header?.pid || 0
    const comm = ev.comm || ev.processName || ev.header?.comm || 'Unknown'
    const cgroupId = ev.cgroupId || ev.header?.cgroupId || ''
    const type = ev.type || 'unknown'
    const id = ev.id || `${type}-${timestamp}-${pid}-${index}`

    const normalized: any = {
      id,
      type,
      header: {
        timestamp,
        pid,
        cgroupId,
        comm,
        ...(ev.ppid !== undefined && { ppid: ev.ppid }),
        ...(ev.header?.ppid !== undefined && { ppid: ev.header.ppid })
      },
      blocked: ev.blocked || false
    }

    if (type === 'exec') {
      normalized.parentComm = ev.parentComm || ''
      normalized.filename = ev.filename || ''
      normalized.commandLine = ev.commandLine || ev.filename || comm
    } else if (type === 'file') {
      normalized.filename = ev.filename || ''
      normalized.flags = ev.flags || 0
      if (ev.ino !== undefined) normalized.ino = ev.ino
      if (ev.dev !== undefined) normalized.dev = ev.dev
    } else if (type === 'connect') {
      normalized.family = ev.family || 0
      normalized.port = ev.port || 0
      normalized.addr = ev.addr || ''
    }

    return normalized
  }

  const searchEvents = async (query?: QueryRequest): Promise<QueryResponse | null> => {
    loading.value = true
    error.value = null

    // remember this query (used by load more)
    const effective: QueryRequest = query ?? {
      filter: {
        types: state.value.filters.types,
        processes: state.value.filters.processes,
        pids: state.value.filters.pids,
        timeWindow: state.value.filters.timeWindow ? {
          start: state.value.filters.timeWindow.start.toISOString(),
          end: state.value.filters.timeWindow.end.toISOString()
        } : undefined
      },
      page: 1,
      limit: currentLimit.value
    }
    lastQuery.value = effective

    try {
      const endpoint = '/api/query'
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(effective)
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      const data = await response.json()

      // Normalize backend events to frontend format
      const events = Array.isArray(data.events) ? data.events.map(normalizeEvent) : []

      // Replace current events and update pagination
      state.value.events = events
      lastBatchCount.value = events.length
      total.value = Number(data.total ?? events.length)
      currentPage.value = Number(data.page ?? 1)
      currentLimit.value = Number(data.limit ?? currentLimit.value)
      totalPages.value = Number(data.totalPages ?? Math.ceil(total.value / (currentLimit.value || 1)))

      // Update type counts from backend if provided
      if (data.type_counts) {
        typeCounts.value = {
          exec: Number(data.type_counts.exec ?? 0),
          file: Number(data.type_counts.file ?? 0),
          connect: Number(data.type_counts.connect ?? 0)
        }
      }

      return { ...data, events: state.value.events }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to search events'
      return null
    } finally {
      loading.value = false
    }
  }


  const explainSelectedEvent = async (question?: string) => {
    if (!state.value.selectedEvent) return null

    return await explainEvent({
      eventId: state.value.selectedEvent.id,
      eventData: state.value.selectedEvent,
      question
    })
  }

  const analyzeProcess = async (pid: number) => {
    return await analyzeContext({
      type: 'process',
      id: pid.toString()
    })
  }

  const setSelectedEvent = (event: any) => {
    state.value.selectedEvent = event
    state.value.aiContext = null
  }

  const clearFilters = () => {
    state.value.filters = {
      types: [],
      processes: [],
      pids: []
    }
  }

  const loadMoreEvents = async (): Promise<QueryResponse | null> => {
    if (loadingMore.value || !hasMore.value) return null
    loadingMore.value = true
    error.value = null

    try {
      const base = lastQuery.value || { page: 1, limit: currentLimit.value }
      const nextReq: QueryRequest = { ...base, page: (currentPage.value + 1), limit: currentLimit.value }
      const endpoint = '/api/query'
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(nextReq)
      })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)

      const data = await response.json()

      // Normalize backend events to frontend format
      const newEvents = Array.isArray(data.events) ? data.events.map(normalizeEvent) : []

      const existing = new Set(state.value.events.map((e: any) => e.id))
      const uniqueNew = newEvents.filter((e: any) => !existing.has(e.id))
      state.value.events = state.value.events.concat(uniqueNew)

      total.value = Number(data.total ?? total.value)
      currentPage.value = Number(data.page ?? currentPage.value + 1)
      currentLimit.value = Number(data.limit ?? currentLimit.value)
      totalPages.value = Number(data.totalPages ?? Math.ceil(total.value / (currentLimit.value || 1)))

      // Update type counts from backend if provided
      if (data.type_counts) {
        typeCounts.value = {
          exec: Number(data.type_counts.exec ?? 0),
          file: Number(data.type_counts.file ?? 0),
          connect: Number(data.type_counts.connect ?? 0)
        }
      }

      // update lastQuery to reflect the most recent request
      lastQuery.value = { ...(lastQuery.value || {}), page: currentPage.value, limit: currentLimit.value }
      return { ...data, events: newEvents }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load more events'
      return null
    } finally {
      loadingMore.value = false
    }
  }

  const refreshEvents = async (): Promise<void> => {
    // Don't refresh if already loading.
    // Note: we *still* refresh while "hasMore" is true (Load More visible),
    // otherwise the list appears "stuck" until the user clicks Load More.
    if (loading.value || loadingMore.value) {
      return
    }

    try {
      const base = lastQuery.value || {
        filter: {
          types: state.value.filters.types,
          processes: state.value.filters.processes,
          pids: state.value.filters.pids
        },
        page: 1,
        limit: currentLimit.value
      }

      const refreshReq: QueryRequest = { ...base, page: 1, limit: currentLimit.value }
      const endpoint = '/api/query'
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(refreshReq)
      })

      if (!response.ok) return

      const data = await response.json()

      // Normalize backend events to frontend format
      const newEvents = Array.isArray(data.events) ? data.events.map(normalizeEvent) : []

      // Merge strategy: always keep page-1 fresh, even when Load More is visible.
      //
      // Problem: when the user has loaded multiple pages, `state.value.events` contains
      // more than the first page. The backend refresh only returns page 1.
      // If we only "prepend unique" events, page-1 items that shifted out (because new
      // events arrived) won't be removed, which makes the list look stuck until user
      // clicks Load More (which re-queries deeper pages and changes the view).
      //
      // Solution: treat the server response as the authoritative page-1 window and
      // replace the existing page-1 segment, while keeping older pages appended.
      const limit = currentLimit.value || 50
      const page1 = newEvents
      const older = state.value.events.slice(limit)

      // Dedupe older against new page1
      const page1Ids = new Set(page1.map((e: any) => e.id))
      const olderDeduped = older.filter((e: any) => !page1Ids.has(e.id))

      state.value.events = [...page1, ...olderDeduped]

      // Sort by timestamp (newest first) and limit to prevent memory issues
      state.value.events.sort((a: any, b: any) => {
        return (b.header?.timestamp || 0) - (a.header?.timestamp || 0)
      })

      // Keep latest 2000 events
      if (state.value.events.length > 2000) {
        state.value.events = state.value.events.slice(0, 2000)
      }

      // Update pagination state based on the latest query
      lastBatchCount.value = newEvents.length
      total.value = Number(data.total ?? total.value)
      currentPage.value = Number(data.page ?? 1)
      currentLimit.value = Number(data.limit ?? currentLimit.value)
      totalPages.value = Number(data.totalPages ?? Math.ceil(total.value / (currentLimit.value || 1)))

      // Update type counts from backend if provided
      if (data.type_counts) {
        typeCounts.value = {
          exec: Number(data.type_counts.exec ?? 0),
          file: Number(data.type_counts.file ?? 0),
          connect: Number(data.type_counts.connect ?? 0)
        }
      }
    } catch (err) {
      // Silently fail on refresh errors to avoid disrupting user experience
      console.error('Failed to refresh events:', err)
    }
  }

  return {
    state,
    loading,
    loadingMore,
    hasMore,
    error,
    hasFilters,
    typeCounts,
    searchEvents,
    loadMoreEvents,
    refreshEvents,
    explainSelectedEvent,
    analyzeProcess,
    setSelectedEvent,
    clearFilters
  }
}


import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useInvestigation } from './useInvestigation'
import type { SecurityEvent } from '../types/events'

export function useInvestigationPage() {
  const {
    state,
    searchEvents,
    loading,
    loadMoreEvents,
    hasMore,
    loadingMore,
    refreshEvents,
    typeCounts
  } = useInvestigation()
  const route = useRoute()

  const filterType = ref<string>('all')
  const searchQuery = ref(typeof route.query.search === 'string' ? route.query.search : '')
  const sortBy = ref<'time' | 'pid' | 'type' | 'process'>('time')
  const sortDir = ref<'asc' | 'desc'>('desc')

  let refreshInterval: ReturnType<typeof setInterval> | null = null

  const stopAutoRefresh = () => {
    if (refreshInterval) {
      clearInterval(refreshInterval)
      refreshInterval = null
    }
  }

  const startAutoRefresh = () => {
    stopAutoRefresh()
    refreshInterval = setInterval(async () => {
      if (!loading.value && !searchQuery.value.trim()) {
        await refreshEvents()
      }
    }, 5000)
  }

  const handleEventSelect = async (event: SecurityEvent) => {
    state.value.selectedEvent = event
  }

  const filteredEvents = computed(() => {
    let result = state.value.events

    if (filterType.value !== 'all') {
      result = result.filter((event) => event.type === filterType.value)
    }

    if (searchQuery.value.trim()) {
      const query = searchQuery.value.toLowerCase()
      result = result.filter((event) =>
        (event.processName || '').toLowerCase().includes(query) ||
        String(event.pid || '').includes(query) ||
        (event.type || '').toLowerCase().includes(query)
      )
    }

    return result
  })

  const eventTypeCounts = computed(() => typeCounts.value)

  const sortedEvents = computed(() => {
    const events = [...filteredEvents.value]
    const direction = sortDir.value === 'asc' ? 1 : -1
    const compareStrings = (left: string, right: string) => left.localeCompare(right) * direction
    const compareNumbers = (left: number, right: number) => ((left ?? 0) - (right ?? 0)) * direction

    switch (sortBy.value) {
      case 'time':
        events.sort((left, right) => compareNumbers(left.timestamp || 0, right.timestamp || 0))
        break
      case 'pid':
        events.sort((left, right) => compareNumbers(left.pid || 0, right.pid || 0))
        break
      case 'type':
        events.sort((left, right) => compareStrings(left.type || '', right.type || ''))
        break
      case 'process':
        events.sort((left, right) => compareStrings(left.processName || '', right.processName || ''))
        break
    }

    return events
  })

  const changeSort = (field: 'time' | 'pid' | 'type' | 'process') => {
    if (sortBy.value === field) {
      sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
      return
    }

    sortBy.value = field
    sortDir.value = field === 'time' || field === 'pid' ? 'desc' : 'asc'
  }

  onMounted(async () => {
    await searchEvents({
      filter: {
        types: [],
        processes: [],
        pids: []
      },
      page: 1,
      limit: 50
    })
    startAutoRefresh()
  })

  onUnmounted(() => {
    stopAutoRefresh()
  })

  return {
    state,
    loading,
    loadMoreEvents,
    hasMore,
    loadingMore,
    filterType,
    searchQuery,
    sortBy,
    sortDir,
    eventTypeCounts,
    sortedEvents,
    handleEventSelect,
    changeSort
  }
}

import { defineComponent, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const searchEvents = vi.fn()
const refreshEvents = vi.fn()
const loadMoreEvents = vi.fn()

const state = ref({
  events: [
    {
      id: '1',
      type: 'exec',
      timestamp: 300,
      pid: 40,
      cgroupId: '1',
      processName: 'zsh',
      parentComm: 'login',
      filename: '/bin/zsh',
      commandLine: 'zsh',
      blocked: false
    },
    {
      id: '2',
      type: 'file',
      timestamp: 200,
      pid: 20,
      cgroupId: '1',
      processName: 'cat',
      filename: '/tmp/one',
      flags: 0,
      blocked: false
    },
    {
      id: '3',
      type: 'connect',
      timestamp: 100,
      pid: 30,
      cgroupId: '1',
      processName: 'curl',
      family: 2,
      port: 443,
      addr: '10.0.0.5',
      blocked: false
    }
  ],
  selectedEvent: null
})

const loading = ref(false)
const hasMore = ref(true)
const loadingMore = ref(false)
const typeCounts = ref({ exec: 1, file: 1, connect: 1 })
const routeQuery = ref<Record<string, unknown>>({})
const baseEvents = [
  {
    id: '1',
    type: 'exec',
    timestamp: 300,
    pid: 40,
    cgroupId: '1',
    processName: 'zsh',
    parentComm: 'login',
    filename: '/bin/zsh',
    commandLine: 'zsh',
    blocked: false
  },
  {
    id: '2',
    type: 'file',
    timestamp: 200,
    pid: 20,
    cgroupId: '1',
    processName: 'cat',
    filename: '/tmp/one',
    flags: 0,
    blocked: false
  },
  {
    id: '3',
    type: 'connect',
    timestamp: 100,
    pid: 30,
    cgroupId: '1',
    processName: 'curl',
    family: 2,
    port: 443,
    addr: '10.0.0.5',
    blocked: false
  }
]

vi.mock('./useInvestigation', () => ({
  useInvestigation: () => ({
    state,
    searchEvents,
    loading,
    loadMoreEvents,
    hasMore,
    loadingMore,
    refreshEvents,
    typeCounts
  })
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({
    query: routeQuery.value
  })
}))

import { useInvestigationPage } from './useInvestigationPage'

const Harness = defineComponent({
  setup() {
    return useInvestigationPage()
  },
  template: '<div />'
})

describe('useInvestigationPage', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    searchEvents.mockReset().mockResolvedValue(null)
    refreshEvents.mockReset().mockResolvedValue(undefined)
    loadMoreEvents.mockReset().mockResolvedValue(null)
    routeQuery.value = {}
    loading.value = false
    hasMore.value = true
    loadingMore.value = false
    typeCounts.value = { exec: 1, file: 1, connect: 1 }
    state.value = {
      events: [...baseEvents],
      selectedEvent: null
    }
  })

  it('loads events on mount and seeds the search query from the route', async () => {
    routeQuery.value = { search: 'curl' }

    const wrapper = mount(Harness)
    await nextTick()

    expect(searchEvents).toHaveBeenCalledWith({
      filter: { types: [], processes: [], pids: [] },
      page: 1,
      limit: 50
    })
    expect(wrapper.vm.searchQuery).toBe('curl')
  })

  it('auto-refreshes only when idle and the search query is empty', async () => {
    const wrapper = mount(Harness)
    await nextTick()

    await vi.advanceTimersByTimeAsync(5000)
    expect(refreshEvents).toHaveBeenCalledTimes(1)

    wrapper.vm.searchQuery = 'busy'
    await nextTick()
    await vi.advanceTimersByTimeAsync(5000)
    expect(refreshEvents).toHaveBeenCalledTimes(1)

    wrapper.vm.searchQuery = ''
    loading.value = true
    await vi.advanceTimersByTimeAsync(5000)
    expect(refreshEvents).toHaveBeenCalledTimes(1)

    wrapper.unmount()
  })

  it('filters, searches, and sorts the derived event list', async () => {
    const wrapper = mount(Harness)
    await nextTick()

    wrapper.vm.filterType = 'file'
    await nextTick()
    expect(wrapper.vm.sortedEvents.map((event: { id: string }) => event.id)).toEqual(['2'])

    wrapper.vm.filterType = 'all'
    wrapper.vm.searchQuery = 'curl'
    await nextTick()
    expect(wrapper.vm.sortedEvents.map((event: { id: string }) => event.id)).toEqual(['3'])

    wrapper.vm.searchQuery = ''
    wrapper.vm.changeSort('process')
    await nextTick()
    expect(wrapper.vm.sortBy).toBe('process')
    expect(wrapper.vm.sortDir).toBe('asc')
    expect(wrapper.vm.sortedEvents.map((event: { processName: string }) => event.processName)).toEqual(['cat', 'curl', 'zsh'])

    wrapper.vm.changeSort('process')
    await nextTick()
    expect(wrapper.vm.sortDir).toBe('desc')
  })
})

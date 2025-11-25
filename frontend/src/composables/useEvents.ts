import { ref, onMounted, onUnmounted } from 'vue'
import { subscribeToEventRates, type EventRates } from '../lib/api'

export interface ExecEvent {
    type: 'exec'
    timestamp: number
    pid: number
    ppid: number
    cgroupId: string
    comm: string
    parentComm: string
    inContainer: boolean
}

export interface ConnectEvent {
    type: 'connect'
    timestamp: number
    pid: number
    cgroupId: string
    family: number
    port: number
    addr: string
    inContainer: boolean
}

export interface FileEvent {
    type: 'file'
    timestamp: number
    pid: number
    cgroupId: string
    flags: number
    filename: string
    inContainer: boolean
}

export type EulerGuardEvent = ExecEvent | ConnectEvent | FileEvent
export { EventRates }

export function useEvents(maxBufferSize = 1000) {
    const events = ref<EulerGuardEvent[]>([])
    const eventRate = ref<EventRates>({ exec: 0, network: 0, file: 0 })
    const isPaused = ref(false)
    const totalEvents = ref({ exec: 0, network: 0, file: 0 })

    let unsubscribe: UnsubscribeFn | null = null

    type UnsubscribeFn = () => void

    onMounted(() => {
        unsubscribe = subscribeToEventRates((rates) => {
            eventRate.value = rates
            totalEvents.value.exec += rates.exec
            totalEvents.value.network += rates.network
            totalEvents.value.file += rates.file
        })
    })

    onUnmounted(() => {
        unsubscribe?.()
    })

    return {
        events,
        eventRate,
        isPaused,
        totalEvents,
        pause: () => { isPaused.value = true },
        resume: () => { isPaused.value = false },
        clear: () => { events.value = [] }
    }
}

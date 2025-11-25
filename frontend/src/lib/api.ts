// API Abstraction Layer
// Supports both Wails native mode and Web server mode

export interface SystemStats {
    processCount: number
    containerCount: number
    eventsPerSec: number
    alertCount: number
    probeStatus: string
}

export interface Alert {
    id: string
    timestamp: number
    severity: string
    ruleName: string
    description: string
    pid: number
    processName: string
    inContainer: boolean
}

export interface EventRates {
    exec: number
    network: number
    file: number
}

type EventCallback<T> = (data: T) => void
type UnsubscribeFn = () => void

export const isWailsMode = typeof (window as any).__wails__ !== 'undefined'
    || typeof (window as any).go !== 'undefined'

export async function getSystemStats(): Promise<SystemStats> {
    if (isWailsMode) {
        const { GetSystemStats } = await import('../../wailsjs/go/gui/App')
        return GetSystemStats()
    }
    const resp = await fetch('/api/stats')
    return resp.json()
}

// Get alerts list
export async function getAlerts(): Promise<Alert[]> {
    if (isWailsMode) {
        const { GetAlerts } = await import('../../wailsjs/go/gui/App')
        return GetAlerts()
    }
    const resp = await fetch('/api/alerts')
    return resp.json()
}

export function subscribeToEventRates(callback: EventCallback<EventRates>): UnsubscribeFn {
    if (isWailsMode) {
        let cleanup: (() => void) | null = null

        import('../../wailsjs/runtime/runtime').then(({ EventsOn, EventsOff }) => {
            EventsOn('stats:rate', callback)
            cleanup = () => EventsOff('stats:rate')
        })

        return () => cleanup?.()
    }

    const eventSource = new EventSource('/api/events')

    eventSource.onmessage = (event) => {
        try {
            callback(JSON.parse(event.data))
        } catch (e) {
            console.error('Failed to parse SSE data:', e)
        }
    }

    return () => eventSource.close()
}

let alertPollingInterval: number | null = null
const alertListeners: Set<EventCallback<Alert[]>> = new Set()

export function subscribeToAlerts(callback: EventCallback<Alert[]>): UnsubscribeFn {
    if (isWailsMode) {
        let cleanup: (() => void) | null = null

        import('../../wailsjs/runtime/runtime').then(({ EventsOn, EventsOff }) => {
            const handleNewAlert = async () => {
                const alerts = await getAlerts()
                callback(alerts)
            }
            EventsOn('alert:new', handleNewAlert)
            cleanup = () => EventsOff('alert:new')
        })

        return () => cleanup?.()
    }

    alertListeners.add(callback)

    if (alertPollingInterval === null) {
        alertPollingInterval = window.setInterval(async () => {
            try {
                const alerts = await getAlerts()
                alertListeners.forEach(cb => cb(alerts))
            } catch (e) {
                console.error('Failed to fetch alerts:', e)
            }
        }, 2000)
    }

    return () => {
        alertListeners.delete(callback)
        if (alertListeners.size === 0 && alertPollingInterval !== null) {
            clearInterval(alertPollingInterval)
            alertPollingInterval = null
        }
    }
}

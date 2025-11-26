// API Abstraction Layer
// Supports both Wails native mode and Web server mode

export interface SystemStats {
    processCount: number
    workloadCount: number
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
    cgroupId: string
}

export interface EventRates {
    exec: number
    network: number
    file: number
}

export interface ProcessInfo {
    pid: number
    ppid: number
    comm: string
    cgroupId: string
    timestamp: number
}

export interface ExecEvent {
    type: 'exec'
    timestamp: number
    pid: number
    ppid: number
    cgroupId: string
    comm: string
    parentComm: string
}

export interface ConnectEvent {
    type: 'connect'
    timestamp: number
    pid: number
    cgroupId: string
    family: number
    port: number
    addr: string
}

export interface FileEvent {
    type: 'file'
    timestamp: number
    pid: number
    cgroupId: string
    flags: number
    filename: string
}

export type StreamEvent = ExecEvent | ConnectEvent | FileEvent

export interface DetectionRule {
    name: string
    description: string
    severity: string
    action: string
    type: 'exec' | 'file' | 'connect'
    match: Record<string, string>
    yaml: string
}

export interface LearningStatus {
    active: boolean
    startTime: number
    duration: number
    patternCount: number
    remainingSeconds: number
}

export interface GeneratedRule {
    name: string
    description: string
    severity: string
    action: string
    yaml: string
    selected: boolean
}

export interface ProbeStats {
    id: string
    name: string
    tracepoint: string
    active: boolean
    eventsRate: number
    totalCount: number
}

export interface Workload {
    id: string
    cgroupPath: string
    execCount: number
    fileCount: number
    connectCount: number
    alertCount: number
    firstSeen: number
    lastSeen: number
}

type EventCallback<T> = (data: T) => void
type UnsubscribeFn = () => void

export const isWailsMode = typeof (window as any).__wails__ !== 'undefined'
    || typeof (window as any).go !== 'undefined'

export async function getSystemStats(): Promise<SystemStats> {
    if (isWailsMode) {
        const { GetSystemStats } = await import('../../wailsjs/go/gui/App')
        return GetSystemStats() as Promise<SystemStats>
    }
    const resp = await fetch('/api/stats')
    return resp.json()
}

export async function getAlerts(): Promise<Alert[]> {
    if (isWailsMode) {
        const { GetAlerts } = await import('../../wailsjs/go/gui/App')
        return GetAlerts()
    }
    const resp = await fetch('/api/alerts')
    return resp.json()
}

export async function getAncestors(pid: number): Promise<ProcessInfo[]> {
    if (isWailsMode) {
        const { GetAncestors } = await import('../../wailsjs/go/gui/App')
        return GetAncestors(pid)
    }
    const resp = await fetch(`/api/ancestors/${pid}`)
    return resp.json()
}

export async function getRules(): Promise<DetectionRule[]> {
    if (isWailsMode) {
        const module = await import('../../wailsjs/go/gui/App') as any
        if (typeof module.GetRules === 'function') {
            return module.GetRules()
        }
        return []
    }
    const resp = await fetch('/api/rules')
    return resp.json()
}

export async function getProbeStats(): Promise<ProbeStats[]> {
    if (isWailsMode) {
        const module = await import('../../wailsjs/go/gui/App') as any
        if (typeof module.GetProbeStats === 'function') {
            return module.GetProbeStats()
        }
        return []
    }
    const resp = await fetch('/api/probes/stats')
    return resp.json()
}

export async function getWorkloads(): Promise<Workload[]> {
    if (isWailsMode) {
        const module = await import('../../wailsjs/go/gui/App') as any
        if (typeof module.GetWorkloads === 'function') {
            return module.GetWorkloads()
        }
        return []
    }
    const resp = await fetch('/api/workloads')
    return resp.json()
}

export async function getWorkload(id: string): Promise<Workload | null> {
    if (isWailsMode) {
        const module = await import('../../wailsjs/go/gui/App') as any
        if (typeof module.GetWorkload === 'function') {
            return module.GetWorkload(id)
        }
        return null
    }
    const resp = await fetch(`/api/workloads/${id}`)
    if (!resp.ok) return null
    return resp.json()
}

export async function getLearningStatus(): Promise<LearningStatus> {
    if (isWailsMode) {
        const { GetLearningStatus } = await import('../../wailsjs/go/gui/App')
        return GetLearningStatus()
    }
    const resp = await fetch('/api/learning/status')
    return resp.json()
}

export async function startLearning(durationSec: number): Promise<void> {
    if (isWailsMode) {
        const { StartLearning } = await import('../../wailsjs/go/gui/App')
        return StartLearning(durationSec)
    }
    const resp = await fetch('/api/learning/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ duration: durationSec })
    })
    if (!resp.ok) {
        const text = await resp.text()
        throw new Error(text || 'Failed to start learning')
    }
}

export async function stopLearning(): Promise<GeneratedRule[]> {
    if (isWailsMode) {
        const { StopLearning } = await import('../../wailsjs/go/gui/App')
        return StopLearning()
    }
    const resp = await fetch('/api/learning/stop', { method: 'POST' })
    if (!resp.ok) {
        const text = await resp.text()
        throw new Error(text || 'Failed to stop learning')
    }
    return resp.json()
}

export async function applyWhitelistRules(ruleIndices: number[]): Promise<void> {
    if (isWailsMode) {
        const { ApplyWhitelistRules } = await import('../../wailsjs/go/gui/App')
        return ApplyWhitelistRules(ruleIndices)
    }
    const resp = await fetch('/api/learning/apply', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ indices: ruleIndices })
    })
    if (!resp.ok) {
        const text = await resp.text()
        throw new Error(text || 'Failed to apply rules')
    }
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

export function subscribeToAllEvents(callback: EventCallback<StreamEvent>): UnsubscribeFn {
    if (isWailsMode) {
        const cleanups: (() => void)[] = []

        import('../../wailsjs/runtime/runtime').then(({ EventsOn, EventsOff }) => {
            EventsOn('event:exec', (data: ExecEvent) => callback({ ...data, type: 'exec' }))
            EventsOn('event:connect', (data: ConnectEvent) => callback({ ...data, type: 'connect' }))
            EventsOn('event:file', (data: FileEvent) => callback({ ...data, type: 'file' }))

            cleanups.push(
                () => EventsOff('event:exec'),
                () => EventsOff('event:connect'),
                () => EventsOff('event:file')
            )
        })

        return () => cleanups.forEach(fn => fn())
    }

    const eventSource = new EventSource('/api/stream')

    eventSource.onmessage = (event) => {
        try {
            callback(JSON.parse(event.data))
        } catch (e) {
            console.error('Failed to parse SSE event:', e)
        }
    }

    return () => eventSource.close()
}

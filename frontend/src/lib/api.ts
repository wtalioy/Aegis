// API Abstraction Layer for the web frontend

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
    severity: string // 'critical', 'high', 'warning', 'info'
    ruleName: string
    description: string
    pid: number
    processName: string
    cgroupId: string
    action: string   // 'alert', 'block', 'allow'
    blocked: boolean // Whether the action was blocked by LSM
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
    blocked?: boolean
}

export interface ConnectEvent {
    type: 'connect'
    timestamp: number
    pid: number
    cgroupId: string
    family: number
    port: number
    addr: string
    blocked?: boolean
}

export interface FileEvent {
    type: 'file'
    timestamp: number
    pid: number
    cgroupId: string
    flags: number
    filename: string
    blocked?: boolean
}

export type StreamEvent = ExecEvent | ConnectEvent | FileEvent


export interface LearningStatus {
    active: boolean
    startTime: number
    duration: number
    patternCount: number
    execCount: number
    fileCount: number
    connectCount: number
    remainingSeconds: number
}

// Unified Rule type - used for both detection rules and generated allow rules
export interface Rule {
    name: string
    description: string
    severity: string
    action: string // 'alert' or 'allow'
    type?: 'exec' | 'file' | 'connect' // May be derived on frontend if not provided
    match?: Record<string, string>
    yaml: string
    selected?: boolean // For generated rules selection
}

// Backward compatibility aliases
export type DetectionRule = Rule
export type GeneratedRule = Rule

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
    blockedCount: number
    firstSeen: number
    lastSeen: number
}

type EventCallback<T> = (data: T) => void
type UnsubscribeFn = () => void

export async function getSystemStats(): Promise<SystemStats> {
    const resp = await fetch('/api/stats')
    return resp.json()
}

export async function getAlerts(): Promise<Alert[]> {
    const resp = await fetch('/api/alerts')
    return resp.json()
}

export async function getAncestors(pid: number): Promise<ProcessInfo[]> {
    const resp = await fetch(`/api/ancestors/${pid}`)
    return resp.json()
}

export async function getRules(): Promise<DetectionRule[]> {
    const resp = await fetch('/api/rules')
    return resp.json()
}

export async function getProbeStats(): Promise<ProbeStats[]> {
    const resp = await fetch('/api/probes/stats')
    return resp.json()
}

export async function getWorkloads(): Promise<Workload[]> {
    const resp = await fetch('/api/workloads')
    return resp.json()
}

export async function getWorkload(id: string): Promise<Workload | null> {
    const resp = await fetch(`/api/workloads/${id}`)
    if (!resp.ok) return null
    return resp.json()
}

export async function getLearningStatus(): Promise<LearningStatus> {
    const resp = await fetch('/api/learning/status')
    return resp.json()
}

export async function startLearning(durationSec: number): Promise<void> {
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
    const resp = await fetch('/api/learning/stop', { method: 'POST' })
    if (!resp.ok) {
        const text = await resp.text()
        throw new Error(text || 'Failed to stop learning')
    }
    return resp.json()
}

export async function applyWhitelistRules(ruleIndices: number[]): Promise<void> {
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

// Subscribe to rules reload events
export function subscribeToRulesReload(callback: () => void): UnsubscribeFn {
    // back-end emits a dedicated SSE event to signal reloads
    const eventSource = new EventSource('/api/stream')

    eventSource.addEventListener('rules:reload', () => {
        callback()
    })

    return () => eventSource.close()
}

// ============================================
// AI Types
// ============================================

export interface AIStatus {
    enabled: boolean
    provider: string
    isLocal: boolean
    status: 'ready' | 'unavailable' | 'disabled'
}

export interface DiagnosisResult {
    analysis: string
    snapshotSummary: string
    provider: string
    isLocal: boolean
    durationMs: number
    timestamp: number
}

export interface ChatMessage {
    role: 'user' | 'assistant' | 'system'
    content: string
    timestamp: number
}

export interface ChatResponse {
    message: string
    sessionId: string
    contextSummary: string
    provider: string
    isLocal: boolean
    durationMs: number
    timestamp: number
    messageCount: number
}

export interface AIError {
    error: string
}

// ============================================
// AI API Functions
// ============================================

export async function getAIStatus(): Promise<AIStatus> {
    const resp = await fetch('/api/ai/status')
    if (!resp.ok) {
        try {
            const error: AIError = await resp.json()
            throw new Error(error.error || 'Failed to fetch AI status')
        } catch (err) {
            const text = await resp.text()
            throw new Error(text || 'Failed to fetch AI status')
        }
    }
    return resp.json()
}

export async function diagnoseSystem(): Promise<DiagnosisResult> {
    const resp = await fetch('/api/ai/diagnose')

    if (!resp.ok) {
        try {
            const error: AIError = await resp.json()
            throw new Error(error.error || 'Diagnosis failed')
        } catch (err) {
            const text = await resp.text()
            throw new Error(text || 'Diagnosis failed')
        }
    }

    return resp.json()
}

export async function sendChatMessage(
    message: string,
    sessionId?: string
): Promise<ChatResponse> {
    const resp = await fetch('/api/ai/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message, sessionId: sessionId || '' })
    })

    if (!resp.ok) {
        const error: AIError = await resp.json()
        throw new Error(error.error || 'Chat failed')
    }

    return resp.json()
}

export interface ChatStreamToken {
    content: string
    done: boolean
    sessionId?: string
    error?: string
}

export async function sendChatMessageStream(
    message: string,
    sessionId: string,
    onToken: (token: ChatStreamToken) => void,
    onError: (error: Error) => void,
    onComplete: () => void
): Promise<void> {
    try {
        const resp = await fetch('/api/ai/chat/stream', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ message, sessionId })
        })

        if (!resp.ok) {
            const error: AIError = await resp.json()
            throw new Error(error.error || 'Chat stream failed')
        }

        const reader = resp.body?.getReader()
        if (!reader) {
            throw new Error('No response body')
        }

        const decoder = new TextDecoder()
        let buffer = ''

        while (true) {
            const { done, value } = await reader.read()

            if (done) {
                onComplete()
                break
            }

            buffer += decoder.decode(value, { stream: true })

            const lines = buffer.split('\n')
            buffer = lines.pop() || ''

            for (const line of lines) {
                if (line.startsWith('data: ')) {
                    try {
                        const data = JSON.parse(line.slice(6)) as ChatStreamToken
                        if (data.error) {
                            onError(new Error(data.error))
                            return
                        }
                        onToken(data)
                    } catch {
                        // Skip malformed lines
                    }
                }
            }
        }
    } catch (e) {
        onError(e instanceof Error ? e : new Error('Unknown error'))
    }
}

export async function getChatHistory(sessionId: string): Promise<ChatMessage[]> {
    const resp = await fetch(`/api/ai/chat/history?sessionId=${encodeURIComponent(sessionId)}`)
    if (!resp.ok) {
        try {
            const error: AIError = await resp.json()
            throw new Error(error.error || 'Failed to load chat history')
        } catch (err) {
            const text = await resp.text()
            throw new Error(text || 'Failed to load chat history')
        }
    }
    return resp.json()
}

export async function clearChatHistory(sessionId: string): Promise<void> {
    const resp = await fetch('/api/ai/chat/clear', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sessionId })
    })
    if (!resp.ok) {
        try {
            const error: AIError = await resp.json()
            throw new Error(error.error || 'Failed to clear chat history')
        } catch (err) {
            const text = await resp.text()
            throw new Error(text || 'Failed to clear chat history')
        }
    }
}

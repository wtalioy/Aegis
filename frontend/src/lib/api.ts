// API Abstraction Layer for the web frontend

const API_BASE = '/api'

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
    processName?: string
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
    ino?: number
    dev?: number
    filename: string
    blocked?: boolean
}

// Unified Rule type - used for both detection rules and generated allow rules
export interface Rule {
    name: string
    description: string
    severity: string
    action: string // 'alert' or 'allow'
    type?: 'exec' | 'file' | 'connect' // May be derived on frontend if not provided
    mode?: string // 'production', 'testing', 'draft'
    match?: Record<string, string>
    yaml: string
    selected?: boolean // For generated rules selection
}

// Backward compatibility aliases
export type DetectionRule = Rule
export type GeneratedRule = Rule

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


export async function getRules(): Promise<DetectionRule[]> {
    const resp = await fetch('/api/rules')
    return resp.json()
}

export async function createRule(rule: DetectionRule, mode: 'testing' | 'production' = 'testing'): Promise<DetectionRule> {
    const resp = await fetch('/api/rules', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ rule, mode })
    })
    if (!resp.ok) {
        const error = await resp.text()
        throw new Error(`Failed to create rule: ${error || resp.statusText}`)
    }
    const data = await resp.json()
    return data.rule
}

export async function updateRule(ruleName: string, rule: DetectionRule): Promise<DetectionRule> {
    const encodedName = encodeURIComponent(ruleName)
    const resp = await fetch(`/api/rules/${encodedName}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ rule })
    })
    if (!resp.ok) {
        const error = await resp.text()
        throw new Error(`Failed to update rule: ${error || resp.statusText}`)
    }
    const data = await resp.json()
    return data.rule
}

export async function deleteRule(ruleName: string): Promise<void> {
    const encodedName = encodeURIComponent(ruleName)
    const resp = await fetch(`/api/rules/${encodedName}`, {
        method: 'DELETE'
    })
    if (!resp.ok) {
        const error = await resp.text()
        throw new Error(`Failed to delete rule: ${error || resp.statusText}`)
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


// ============================================
// AI Types
// ============================================

export interface AIStatus {
    provider: string
    isLocal: boolean
    status: 'ready' | 'unavailable'
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

// Settings API
export interface AISettings {
    mode: 'ollama' | 'openai'
    ollama: {
        endpoint: string
        model: string
        timeout: number
    }
    openai: {
        endpoint: string
        apiKey: string
        model: string
        timeout: number
    }
}

export interface Settings {
    ai: AISettings
}

export async function getSettings(): Promise<Settings> {
    const resp = await fetch(`${API_BASE}/settings`)
    if (!resp.ok) {
        throw new Error('Failed to load settings')
    }
    return resp.json()
}

export async function updateSettings(settings: Settings): Promise<void> {
    const resp = await fetch(`${API_BASE}/settings`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(settings)
    })
    if (!resp.ok) {
        const error = await resp.json().catch(() => ({ error: 'Failed to update settings' }))
        throw new Error(error.error || 'Failed to update settings')
    }
}

// ============================================
// Rule Validation API (NEW)
// ============================================

export async function getRuleValidation(ruleId: string) {
    const response = await fetch(`${API_BASE}/rules/validation/${ruleId}`)
    if (!response.ok) throw new Error('Failed to get rule validation')
    return response.json()
}

export async function getTestingRules() {
    const response = await fetch(`${API_BASE}/rules/testing`)
    if (!response.ok) throw new Error('Failed to get testing rules')
    return response.json()
}

export async function promoteRule(ruleId: string) {
    const response = await fetch(`${API_BASE}/rules/validation/${ruleId}/promote`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }
    })
    if (!response.ok) throw new Error('Failed to promote rule')
    return response.json()
}

export async function demoteRule(ruleId: string) {
    const response = await fetch(`${API_BASE}/rules/validation/${ruleId}/demote`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }
    })
    if (!response.ok) throw new Error('Failed to demote rule')
    return response.json()
}


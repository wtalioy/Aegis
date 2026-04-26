import { requestJSON } from '../http'
import type {
  AnalyzeRequest,
  AnalyzeResponse,
  AskInsightRequest,
  AskInsightResponse,
  ExplainRequest,
  ExplainResponse,
  RuleGenRequest,
  RuleGenResponse
} from '../../types/ai'

const API_BASE = '/api/v1/analysis'

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

interface AIError {
  error: string
}

export interface ChatStreamToken {
  content: string
  done: boolean
  sessionId?: string
  error?: string
}

export async function getAIStatus(): Promise<AIStatus> {
  return requestJSON<AIStatus>(`${API_BASE}/status`)
}

export async function diagnoseSystem(): Promise<DiagnosisResult> {
  return requestJSON<DiagnosisResult>(`${API_BASE}/diagnose`)
}

export async function generateRule(req: RuleGenRequest): Promise<RuleGenResponse> {
  return requestJSON<RuleGenResponse>(`${API_BASE}/generate-rule`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req)
  })
}

export async function explainEvent(req: ExplainRequest): Promise<ExplainResponse> {
  return requestJSON<ExplainResponse>(`${API_BASE}/explain`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req)
  })
}

export async function analyzeContext(req: AnalyzeRequest): Promise<AnalyzeResponse> {
  return requestJSON<AnalyzeResponse>(`${API_BASE}/analyze`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req)
  })
}

export async function askAboutInsight(req: AskInsightRequest): Promise<AskInsightResponse> {
  return requestJSON<AskInsightResponse>('/api/v1/sentinel/ask', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req)
  })
}

export async function sendChatMessage(message: string, sessionId?: string): Promise<ChatResponse> {
  return requestJSON<ChatResponse>(`${API_BASE}/chat`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message, sessionId: sessionId || '' })
  })
}

export async function sendChatMessageStream(
  message: string,
  sessionId: string,
  onToken: (token: ChatStreamToken) => void,
  onError: (error: Error) => void,
  onComplete: () => void
): Promise<void> {
  try {
    const response = await fetch(`${API_BASE}/chat/stream`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message, sessionId })
    })

    if (!response.ok) {
      const error: AIError = await response.json()
      throw new Error(error.error || 'Chat stream failed')
    }

    const reader = response.body?.getReader()
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
        if (!line.startsWith('data: ')) {
          continue
        }
        try {
          const token = JSON.parse(line.slice(6)) as ChatStreamToken
          if (token.error) {
            onError(new Error(token.error))
            return
          }
          onToken(token)
        } catch {
          // Ignore malformed SSE lines.
        }
      }
    }
  } catch (error) {
    onError(error instanceof Error ? error : new Error('Unknown error'))
  }
}

export async function getChatHistory(sessionId: string): Promise<ChatMessage[]> {
  return requestJSON<ChatMessage[]>(`${API_BASE}/chat/history?sessionId=${encodeURIComponent(sessionId)}`)
}

export async function clearChatHistory(sessionId: string): Promise<void> {
  await requestJSON<{ success: boolean }>(`${API_BASE}/chat/clear`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ sessionId })
  })
}

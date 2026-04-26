import { requestJSON } from '../http'
import type { QueryRequest, QueryResponse } from '../../types/events'

const API_BASE = '/api/v1/events'

export interface EventRates {
  exec: number
  network: number
  file: number
}

type EventCallback<T> = (data: T) => void
type UnsubscribeFn = () => void

export async function queryEvents(query: QueryRequest): Promise<QueryResponse> {
  return requestJSON<QueryResponse>(`${API_BASE}/query`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(query)
  })
}

export function subscribeToEventRates(callback: EventCallback<EventRates>): UnsubscribeFn {
  const eventSource = new EventSource(`${API_BASE}/stream`)

  eventSource.onmessage = (event) => {
    try {
      callback(JSON.parse(event.data))
    } catch (error) {
      console.error('Failed to parse SSE data:', error)
    }
  }

  return () => eventSource.close()
}

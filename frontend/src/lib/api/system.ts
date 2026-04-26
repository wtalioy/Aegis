import { requestJSON } from '../http'

const API_BASE = '/api/v1/system'

export interface SystemStats {
  processCount: number
  workloadCount: number
  eventsPerSec: number
  alertCount: number
  probeStatus: string
  probeError?: string
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
  action: string
  blocked: boolean
}

type EventCallback<T> = (data: T) => void
type UnsubscribeFn = () => void

let alertPollingInterval: number | null = null
const alertListeners: Set<EventCallback<Alert[]>> = new Set()

export async function getSystemStats(): Promise<SystemStats> {
  return requestJSON<SystemStats>(`${API_BASE}/stats`)
}

export async function getAlerts(): Promise<Alert[]> {
  return requestJSON<Alert[]>(`${API_BASE}/alerts`)
}

export function subscribeToAlerts(callback: EventCallback<Alert[]>): UnsubscribeFn {
  alertListeners.add(callback)

  if (alertPollingInterval === null) {
    alertPollingInterval = window.setInterval(async () => {
      try {
        const alerts = await getAlerts()
        alertListeners.forEach((listener) => listener(alerts))
      } catch (error) {
        console.error('Failed to fetch alerts:', error)
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

// Event Types - Phase 4

export type EventType = 'exec' | 'file' | 'connect'

export interface ExecEvent {
  id: string
  type: 'exec'
  timestamp: number
  pid: number
  ppid?: number
  cgroupId: string
  processName: string
  parentComm: string
  filename: string
  commandLine: string
  blocked: boolean
}

export interface FileEvent {
  id: string
  type: 'file'
  timestamp: number
  pid: number
  cgroupId: string
  processName: string
  filename: string
  flags: number
  ino?: number
  dev?: number
  blocked: boolean
}

export interface ConnectEvent {
  id: string
  type: 'connect'
  timestamp: number
  pid: number
  cgroupId: string
  processName: string
  family: number
  port: number
  addr: string
  blocked: boolean
}

export type SecurityEvent = ExecEvent | FileEvent | ConnectEvent

export interface QueryFilter {
  types?: EventType[]
  processes?: string[]
  actions?: string[]
  pids?: number[]
  cgroupIds?: string[]
  timeWindow?: {
    start: string
    end: string
  }
  correlation?: boolean
}

export interface QueryRequest {
  filter?: QueryFilter
  semantic?: string
  page?: number
  limit?: number
  sortBy?: 'time' | 'relevance'
  sortOrder?: 'asc' | 'desc'
}

export interface QueryResponse {
  events: SecurityEvent[]
  total: number
  page: number
  limit: number
  totalPages: number
  typeCounts: {
    exec: number
    file: number
    connect: number
  }
}

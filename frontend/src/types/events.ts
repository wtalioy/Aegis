// Event Types - Phase 4

export type EventType = 'exec' | 'file' | 'connect'

export interface EventHeader {
  timestamp: number
  pid: number
  ppid?: number
  cgroupId: string
  uid?: number
  gid?: number
  comm: string
}

export interface ExecEvent {
  id: string
  type: 'exec'
  header: EventHeader
  parentComm: string
  filename: string
  argv0?: string
  blocked?: boolean
}

export interface FileEvent {
  id: string
  type: 'file'
  header: EventHeader
  filename: string
  flags: number
  ino?: number
  dev?: number
  blocked?: boolean
}

export interface ConnectEvent {
  id: string
  type: 'connect'
  header: EventHeader
  family: number
  port: number
  addr: string
  blocked?: boolean
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
  type_counts?: {
    exec: number
    file: number
    connect: number
  }
}


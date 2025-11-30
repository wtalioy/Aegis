<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Terminal, Globe, FileText, Database, Zap, ShieldOff, ShieldCheck } from 'lucide-vue-next'
import { probes, type ProbeInfo } from '../../data/probes'
import { subscribeToAllEvents, type ProbeStats, type StreamEvent } from '../../lib/api'

const props = defineProps<{
  probeStats: ProbeStats[]
}>()

defineEmits<{
  selectProbe: [probe: ProbeInfo]
}>()

interface RecentEvent {
  id: string
  name: string
  pid: number
  type: 'exec' | 'file' | 'connect'
  timestamp: number
  syscall: string
  blocked: boolean
}

const recentEvents = ref<RecentEvent[]>([])
const maxRecentEvents = 5

const activeFlows = ref<{ id: string; type: string; blocked: boolean; startTime: number }[]>([])
const pulsingProbes = ref<Set<string>>(new Set())

let unsubscribe: (() => void) | null = null

const getProbeIcon = (category: string) => {
  switch (category) {
    case 'process': return Terminal
    case 'network': return Globe
    case 'file': return FileText
    default: return Terminal
  }
}

const getStatsForProbe = (probeId: string): ProbeStats | undefined => {
  return props.probeStats.find(s => s.id === probeId)
}

const totalEventsPerSec = computed(() => {
  return props.probeStats.reduce((sum, s) => sum + s.eventsRate, 0)
})

const totalEvents = computed(() => {
  return props.probeStats.reduce((sum, s) => sum + s.totalCount, 0)
})

const formatNumber = (n: number): string => {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return n.toString()
}

const handleEvent = (event: StreamEvent) => {
  const id = `${event.type}-${Date.now()}-${Math.random()}`
  const blocked = event.blocked === true

  let process: RecentEvent
  let probeType: string

  if (event.type === 'exec') {
    process = {
      id,
      name: event.comm,
      pid: event.pid,
      type: 'exec',
      timestamp: event.timestamp,
      syscall: 'execve()',
      blocked
    }
    probeType = 'exec'
  } else if (event.type === 'file') {
    process = {
      id,
      name: event.filename.split('/').pop() || event.filename,
      pid: event.pid,
      type: 'file',
      timestamp: event.timestamp,
      syscall: 'open()',
      blocked
    }
    probeType = 'openat'
  } else {
    process = {
      id,
      name: `${event.addr}`,
      pid: event.pid,
      type: 'connect',
      timestamp: event.timestamp,
      syscall: 'connect()',
      blocked
    }
    probeType = 'connect'
  }

  recentEvents.value = [process, ...recentEvents.value].slice(0, maxRecentEvents)

  activeFlows.value.push({ id, type: event.type, blocked, startTime: Date.now() })

  pulsingProbes.value.add(probeType)
  setTimeout(() => {
    pulsingProbes.value.delete(probeType)
  }, 300)

  setTimeout(() => {
    activeFlows.value = activeFlows.value.filter(f => f.id !== id)
  }, 1000)
}

const getEventIcon = (type: string) => {
  switch (type) {
    case 'exec': return Terminal
    case 'file': return FileText
    case 'connect': return Globe
    default: return Zap
  }
}

const getEventColorClass = (type: string) => {
  switch (type) {
    case 'exec': return 'process'
    case 'file': return 'file'
    case 'connect': return 'network'
    default: return 'process'
  }
}

const isFlowActive = (type: string) => {
  return activeFlows.value.some(f => f.type === type)
}

const isProbePulsing = (probeId: string) => {
  return pulsingProbes.value.has(probeId)
}

onMounted(() => {
  unsubscribe = subscribeToAllEvents(handleEvent)
})

onUnmounted(() => {
  if (unsubscribe) {
    unsubscribe()
  }
})
</script>

<template>
  <div class="architecture-diagram">
    <!-- Live Activity Header -->
    <div class="live-header">
      <div class="live-indicator">
        <span class="live-dot"></span>
        <span class="live-text">LIVE</span>
      </div>
      <div class="lsm-status">
        <ShieldOff :size="14" />
        <span>LSM Active Defense</span>
      </div>
      <div class="activity-meter">
        <span class="meter-label">Throughput</span>
        <div class="meter-bar">
          <div class="meter-fill" :style="{ width: Math.min(totalEventsPerSec * 5, 100) + '%' }"></div>
        </div>
        <span class="meter-value">{{ totalEventsPerSec }}/s</span>
      </div>
    </div>

    <!-- User Space -->
    <div class="space-section user-space">
      <div class="space-label">USER SPACE</div>

      <!-- Recent Events - Live Feed -->
      <div class="events-live">
        <TransitionGroup name="event-slide">
          <div v-for="evt in recentEvents" :key="evt.id" class="event-box"
            :class="[getEventColorClass(evt.type), { blocked: evt.blocked }]">
            <div class="event-status">
              <ShieldOff v-if="evt.blocked" :size="12" class="status-icon blocked" />
              <ShieldCheck v-else :size="12" class="status-icon allowed" />
            </div>
            <component :is="getEventIcon(evt.type)" :size="16" class="event-icon" />
            <div class="event-info">
              <span class="event-name">{{ evt.name }}</span>
              <code class="event-pid">PID: {{ evt.pid }}</code>
            </div>
            <span class="event-syscall">{{ evt.syscall }}</span>
          </div>
        </TransitionGroup>

        <div v-if="recentEvents.length === 0" class="empty-events">
          <Zap :size="24" class="empty-icon" />
          <span>Waiting for events...</span>
        </div>
      </div>

      <!-- Syscall Flows -->
      <div class="syscall-flows">
        <div class="flow-lane" :class="{ active: isFlowActive('exec') }">
          <div class="flow-label">execve()</div>
          <div class="flow-track">
            <div class="flow-particle" v-for="flow in activeFlows.filter(f => f.type === 'exec')" :key="flow.id"
              :class="{ blocked: flow.blocked }"></div>
          </div>
        </div>
        <div class="flow-lane" :class="{ active: isFlowActive('file') }">
          <div class="flow-label">open()</div>
          <div class="flow-track">
            <div class="flow-particle" v-for="flow in activeFlows.filter(f => f.type === 'file')" :key="flow.id"
              :class="{ blocked: flow.blocked }"></div>
          </div>
        </div>
        <div class="flow-lane" :class="{ active: isFlowActive('connect') }">
          <div class="flow-label">connect()</div>
          <div class="flow-track">
            <div class="flow-particle" v-for="flow in activeFlows.filter(f => f.type === 'connect')" :key="flow.id"
              :class="{ blocked: flow.blocked }"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- LSM Boundary -->
    <div class="lsm-boundary">
      <div class="boundary-line">
        <div class="boundary-pulse" :class="{ active: activeFlows.length > 0 }"></div>
      </div>
      <span class="boundary-label">
        <ShieldOff :size="12" />
        LSM Decision Point
      </span>
      <div class="boundary-line">
        <div class="boundary-pulse" :class="{ active: activeFlows.length > 0 }"></div>
      </div>
    </div>

    <!-- Kernel Space -->
    <div class="space-section kernel-space" :class="{ active: activeFlows.length > 0 }">
      <div class="space-label">KERNEL SPACE</div>

      <!-- LSM Hooks -->
      <div class="hooks-row">
        <div v-for="probe in probes" :key="probe.id" class="hook-node" :class="[
          probe.category,
          {
            active: getStatsForProbe(probe.id)?.active,
            pulsing: isProbePulsing(probe.id)
          }
        ]" @click="$emit('selectProbe', probe)">
          <div class="hook-glow"></div>
          <div class="hook-indicator"></div>
          <component :is="getProbeIcon(probe.category)" :size="20" class="hook-icon" />
          <span class="hook-name">{{ probe.name }}</span>
          <code class="hook-signature">{{ probe.hook.split('/').pop() }}</code>
          <div class="hook-capability">
            <ShieldOff :size="10" />
            <span>BLOCK</span>
          </div>
          <div class="hook-stats" v-if="getStatsForProbe(probe.id)">
            <div class="stat-rate-container">
              <span class="stat-rate">{{ getStatsForProbe(probe.id)?.eventsRate || 0 }}</span>
              <span class="stat-unit">/sec</span>
            </div>
            <div class="stat-bar">
              <div class="stat-bar-fill"
                :style="{ width: Math.min((getStatsForProbe(probe.id)?.eventsRate || 0) * 10, 100) + '%' }"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Data Flow to Ring Buffer -->
      <div class="data-flow">
        <div class="flow-streams">
          <div v-for="i in 3" :key="i" class="flow-stream" :class="{ active: activeFlows.length > 0 }">
            <div class="stream-particle"></div>
          </div>
        </div>
      </div>

      <!-- Ring Buffer -->
      <div class="ring-buffer" :class="{ receiving: activeFlows.length > 0 }">
        <div class="buffer-icon-container">
          <Database :size="24" class="buffer-icon" />
          <div class="buffer-pulse"></div>
        </div>
        <div class="buffer-info">
          <span class="buffer-title">RING BUFFER</span>
          <div class="buffer-stats-row">
            <span class="buffer-count">{{ formatNumber(totalEvents) }}</span>
            <span class="buffer-label">events</span>
          </div>
        </div>
        <div class="buffer-meter">
          <div class="meter-ring">
            <svg viewBox="0 0 36 36" class="circular-chart">
              <path class="circle-bg" d="M18 2.0845
                  a 15.9155 15.9155 0 0 1 0 31.831
                  a 15.9155 15.9155 0 0 1 0 -31.831" />
              <path class="circle-fill" :stroke-dasharray="`${Math.min(totalEventsPerSec * 2, 100)}, 100`" d="M18 2.0845
                  a 15.9155 15.9155 0 0 1 0 31.831
                  a 15.9155 15.9155 0 0 1 0 -31.831" />
            </svg>
            <span class="meter-text">{{ totalEventsPerSec }}</span>
          </div>
          <span class="meter-label">/sec</span>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div class="diagram-legend">
      <div class="legend-items">
        <div class="legend-item process">
          <Terminal :size="14" />
          <span>Process Exec</span>
        </div>
        <div class="legend-item file">
          <FileText :size="14" />
          <span>File Access</span>
        </div>
        <div class="legend-item network">
          <Globe :size="14" />
          <span>Network</span>
        </div>
        <div class="legend-item blocked">
          <ShieldOff :size="14" />
          <span>Blocked</span>
        </div>
        <div class="legend-item allowed">
          <ShieldCheck :size="14" />
          <span>Allowed</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.architecture-diagram {
  background: linear-gradient(180deg, var(--bg-surface) 0%, var(--bg-elevated) 100%);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow: hidden;
}

/* Live Header */
.live-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.live-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
}

.live-dot {
  width: 8px;
  height: 8px;
  background: #ef4444;
  border-radius: 50%;
  animation: live-pulse 1s ease-in-out infinite;
}

@keyframes live-pulse {

  0%,
  100% {
    opacity: 1;
    box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7);
  }

  50% {
    opacity: 0.8;
    box-shadow: 0 0 0 8px rgba(239, 68, 68, 0);
  }
}

.live-text {
  font-size: 11px;
  font-weight: 700;
  color: #ef4444;
  letter-spacing: 0.1em;
}

.lsm-status {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  background: rgba(139, 92, 246, 0.1);
  border: 1px solid rgba(139, 92, 246, 0.3);
  border-radius: var(--radius-full);
  font-size: 10px;
  font-weight: 500;
  color: #8b5cf6;
}

.activity-meter {
  display: flex;
  align-items: center;
  gap: 12px;
}

.meter-label {
  font-size: 11px;
  color: var(--text-muted);
}

.meter-bar {
  width: 100px;
  height: 6px;
  background: var(--bg-void);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.meter-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--status-safe), var(--accent-primary), var(--status-warning));
  border-radius: var(--radius-full);
  transition: width 0.3s ease;
}

.meter-value {
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
  color: var(--accent-primary);
  min-width: 50px;
}

/* Space Sections */
.space-section {
  padding: 20px;
  border-radius: var(--radius-lg);
  position: relative;
  transition: all 0.3s ease;
}

.user-space {
  background: linear-gradient(135deg, var(--bg-elevated) 0%, var(--bg-surface) 100%);
  border: 1px solid var(--border-subtle);
}

.kernel-space {
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.03) 0%, rgba(96, 165, 250, 0.03) 100%);
  border: 2px solid #8b5cf6;
}

.kernel-space.active {
  border-color: var(--status-learning);
  box-shadow: 0 0 30px rgba(139, 92, 246, 0.1);
}

.space-label {
  position: absolute;
  top: -10px;
  left: 20px;
  padding: 2px 12px;
  background: var(--bg-surface);
  font-size: 10px;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.15em;
}

.kernel-space .space-label {
  color: #8b5cf6;
  background: linear-gradient(135deg, var(--bg-surface), var(--bg-elevated));
}

/* Live Events Feed */
.events-live {
  display: flex;
  gap: 12px;
  min-height: 60px;
  margin-bottom: 20px;
  overflow-x: auto;
  padding: 4px;
}

.event-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: var(--bg-overlay);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  transition: all 0.3s ease;
  flex-shrink: 0;
  min-width: 180px;
  position: relative;
}

.event-box.blocked {
  border-color: var(--status-blocked);
  background: linear-gradient(135deg, var(--status-blocked-dim), transparent);
}

.event-box.process {
  border-color: var(--chart-exec);
}

.event-box.file {
  border-color: var(--chart-file);
}

.event-box.network {
  border-color: var(--chart-network);
}

.event-box.process.blocked,
.event-box.file.blocked,
.event-box.network.blocked {
  border-color: var(--status-blocked);
}

.event-status {
  position: absolute;
  top: -6px;
  right: -6px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--bg-surface);
}

.status-icon.blocked {
  color: var(--status-blocked);
}

.status-icon.allowed {
  color: var(--status-safe);
}

.event-icon {
  flex-shrink: 0;
}

.event-box.process .event-icon {
  color: var(--chart-exec);
}

.event-box.file .event-icon {
  color: var(--chart-file);
}

.event-box.network .event-icon {
  color: var(--chart-network);
}

.event-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.event-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.event-pid {
  font-family: var(--font-mono);
  font-size: 9px;
  color: var(--text-muted);
}

.event-syscall {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--text-muted);
  background: var(--bg-void);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

/* Event slide animation */
.event-slide-enter-active {
  animation: slide-in 0.4s ease-out;
}

.event-slide-leave-active {
  animation: slide-out 0.3s ease-in;
}

.event-slide-move {
  transition: transform 0.3s ease;
}

@keyframes slide-in {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.9);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes slide-out {
  from {
    opacity: 1;
    transform: translateX(0);
  }

  to {
    opacity: 0;
    transform: translateX(20px);
  }
}

.empty-events {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  color: var(--text-muted);
  font-size: 13px;
}

.empty-icon {
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {

  0%,
  100% {
    opacity: 0.5;
  }

  50% {
    opacity: 1;
  }
}

/* Syscall Flow Lanes */
.syscall-flows {
  display: flex;
  justify-content: center;
  gap: 40px;
}

.flow-lane {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.flow-label {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--text-muted);
  transition: color 0.3s ease;
}

.flow-lane.active .flow-label {
  color: var(--accent-primary);
}

.flow-track {
  width: 3px;
  height: 40px;
  background: var(--bg-void);
  border-radius: var(--radius-full);
  position: relative;
  overflow: hidden;
}

.flow-lane.active .flow-track {
  background: linear-gradient(to bottom, var(--accent-primary), #8b5cf6);
}

.flow-particle {
  position: absolute;
  width: 100%;
  height: 12px;
  background: linear-gradient(to bottom, transparent, #8b5cf6, transparent);
  border-radius: var(--radius-full);
  animation: flow-down 0.8s ease-out;
}

.flow-particle.blocked {
  background: linear-gradient(to bottom, transparent, var(--status-blocked), transparent);
}

@keyframes flow-down {
  0% {
    top: -12px;
    opacity: 1;
  }

  100% {
    top: 100%;
    opacity: 0;
  }
}

/* LSM Boundary */
.lsm-boundary {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 20px;
}

.boundary-line {
  flex: 1;
  height: 2px;
  background: var(--status-blocked);
  position: relative;
  overflow: hidden;
  border-radius: var(--radius-full);
}

.boundary-pulse {
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, #fff, transparent);
  opacity: 0;
}

.boundary-pulse.active {
  animation: boundary-sweep 1s ease-out;
}

@keyframes boundary-sweep {
  0% {
    left: -100%;
    opacity: 1;
  }

  100% {
    left: 100%;
    opacity: 0;
  }
}

.boundary-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 10px;
  font-weight: 600;
  color: var(--status-blocked);
  white-space: nowrap;
  padding: 4px 12px;
  background: var(--bg-surface);
  border-radius: var(--radius-full);
  border: 1px solid var(--status-blocked);
}

/* Hooks Row */
.hooks-row {
  display: flex;
  justify-content: center;
  gap: 20px;
  margin-bottom: 20px;
}

.hook-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 20px;
  background: var(--bg-elevated);
  border-radius: var(--radius-lg);
  border: 2px solid var(--border-default);
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  min-width: 150px;
}

.hook-glow {
  position: absolute;
  inset: -2px;
  border-radius: var(--radius-lg);
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
}

.hook-node.process .hook-glow {
  box-shadow: 0 0 30px var(--chart-exec);
}

.hook-node.file .hook-glow {
  box-shadow: 0 0 30px var(--chart-file);
}

.hook-node.network .hook-glow {
  box-shadow: 0 0 30px var(--chart-network);
}

.hook-node.pulsing .hook-glow {
  opacity: 0.5;
  animation: glow-pulse 0.3s ease-out;
}

@keyframes glow-pulse {
  0% {
    opacity: 0.8;
    transform: scale(1);
  }

  100% {
    opacity: 0;
    transform: scale(1.1);
  }
}

.hook-node:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.hook-node.pulsing {
  transform: scale(1.02);
}

.hook-indicator {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  animation: indicator-pulse 2s ease-in-out infinite;
}

.hook-node.process .hook-indicator {
  background: var(--chart-exec);
  box-shadow: 0 0 10px var(--chart-exec);
}

.hook-node.file .hook-indicator {
  background: var(--chart-file);
  box-shadow: 0 0 10px var(--chart-file);
}

.hook-node.network .hook-indicator {
  background: var(--chart-network);
  box-shadow: 0 0 10px var(--chart-network);
}

@keyframes indicator-pulse {

  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }

  50% {
    opacity: 0.6;
    transform: scale(1.2);
  }
}

.hook-node.process {
  border-color: var(--chart-exec);
}

.hook-node.file {
  border-color: var(--chart-file);
}

.hook-node.network {
  border-color: var(--chart-network);
}

.hook-icon {
  transition: transform 0.3s ease;
}

.hook-node:hover .hook-icon {
  transform: scale(1.1);
}

.hook-node.process .hook-icon {
  color: var(--chart-exec);
}

.hook-node.file .hook-icon {
  color: var(--chart-file);
}

.hook-node.network .hook-icon {
  color: var(--chart-network);
}

.hook-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  text-align: center;
}

.hook-signature {
  font-family: var(--font-mono);
  font-size: 9px;
  color: #8b5cf6;
  background: var(--bg-void);
  padding: 3px 8px;
  border-radius: var(--radius-sm);
}

.hook-capability {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  background: rgba(239, 68, 68, 0.1);
  color: rgba(239, 68, 68, 0.75);
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: var(--radius-sm);
  font-size: 9px;
  font-weight: 500;
}

.hook-stats {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding-top: 10px;
  border-top: 1px solid var(--border-subtle);
  width: 100%;
}

.stat-rate-container {
  display: flex;
  align-items: baseline;
  gap: 2px;
}

.stat-rate {
  font-family: var(--font-mono);
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

.stat-unit {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--text-muted);
}

.stat-bar {
  width: 100%;
  height: 4px;
  background: var(--bg-void);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.stat-bar-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width 0.3s ease;
}

.hook-node.process .stat-bar-fill {
  background: var(--chart-exec);
}

.hook-node.file .stat-bar-fill {
  background: var(--chart-file);
}

.hook-node.network .stat-bar-fill {
  background: var(--chart-network);
}

/* Data Flow */
.data-flow {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.flow-streams {
  display: flex;
  gap: 60px;
}

.flow-stream {
  width: 2px;
  height: 30px;
  background: var(--border-subtle);
  position: relative;
  border-radius: var(--radius-full);
}

.flow-stream.active {
  background: linear-gradient(to bottom, #8b5cf6, var(--accent-primary));
}

.stream-particle {
  position: absolute;
  width: 6px;
  height: 6px;
  left: -2px;
  background: #8b5cf6;
  border-radius: 50%;
  opacity: 0;
}

.flow-stream.active .stream-particle {
  animation: stream-flow 0.6s ease-out infinite;
}

@keyframes stream-flow {
  0% {
    top: 0;
    opacity: 1;
  }

  100% {
    top: 100%;
    opacity: 0;
  }
}

/* Ring Buffer */
.ring-buffer {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 24px;
  background: linear-gradient(135deg, var(--bg-overlay), var(--bg-elevated));
  border-radius: var(--radius-lg);
  border: 2px solid #8b5cf6;
  max-width: 420px;
  margin: 0 auto;
  transition: all 0.3s ease;
}

.ring-buffer.receiving {
  border-color: var(--accent-primary);
  box-shadow: 0 0 20px rgba(139, 92, 246, 0.2);
}

.buffer-icon-container {
  position: relative;
}

.buffer-icon {
  color: #8b5cf6;
  transition: transform 0.3s ease;
}

.ring-buffer.receiving .buffer-icon {
  animation: buffer-bounce 0.3s ease;
}

@keyframes buffer-bounce {

  0%,
  100% {
    transform: scale(1);
  }

  50% {
    transform: scale(1.2);
  }
}

.buffer-pulse {
  position: absolute;
  inset: -8px;
  border: 2px solid #8b5cf6;
  border-radius: 50%;
  opacity: 0;
}

.ring-buffer.receiving .buffer-pulse {
  animation: buffer-ring 0.6s ease-out;
}

@keyframes buffer-ring {
  0% {
    transform: scale(0.8);
    opacity: 1;
  }

  100% {
    transform: scale(1.5);
    opacity: 0;
  }
}

.buffer-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.buffer-title {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.1em;
}

.buffer-stats-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.buffer-count {
  font-family: var(--font-mono);
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

.buffer-label {
  font-size: 11px;
  color: var(--text-muted);
}

.buffer-meter {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.meter-ring {
  position: relative;
  width: 50px;
  height: 50px;
}

.circular-chart {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.circle-bg {
  fill: none;
  stroke: var(--bg-void);
  stroke-width: 3;
}

.circle-fill {
  fill: none;
  stroke: #8b5cf6;
  stroke-width: 3;
  stroke-linecap: round;
  transition: stroke-dasharray 0.3s ease;
}

.meter-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 700;
  color: var(--text-primary);
}

.buffer-meter .meter-label {
  font-size: 10px;
  color: var(--text-muted);
}

/* Legend */
.diagram-legend {
  display: flex;
  justify-content: center;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.legend-items {
  display: flex;
  justify-content: center;
  gap: 16px;
  flex-wrap: wrap;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-secondary);
  padding: 4px 10px;
  border-radius: var(--radius-full);
  background: var(--bg-overlay);
}

.legend-item.process {
  color: var(--chart-exec);
}

.legend-item.file {
  color: var(--chart-file);
}

.legend-item.network {
  color: var(--chart-network);
}

.legend-item.blocked {
  color: var(--status-blocked);
}

.legend-item.allowed {
  color: var(--status-safe);
}
</style>

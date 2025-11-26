<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Terminal, Globe, FileText, Database, Zap } from 'lucide-vue-next'
import { probes, type ProbeInfo } from '../../data/probes'
import { subscribeToAllEvents, type ProbeStats, type StreamEvent } from '../../lib/api'

const props = defineProps<{
  probeStats: ProbeStats[]
}>()

defineEmits<{
  selectProbe: [probe: ProbeInfo]
}>()

interface RecentProcess {
  id: string
  name: string
  pid: number
  type: 'exec' | 'file' | 'connect'
  timestamp: number
  syscall: string
  detail: string
}

const recentProcesses = ref<RecentProcess[]>([])
const maxRecentProcesses = 5

const activeFlows = ref<{ id: string; type: string; startTime: number }[]>([])
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
  
  let process: RecentProcess
  let probeType: string
  
  if (event.type === 'exec') {
    process = {
      id,
      name: event.comm,
      pid: event.pid,
      type: 'exec',
      timestamp: event.timestamp,
      syscall: 'execve()',
      detail: `from ${event.parentComm}`
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
      detail: event.filename.length > 30 ? '...' + event.filename.slice(-27) : event.filename
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
      detail: `port ${event.port}`
    }
    probeType = 'connect'
  }
  
  recentProcesses.value = [process, ...recentProcesses.value].slice(0, maxRecentProcesses)
  
  activeFlows.value.push({ id, type: event.type, startTime: Date.now() })
  
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
      <div class="activity-meter">
        <span class="meter-label">Activity</span>
        <div class="meter-bar">
          <div 
            class="meter-fill" 
            :style="{ width: Math.min(totalEventsPerSec * 5, 100) + '%' }"
          ></div>
        </div>
        <span class="meter-value">{{ totalEventsPerSec }}/s</span>
      </div>
    </div>

    <!-- User Space -->
    <div class="space-section user-space">
      <div class="space-label">USER SPACE</div>
      
      <!-- Recent Processes - Live Feed -->
      <div class="processes-live">
        <TransitionGroup name="process-slide">
          <div 
            v-for="proc in recentProcesses"
            :key="proc.id"
            class="process-box"
            :class="[getEventColorClass(proc.type), { entering: true }]"
          >
            <component :is="getEventIcon(proc.type)" :size="16" class="proc-icon" />
            <div class="proc-info">
              <span class="proc-name">{{ proc.name }}</span>
              <code class="proc-pid">PID: {{ proc.pid }}</code>
            </div>
            <span class="proc-syscall">{{ proc.syscall }}</span>
          </div>
        </TransitionGroup>
        
        <!-- Empty state -->
        <div v-if="recentProcesses.length === 0" class="empty-processes">
          <Zap :size="24" class="empty-icon" />
          <span>Waiting for events...</span>
        </div>
      </div>
      
      <!-- Animated Syscall Flows -->
      <div class="syscall-flows">
        <div class="flow-lane" :class="{ active: isFlowActive('exec') }">
          <div class="flow-label">execve()</div>
          <div class="flow-track">
            <div class="flow-particle" v-for="flow in activeFlows.filter(f => f.type === 'exec')" :key="flow.id"></div>
          </div>
        </div>
        <div class="flow-lane" :class="{ active: isFlowActive('file') }">
          <div class="flow-label">open()</div>
          <div class="flow-track">
            <div class="flow-particle" v-for="flow in activeFlows.filter(f => f.type === 'file')" :key="flow.id"></div>
          </div>
        </div>
        <div class="flow-lane" :class="{ active: isFlowActive('connect') }">
          <div class="flow-label">connect()</div>
          <div class="flow-track">
            <div class="flow-particle" v-for="flow in activeFlows.filter(f => f.type === 'connect')" :key="flow.id"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Kernel Boundary - Animated -->
    <div class="kernel-boundary">
      <div class="boundary-line">
        <div class="boundary-pulse" :class="{ active: activeFlows.length > 0 }"></div>
      </div>
      <span class="boundary-label">
        <Zap :size="12" />
        System Call Interface
      </span>
      <div class="boundary-line">
        <div class="boundary-pulse" :class="{ active: activeFlows.length > 0 }"></div>
      </div>
    </div>

    <!-- Kernel Space -->
    <div class="space-section kernel-space" :class="{ active: activeFlows.length > 0 }">
      <div class="space-label">KERNEL SPACE</div>
      
      <!-- eBPF Probes -->
      <div class="probes-row">
        <div 
          v-for="probe in probes"
          :key="probe.id"
          class="probe-node"
          :class="[
            probe.category, 
            { 
              active: getStatsForProbe(probe.id)?.active,
              pulsing: isProbePulsing(probe.id)
            }
          ]"
          @click="$emit('selectProbe', probe)"
        >
          <div class="probe-glow"></div>
          <div class="probe-indicator"></div>
          <component :is="getProbeIcon(probe.category)" :size="20" class="probe-icon" />
          <span class="probe-name">{{ probe.name.replace(' Monitor', '') }}</span>
          <code class="probe-tracepoint">{{ probe.tracepoint.split('/').pop() }}</code>
          <div class="probe-stats" v-if="getStatsForProbe(probe.id)">
            <div class="stat-rate-container">
              <span class="stat-rate">{{ getStatsForProbe(probe.id)?.eventsRate || 0 }}</span>
              <span class="stat-unit">/sec</span>
            </div>
            <div class="stat-bar">
              <div 
                class="stat-bar-fill" 
                :style="{ width: Math.min((getStatsForProbe(probe.id)?.eventsRate || 0) * 10, 100) + '%' }"
              ></div>
            </div>
            <span class="stat-total">{{ formatNumber(getStatsForProbe(probe.id)?.totalCount || 0) }} captured</span>
          </div>
        </div>
      </div>

      <!-- Data Flow to Ring Buffer -->
      <div class="data-flow">
        <div class="flow-streams">
          <div 
            v-for="i in 3" 
            :key="i"
            class="flow-stream"
            :class="{ active: activeFlows.length > 0 }"
          >
            <div class="stream-particle"></div>
          </div>
        </div>
      </div>

      <!-- Ring Buffer - Animated -->
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
              <path
                class="circle-bg"
                d="M18 2.0845
                  a 15.9155 15.9155 0 0 1 0 31.831
                  a 15.9155 15.9155 0 0 1 0 -31.831"
              />
              <path
                class="circle-fill"
                :stroke-dasharray="`${Math.min(totalEventsPerSec * 2, 100)}, 100`"
                d="M18 2.0845
                  a 15.9155 15.9155 0 0 1 0 31.831
                  a 15.9155 15.9155 0 0 1 0 -31.831"
              />
            </svg>
            <span class="meter-text">{{ totalEventsPerSec }}</span>
          </div>
          <span class="meter-label">/sec</span>
        </div>
      </div>
    </div>

    <!-- Interactive Legend -->
    <div class="diagram-legend">
      <div class="legend-stats">
        <div class="legend-stat">
          <span class="stat-number">{{ formatNumber(totalEvents) }}</span>
          <span class="stat-label">Total Events</span>
        </div>
        <div class="legend-stat">
          <span class="stat-number">{{ totalEventsPerSec }}</span>
          <span class="stat-label">Events/sec</span>
        </div>
        <div class="legend-stat">
          <span class="stat-number">{{ recentProcesses.length }}</span>
          <span class="stat-label">In Queue</span>
        </div>
      </div>
      <div class="legend-items">
        <div class="legend-item process">
          <Terminal :size="14" />
          <span>Process</span>
        </div>
        <div class="legend-item file">
          <FileText :size="14" />
          <span>File</span>
        </div>
        <div class="legend-item network">
          <Globe :size="14" />
          <span>Network</span>
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
  0%, 100% { opacity: 1; box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7); }
  50% { opacity: 0.8; box-shadow: 0 0 0 8px rgba(239, 68, 68, 0); }
}

.live-text {
  font-size: 11px;
  font-weight: 700;
  color: #ef4444;
  letter-spacing: 0.1em;
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
  background: linear-gradient(135deg, rgba(96, 165, 250, 0.03) 0%, rgba(139, 92, 246, 0.03) 100%);
  border: 2px solid var(--accent-primary);
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
  color: var(--accent-primary);
  background: linear-gradient(135deg, var(--bg-surface), var(--bg-elevated));
}

/* Live Processes Feed */
.processes-live {
  display: flex;
  gap: 12px;
  min-height: 60px;
  margin-bottom: 20px;
  overflow-x: auto;
  padding: 4px;
}

.process-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: var(--bg-overlay);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  transition: all 0.3s ease;
  flex-shrink: 0;
  min-width: 160px;
}

.process-box.process {
  border-color: var(--status-info);
  background: linear-gradient(135deg, rgba(96, 165, 250, 0.1), transparent);
}

.process-box.file {
  border-color: var(--status-safe);
  background: linear-gradient(135deg, rgba(52, 211, 153, 0.1), transparent);
}

.process-box.network {
  border-color: var(--status-warning);
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.1), transparent);
}

.proc-icon {
  flex-shrink: 0;
}

.process-box.process .proc-icon { color: var(--status-info); }
.process-box.file .proc-icon { color: var(--status-safe); }
.process-box.network .proc-icon { color: var(--status-warning); }

.proc-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.proc-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.proc-pid {
  font-family: var(--font-mono);
  font-size: 9px;
  color: var(--text-muted);
}

.proc-syscall {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--text-muted);
  background: var(--bg-void);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

/* Process slide animation */
.process-slide-enter-active {
  animation: slide-in 0.4s ease-out;
}

.process-slide-leave-active {
  animation: slide-out 0.3s ease-in;
}

.process-slide-move {
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

.empty-processes {
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
  0%, 100% { opacity: 0.5; }
  50% { opacity: 1; }
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
  background: linear-gradient(to bottom, var(--accent-primary), var(--status-learning));
}

.flow-particle {
  position: absolute;
  width: 100%;
  height: 12px;
  background: linear-gradient(to bottom, transparent, var(--status-learning), transparent);
  border-radius: var(--radius-full);
  animation: flow-down 0.8s ease-out;
}

@keyframes flow-down {
  0% { top: -12px; opacity: 1; }
  100% { top: 100%; opacity: 0; }
}

/* Kernel Boundary */
.kernel-boundary {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 20px;
}

.boundary-line {
  flex: 1;
  height: 2px;
  background: var(--border-subtle);
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
  background: linear-gradient(90deg, transparent, var(--accent-primary), transparent);
  opacity: 0;
}

.boundary-pulse.active {
  animation: boundary-sweep 1s ease-out;
}

@keyframes boundary-sweep {
  0% { left: -100%; opacity: 1; }
  100% { left: 100%; opacity: 0; }
}

.boundary-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 10px;
  font-weight: 600;
  color: var(--accent-primary);
  white-space: nowrap;
  padding: 4px 12px;
  background: var(--bg-surface);
  border-radius: var(--radius-full);
  border: 1px solid var(--accent-primary);
}

/* Probes Row */
.probes-row {
  display: flex;
  justify-content: center;
  gap: 20px;
  margin-bottom: 20px;
}

.probe-node {
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
  min-width: 140px;
}

.probe-glow {
  position: absolute;
  inset: -2px;
  border-radius: var(--radius-lg);
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
}

.probe-node.process .probe-glow {
  box-shadow: 0 0 30px var(--status-info);
}

.probe-node.file .probe-glow {
  box-shadow: 0 0 30px var(--status-safe);
}

.probe-node.network .probe-glow {
  box-shadow: 0 0 30px var(--status-warning);
}

.probe-node.pulsing .probe-glow {
  opacity: 0.5;
  animation: glow-pulse 0.3s ease-out;
}

@keyframes glow-pulse {
  0% { opacity: 0.8; transform: scale(1); }
  100% { opacity: 0; transform: scale(1.1); }
}

.probe-node:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.probe-node.pulsing {
  transform: scale(1.02);
}

.probe-indicator {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  animation: indicator-pulse 2s ease-in-out infinite;
}

.probe-node.process .probe-indicator {
  background: var(--status-info);
  box-shadow: 0 0 10px var(--status-info);
}

.probe-node.file .probe-indicator {
  background: var(--status-safe);
  box-shadow: 0 0 10px var(--status-safe);
}

.probe-node.network .probe-indicator {
  background: var(--status-warning);
  box-shadow: 0 0 10px var(--status-warning);
}

@keyframes indicator-pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.6; transform: scale(1.2); }
}

.probe-node.process { border-color: var(--status-info); }
.probe-node.file { border-color: var(--status-safe); }
.probe-node.network { border-color: var(--status-warning); }

.probe-icon {
  transition: transform 0.3s ease;
}

.probe-node:hover .probe-icon {
  transform: scale(1.1);
}

.probe-node.process .probe-icon { color: var(--status-info); }
.probe-node.file .probe-icon { color: var(--status-safe); }
.probe-node.network .probe-icon { color: var(--status-warning); }

.probe-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  text-align: center;
}

.probe-tracepoint {
  font-family: var(--font-mono);
  font-size: 9px;
  color: var(--text-muted);
  background: var(--bg-void);
  padding: 3px 8px;
  border-radius: var(--radius-sm);
}

.probe-stats {
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

.probe-node.process .stat-bar-fill { background: var(--status-info); }
.probe-node.file .stat-bar-fill { background: var(--status-safe); }
.probe-node.network .stat-bar-fill { background: var(--status-warning); }

.stat-total {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--text-muted);
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
  background: linear-gradient(to bottom, var(--accent-primary), var(--status-learning));
}

.stream-particle {
  position: absolute;
  width: 6px;
  height: 6px;
  left: -2px;
  background: var(--accent-primary);
  border-radius: 50%;
  opacity: 0;
}

.flow-stream.active .stream-particle {
  animation: stream-flow 0.6s ease-out infinite;
}

@keyframes stream-flow {
  0% { top: 0; opacity: 1; }
  100% { top: 100%; opacity: 0; }
}

/* Ring Buffer */
.ring-buffer {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 24px;
  background: linear-gradient(135deg, var(--bg-overlay), var(--bg-elevated));
  border-radius: var(--radius-lg);
  border: 2px solid var(--accent-primary);
  max-width: 420px;
  margin: 0 auto;
  transition: all 0.3s ease;
}

.ring-buffer.receiving {
  border-color: var(--status-learning);
  box-shadow: 0 0 20px rgba(139, 92, 246, 0.2);
}

.buffer-icon-container {
  position: relative;
}

.buffer-icon {
  color: var(--accent-primary);
  transition: transform 0.3s ease;
}

.ring-buffer.receiving .buffer-icon {
  animation: buffer-bounce 0.3s ease;
}

@keyframes buffer-bounce {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.2); }
}

.buffer-pulse {
  position: absolute;
  inset: -8px;
  border: 2px solid var(--accent-primary);
  border-radius: 50%;
  opacity: 0;
}

.ring-buffer.receiving .buffer-pulse {
  animation: buffer-ring 0.6s ease-out;
}

@keyframes buffer-ring {
  0% { transform: scale(0.8); opacity: 1; }
  100% { transform: scale(1.5); opacity: 0; }
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
  stroke: var(--accent-primary);
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
  flex-direction: column;
  gap: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.legend-stats {
  display: flex;
  justify-content: center;
  gap: 32px;
}

.legend-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.stat-number {
  font-family: var(--font-mono);
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

.stat-label {
  font-size: 10px;
  color: var(--text-muted);
}

.legend-items {
  display: flex;
  justify-content: center;
  gap: 24px;
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

.legend-item.process { color: var(--status-info); }
.legend-item.file { color: var(--status-safe); }
.legend-item.network { color: var(--status-warning); }
</style>

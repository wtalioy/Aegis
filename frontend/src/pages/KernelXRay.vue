<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Cpu, BookOpen, Info } from 'lucide-vue-next'
import ArchitectureDiagram from '../components/kernel/ArchitectureDiagram.vue'
import ProbeCard from '../components/kernel/ProbeCard.vue'
import { probes, kernelStructs, type ProbeInfo } from '../data/probes'
import { getProbeStats, type ProbeStats } from '../lib/api'

const selectedProbe = ref<ProbeInfo | null>(null)
const probeStats = ref<ProbeStats[]>([])
let pollInterval: number | null = null

const selectProbe = (probe: ProbeInfo) => {
  selectedProbe.value = probe
}

const closeProbeCard = () => {
  selectedProbe.value = null
}

const fetchProbeStats = async () => {
  try {
    probeStats.value = await getProbeStats()
  } catch (e) {
    console.error('Failed to fetch probe stats:', e)
  }
}

const getProbeStatsById = (id: string): ProbeStats | undefined => {
  return probeStats.value.find(p => p.id === id)
}

onMounted(() => {
  fetchProbeStats()
  pollInterval = window.setInterval(fetchProbeStats, 1000)
})

onUnmounted(() => {
  if (pollInterval) {
    clearInterval(pollInterval)
  }
})
</script>

<template>
  <div class="kernel-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <Cpu :size="24" class="title-icon" />
          Kernel X-Ray
        </h1>
        <span class="page-subtitle">Interactive eBPF probe visualization and kernel internals</span>
      </div>
    </div>

    <!-- Main Content -->
    <div class="kernel-content">
      <!-- Architecture Diagram -->
      <section class="content-section">
        <div class="section-header">
          <h2 class="section-title">
            <BookOpen :size="18" />
            System Architecture
          </h2>
          <span class="section-description">
            How EulerGuard monitors your system with eBPF
          </span>
        </div>
        <ArchitectureDiagram :probe-stats="probeStats" @select-probe="selectProbe" />
      </section>

      <!-- Probe Overview -->
      <section class="content-section probes-overview">
        <div class="section-header">
          <h2 class="section-title">
            <Info :size="18" />
            Active Probes
          </h2>
        </div>
        <div class="probes-grid">
          <div 
            v-for="probe in probes" 
            :key="probe.id"
            class="probe-summary"
            :class="probe.category"
            @click="selectProbe(probe)"
          >
            <div class="summary-header">
              <span class="summary-name">{{ probe.name }}</span>
              <span class="summary-badge">{{ probe.category.toUpperCase() }}</span>
            </div>
            <code class="summary-tracepoint">{{ probe.tracepoint }}</code>
            <p class="summary-desc">{{ probe.description.slice(0, 100) }}...</p>
            <span class="summary-action">Click to learn more →</span>
          </div>
        </div>
      </section>

      <!-- Kernel Structures Reference -->
      <section class="content-section">
        <div class="section-header">
          <h2 class="section-title">
            <Cpu :size="18" />
            Kernel Structures Reference
          </h2>
          <span class="section-description">
            Key data structures accessed by eBPF probes
          </span>
        </div>
        <div class="structs-grid">
          <div v-for="struct in kernelStructs" :key="struct.name" class="struct-card">
            <div class="struct-header">
              <code class="struct-name">{{ struct.name }}</code>
            </div>
            <p class="struct-desc">{{ struct.description }}</p>
            <div class="struct-fields">
              <div v-for="field in struct.fields" :key="field.name" class="field-row">
                <code class="field-name">{{ field.name }}</code>
                <code class="field-type">{{ field.type }}</code>
                <span class="field-desc">{{ field.description }}</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Educational Note -->
      <section class="content-section educational-note">
        <div class="note-content">
          <h3 class="note-title">🎓 What is eBPF?</h3>
          <p class="note-text">
            eBPF (extended Berkeley Packet Filter) is a revolutionary technology that allows 
            running sandboxed programs in the Linux kernel without changing kernel source code 
            or loading kernel modules. EulerGuard uses eBPF to safely and efficiently monitor 
            system calls in real-time, enabling deep visibility into process executions, file 
            accesses, and network connections without impacting system performance.
          </p>
          <div class="note-features">
            <div class="feature">
              <span class="feature-icon">⚡</span>
              <span>Near-zero overhead</span>
            </div>
            <div class="feature">
              <span class="feature-icon">🔒</span>
              <span>Safe & sandboxed</span>
            </div>
            <div class="feature">
              <span class="feature-icon">👁️</span>
              <span>Deep visibility</span>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- Probe Detail Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <ProbeCard 
          v-if="selectedProbe" 
          :probe="selectedProbe"
          @close="closeProbeCard"
        />
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.kernel-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.title-icon {
  color: var(--accent-primary);
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-muted);
}

/* Content Sections */
.kernel-content {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.content-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.section-title svg {
  color: var(--text-muted);
}

.section-description {
  font-size: 13px;
  color: var(--text-muted);
  margin-left: 28px;
}

/* Probes Grid */
.probes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.probe-summary {
  padding: 20px;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.probe-summary:hover {
  border-color: var(--border-default);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.probe-summary.process:hover { border-color: var(--status-info); }
.probe-summary.file:hover { border-color: var(--status-safe); }
.probe-summary.network:hover { border-color: var(--status-warning); }

.summary-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.summary-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.summary-badge {
  padding: 3px 8px;
  background: var(--bg-overlay);
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 600;
  color: var(--text-muted);
}

.probe-summary.process .summary-badge { color: var(--status-info); background: var(--status-info-dim); }
.probe-summary.file .summary-badge { color: var(--status-safe); background: var(--status-safe-dim); }
.probe-summary.network .summary-badge { color: var(--status-warning); background: var(--status-warning-dim); }

.summary-tracepoint {
  display: block;
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--accent-primary);
  background: var(--bg-void);
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  margin-bottom: 12px;
}

.summary-desc {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.6;
  margin: 0 0 12px 0;
}

.summary-action {
  font-size: 12px;
  color: var(--accent-primary);
  font-weight: 500;
}

/* Structs Grid */
.structs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 16px;
}

.struct-card {
  padding: 20px;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
}

.struct-header {
  margin-bottom: 8px;
}

.struct-name {
  font-family: var(--font-mono);
  font-size: 16px;
  font-weight: 600;
  color: var(--accent-primary);
}

.struct-desc {
  font-size: 13px;
  color: var(--text-muted);
  margin: 0 0 16px 0;
  line-height: 1.5;
}

.struct-fields {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-row {
  display: grid;
  grid-template-columns: 120px 140px 1fr;
  gap: 12px;
  padding: 8px 12px;
  background: var(--bg-elevated);
  border-radius: var(--radius-sm);
  align-items: center;
}

.field-name {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-primary);
}

.field-type {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--status-learning);
}

.field-desc {
  font-size: 11px;
  color: var(--text-muted);
}

/* Educational Note */
.educational-note {
  background: linear-gradient(135deg, var(--bg-surface) 0%, var(--bg-elevated) 100%);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  padding: 24px;
}

.note-content {
  max-width: 800px;
}

.note-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 12px 0;
}

.note-text {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.7;
  margin: 0 0 20px 0;
}

.note-features {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
}

.feature {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-primary);
}

.feature-icon {
  font-size: 16px;
}

/* Modal Transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

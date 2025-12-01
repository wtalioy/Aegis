<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { Cpu, BookOpen, Info, ShieldOff, ShieldCheck, Zap, Terminal, FileText, Globe, Code } from 'lucide-vue-next'
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

const getProbeIcon = (category: string) => {
  switch (category) {
    case 'process': return Terminal
    case 'file': return FileText
    case 'network': return Globe
    default: return Terminal
  }
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
        <span class="page-subtitle">Interactive eBPF & LSM visualization</span>
      </div>
      <div class="header-badge">
        <ShieldOff :size="14" />
        <span>LSM Active Defense</span>
      </div>
    </div>

    <!-- LSM Overview Banner -->
    <div class="lsm-banner">
      <div class="banner-content">
        <div class="banner-icon">
          <Zap :size="24" />
        </div>
        <div class="banner-text">
          <h3>BPF LSM Hooks Enabled</h3>
          <p>EulerGuard uses Linux Security Modules (LSM) hooks for <strong>active defense</strong>.
            Unlike passive tracepoints, LSM hooks can block malicious operations
            in real-time by returning <code>-EPERM</code>.</p>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="kernel-content">
      <!-- Architecture Diagram -->
      <section class="content-section">
        <div class="section-header">
          <h2 class="section-title">
            <BookOpen :size="18" />
            Live System Architecture
          </h2>
          <span class="section-description">
            Real-time visualization of eBPF LSM hooks intercepting system calls
          </span>
        </div>
        <ArchitectureDiagram :probe-stats="probeStats" @select-probe="selectProbe" />
      </section>

      <!-- LSM Hooks Overview -->
      <section class="content-section hooks-overview">
        <div class="section-header">
          <h2 class="section-title">
            <ShieldCheck :size="18" />
            LSM Security Hooks
          </h2>
          <span class="section-description">
            Click any hook to view implementation details
          </span>
        </div>
        <div class="hooks-grid">
          <div v-for="probe in probes" :key="probe.id" class="hook-card" :class="probe.category"
            @click="selectProbe(probe)">
            <div class="hook-header">
              <div class="hook-icon" :class="probe.category">
                <component :is="getProbeIcon(probe.category)" :size="20" />
              </div>
              <div class="hook-badges">
                <span class="hook-type">LSM</span>
                <span class="hook-capability" :class="probe.capability">
                  <component :is="probe.capability === 'block' ? ShieldOff : ShieldCheck" :size="10" />
                  {{ probe.capability.toUpperCase() }}
                </span>
              </div>
            </div>
            <h3 class="hook-name">{{ probe.name }}</h3>
            <code class="hook-signature">{{ probe.hook }}</code>
            <p class="hook-desc">{{ probe.description.slice(0, 120) }}...</p>
            <div class="hook-footer">
              <span class="hook-action">View Implementation →</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Blocking vs Alerting Comparison -->
      <section class="content-section comparison-section">
        <div class="section-header">
          <h2 class="section-title">
            <Info :size="18" />
            Active Defense vs Passive Monitoring
          </h2>
        </div>
        <div class="comparison-grid">
          <div class="comparison-card block">
            <div class="comparison-header">
              <ShieldOff :size="24" />
              <h3>Block (Active Defense)</h3>
            </div>
            <p class="comparison-desc">
              LSM hooks return <code>-EPERM</code> to deny the operation at kernel level.
              The malicious action is <strong>prevented</strong> before it can cause harm.
            </p>
            <div class="comparison-example">
              <span class="example-label">Example:</span>
              <code>Block access to /etc/shadow</code>
            </div>
            <div class="comparison-flow">
              <span class="flow-step">syscall()</span>
              <span class="flow-arrow">→</span>
              <span class="flow-step lsm">LSM Hook</span>
              <span class="flow-arrow">→</span>
              <span class="flow-step denied">-EPERM</span>
            </div>
          </div>
          <div class="comparison-card alert">
            <div class="comparison-header">
              <ShieldCheck :size="24" />
              <h3>Alert (Passive Monitoring)</h3>
            </div>
            <p class="comparison-desc">
              LSM hooks return <code>0</code> to allow the operation but emit an event
              to the ring buffer for userspace analysis and alerting.
            </p>
            <div class="comparison-example">
              <span class="example-label">Example:</span>
              <code>Monitor access to /etc/passwd</code>
            </div>
            <div class="comparison-flow">
              <span class="flow-step">syscall()</span>
              <span class="flow-arrow">→</span>
              <span class="flow-step lsm">LSM Hook</span>
              <span class="flow-arrow">→</span>
              <span class="flow-step allowed">0 (allow)</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Kernel Structures Reference -->
      <section class="content-section">
        <div class="section-header">
          <h2 class="section-title">
            <Code :size="18" />
            Kernel Structures Reference
          </h2>
          <span class="section-description">
            Key data structures accessed by LSM hooks
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
          <h3 class="note-title">🎓 What is BPF LSM?</h3>
          <p class="note-text">
            BPF LSM (Linux Security Modules with eBPF) is a powerful security framework that allows
            running custom security policies in the kernel. Unlike traditional eBPF tracepoints which
            only observe events, LSM hooks are <strong>decision points</strong> that can actively
            allow or deny operations. EulerGuard leverages this for real-time threat prevention.
          </p>
          <div class="note-features">
            <div class="feature">
              <span class="feature-icon">🛡️</span>
              <span>Active blocking</span>
            </div>
            <div class="feature">
              <span class="feature-icon">⚡</span>
              <span>Kernel-speed decisions</span>
            </div>
            <div class="feature">
              <span class="feature-icon">🔒</span>
              <span>Tamper-resistant</span>
            </div>
            <div class="feature">
              <span class="feature-icon">📊</span>
              <span>Rich context capture</span>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- Probe Detail Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <ProbeCard v-if="selectedProbe" :probe="selectedProbe" @close="closeProbeCard" />
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
  align-items: center;
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

.header-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: rgba(139, 92, 246, 0.1);
  border: 1px solid rgba(139, 92, 246, 0.3);
  border-radius: var(--radius-full);
  color: #8b5cf6;
  font-size: 11px;
  font-weight: 500;
}

/* LSM Banner */
.lsm-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.08), rgba(96, 165, 250, 0.05));
  border: 1px solid rgba(139, 92, 246, 0.3);
  border-radius: var(--radius-lg);
}

.banner-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.banner-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: rgba(139, 92, 246, 0.15);
  border-radius: var(--radius-md);
  color: #8b5cf6;
  flex-shrink: 0;
}

.banner-text h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 6px 0;
}

.banner-text p {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0;
  max-width: 600px;
  line-height: 1.6;
}

.banner-text code {
  background: var(--bg-void);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-size: 12px;
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

/* Hooks Grid */
.hooks-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.hook-card {
  padding: 20px;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  cursor: pointer;
  transition: all var(--transition-fast);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.hook-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.hook-card.process:hover {
  border-color: var(--chart-exec);
}

.hook-card.file:hover {
  border-color: var(--chart-file);
}

.hook-card.network:hover {
  border-color: var(--chart-network);
}

.hook-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.hook-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
}

.hook-icon.process {
  background: rgba(96, 165, 250, 0.15);
  color: var(--chart-exec);
}

.hook-icon.file {
  background: rgba(16, 185, 129, 0.15);
  color: var(--chart-file);
}

.hook-icon.network {
  background: rgba(245, 158, 11, 0.15);
  color: var(--chart-network);
}

.hook-badges {
  display: flex;
  gap: 6px;
}

.hook-type {
  padding: 3px 8px;
  background: rgba(139, 92, 246, 0.15);
  color: #8b5cf6;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 600;
}

.hook-capability {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 500;
}

.hook-capability.block {
  background: rgba(239, 68, 68, 0.1);
  color: rgba(239, 68, 68, 0.75);
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.hook-capability.monitor {
  background: rgba(251, 191, 36, 0.1);
  color: rgba(251, 191, 36, 0.75);
  border: 1px solid rgba(251, 191, 36, 0.2);
}

.hook-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.hook-signature {
  display: block;
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--accent-primary);
  background: var(--bg-void);
  padding: 6px 10px;
  border-radius: var(--radius-sm);
}

.hook-desc {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.6;
  margin: 0;
  flex: 1;
}

.hook-footer {
  padding-top: 12px;
  border-top: 1px solid var(--border-subtle);
}

.hook-action {
  font-size: 12px;
  color: var(--accent-primary);
  font-weight: 500;
}

/* Comparison Section */
.comparison-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.comparison-card {
  padding: 24px;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 2px solid var(--border-subtle);
}

.comparison-card.block {
  border-color: var(--status-blocked);
  background: linear-gradient(135deg, var(--status-blocked-dim), var(--bg-surface) 60%);
}

.comparison-card.alert {
  border-color: var(--status-safe);
  background: linear-gradient(135deg, var(--status-safe-dim), var(--bg-surface) 60%);
}

.comparison-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.comparison-card.block .comparison-header {
  color: var(--status-blocked);
}

.comparison-card.alert .comparison-header {
  color: var(--status-safe);
}

.comparison-header h3 {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary);
}

.comparison-desc {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 0 0 16px 0;
}

.comparison-desc code {
  background: var(--bg-void);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.comparison-example {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  margin-bottom: 16px;
}

.example-label {
  font-size: 11px;
  color: var(--text-muted);
}

.comparison-example code {
  font-size: 12px;
  color: var(--text-primary);
}

.comparison-flow {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.flow-step {
  padding: 6px 12px;
  background: var(--bg-elevated);
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-secondary);
}

.flow-step.lsm {
  background: rgba(139, 92, 246, 0.15);
  color: #8b5cf6;
  font-weight: 600;
}

.flow-step.denied {
  background: var(--status-blocked);
  color: #fff;
}

.flow-step.allowed {
  background: var(--status-safe);
  color: #fff;
}

.flow-arrow {
  color: var(--text-muted);
  font-size: 14px;
}

/* Structs Grid */
.structs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(380px, 1fr));
  gap: 20px;
}

.struct-card {
  padding: 24px;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
}

.struct-header {
  margin-bottom: 10px;
}

.struct-name {
  font-family: var(--font-mono);
  font-size: 16px;
  font-weight: 600;
  color: var(--accent-primary);
}

.struct-desc {
  font-size: 14px;
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
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  padding: 12px 14px;
  background: var(--bg-elevated);
  border-radius: var(--radius-sm);
  align-items: baseline;
}

.field-name {
  font-family: var(--font-mono);
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  flex-shrink: 0;
}

.field-type {
  font-family: var(--font-mono);
  font-size: 12px;
  color: #8b5cf6;
  flex-shrink: 0;
}

.field-desc {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.4;
  flex: 1;
  min-width: 120px;
}

/* Educational Note */
.educational-note {
  background: linear-gradient(135deg, var(--bg-surface) 0%, var(--bg-elevated) 100%);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  padding: 24px;
}

.note-content {
  max-width: 900px;
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
  gap: 20px;
  flex-wrap: wrap;
}

.feature {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-primary);
  padding: 8px 14px;
  background: var(--bg-overlay);
  border-radius: var(--radius-md);
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

/* Responsive */
@media (max-width: 1100px) {
  .hooks-grid {
    grid-template-columns: 1fr;
  }

  .comparison-grid {
    grid-template-columns: 1fr;
  }

  .lsm-banner {
    flex-direction: column;
    gap: 20px;
  }

  .banner-stats {
    width: 100%;
    justify-content: center;
  }
}

@media (max-width: 600px) {
  .structs-grid {
    grid-template-columns: 1fr;
  }

  .field-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .field-desc {
    min-width: 0;
  }
}
</style>

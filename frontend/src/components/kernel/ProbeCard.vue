<script setup lang="ts">
import { X, Terminal, Globe, FileText, Code, Database, ShieldOff, Zap } from 'lucide-vue-next'
import type { ProbeInfo } from '../../data/probes'

defineProps<{
  probe: ProbeInfo
}>()

defineEmits<{
  close: []
}>()

const getCategoryIcon = (category: string) => {
  switch (category) {
    case 'process': return Terminal
    case 'network': return Globe
    case 'file': return FileText
    default: return Terminal
  }
}
</script>

<template>
  <div class="probe-card-overlay" @click.self="$emit('close')">
    <div class="probe-card">
      <div class="card-header">
        <div class="header-info">
          <div class="probe-icon" :class="probe.category">
            <component :is="getCategoryIcon(probe.category)" :size="20" />
          </div>
          <div class="probe-title">
            <h2 class="title">{{ probe.name }}</h2>
            <div class="probe-badges">
              <code class="hook-name">{{ probe.hook }}</code>
              <span class="hook-type">{{ probe.hookType.toUpperCase() }}</span>
              <span class="capability" :class="probe.capability">
                <ShieldOff v-if="probe.capability === 'block'" :size="10" />
                <Zap v-else :size="10" />
                {{ probe.capability.toUpperCase() }}
              </span>
            </div>
          </div>
        </div>
        <button class="close-btn" @click="$emit('close')">
          <X :size="20" />
        </button>
      </div>

      <div class="card-content">
        <!-- Description -->
        <section class="card-section">
          <h3 class="section-title">
            <span class="section-icon">📋</span>
            Description
          </h3>
          <p class="description">{{ probe.description }}</p>
        </section>

        <!-- Capability Info -->
        <section class="card-section capability-info" :class="probe.capability">
          <div class="capability-header">
            <ShieldOff v-if="probe.capability === 'block'" :size="20" />
            <Zap v-else :size="20" />
            <h3>{{ probe.capability === 'block' ? 'Active Defense Capability' : 'Monitoring Capability' }}</h3>
          </div>
          <p class="capability-desc" v-if="probe.capability === 'block'">
            This LSM hook can <strong>actively block</strong> malicious operations by returning 
            <code>-EPERM</code>. When a rule with <code>action: block</code> matches, the kernel 
            denies the operation before it can execute.
          </p>
          <p class="capability-desc" v-else>
            This hook monitors operations and emits events to userspace for analysis and alerting.
          </p>
        </section>

        <!-- BPF Source Code -->
        <section class="card-section">
          <h3 class="section-title">
            <Code :size="16" class="section-icon" />
            eBPF LSM Implementation
          </h3>
          <div class="code-block">
            <pre><code>{{ probe.sourceCode }}</code></pre>
          </div>
        </section>

        <!-- Kernel Structures -->
        <section class="card-section">
          <h3 class="section-title">
            <Database :size="16" class="section-icon" />
            Kernel Structures Accessed
          </h3>
          <div class="struct-list">
            <div v-for="struct in probe.kernelStructs" :key="struct" class="struct-item">
              <span class="struct-arrow">→</span>
              <code>{{ struct }}</code>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<style scoped>
.probe-card-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 40px;
  backdrop-filter: blur(4px);
}

.probe-card {
  width: 100%;
  max-width: 750px;
  max-height: calc(100vh - 80px);
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-default);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 20px 24px;
  background: var(--bg-elevated);
  border-bottom: 1px solid var(--border-subtle);
}

.header-info {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.probe-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  flex-shrink: 0;
}

.probe-icon.process {
  background: rgba(96, 165, 250, 0.15);
  color: var(--chart-exec);
}

.probe-icon.network {
  background: rgba(245, 158, 11, 0.15);
  color: var(--chart-network);
}

.probe-icon.file {
  background: rgba(16, 185, 129, 0.15);
  color: var(--chart-file);
}

.probe-title {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.probe-badges {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.hook-name {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--accent-primary);
  background: var(--bg-void);
  padding: 4px 8px;
  border-radius: var(--radius-sm);
}

.hook-type {
  font-size: 10px;
  font-weight: 600;
  color: #8b5cf6;
  background: rgba(139, 92, 246, 0.15);
  padding: 3px 8px;
  border-radius: var(--radius-sm);
}

.capability {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: var(--radius-sm);
}

.capability.block {
  background: var(--status-blocked);
  color: #fff;
}

.capability.monitor {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.close-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.close-btn:hover {
  background: var(--bg-overlay);
  color: var(--text-primary);
}

.card-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.card-section {
  margin-bottom: 24px;
}

.card-section:last-child {
  margin-bottom: 0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  margin: 0 0 12px 0;
}

.section-icon {
  color: var(--text-muted);
}

.description {
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-secondary);
  margin: 0;
}

/* Capability Info Box */
.capability-info {
  padding: 16px 20px;
  border-radius: var(--radius-md);
  margin-bottom: 24px;
}

.capability-info.block {
  background: linear-gradient(135deg, var(--status-blocked-dim), transparent 60%);
  border: 1px solid var(--status-blocked);
}

.capability-info.monitor {
  background: linear-gradient(135deg, var(--status-warning-dim), transparent 60%);
  border: 1px solid var(--status-warning);
}

.capability-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.capability-info.block .capability-header {
  color: var(--status-blocked);
}

.capability-info.monitor .capability-header {
  color: var(--status-warning);
}

.capability-header h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.capability-desc {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
  margin: 0;
}

.capability-desc code {
  background: var(--bg-void);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-size: 11px;
}

.code-block {
  background: var(--bg-void);
  border-radius: var(--radius-md);
  overflow-x: auto;
}

.code-block pre {
  margin: 0;
  padding: 16px;
}

.code-block code {
  font-family: var(--font-mono);
  font-size: 11px;
  line-height: 1.6;
  color: var(--text-secondary);
  white-space: pre;
}

.struct-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.struct-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
}

.struct-arrow {
  color: var(--accent-primary);
  font-weight: bold;
}

.struct-item code {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-primary);
}
</style>

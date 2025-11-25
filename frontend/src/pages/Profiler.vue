<script setup lang="ts">
import Card from '../components/common/Card.vue'
</script>

<template>
  <div class="profiler-page">
    <div class="page-header">
      <h1 class="page-title">Behavior Profiler</h1>
      <span class="page-subtitle">Learning mode and whitelist generation</span>
    </div>
    <Card>
      <div class="placeholder">
        <span class="placeholder-icon">🧪</span>
        <span class="placeholder-text">Profiler lab coming soon</span>
      </div>
    </Card>
  </div>
</template>

<style scoped>
.page-header { margin-bottom: 24px; }
.page-title { font-size: 24px; font-weight: 700; color: var(--text-primary); margin: 0; }
.page-subtitle { font-size: 14px; color: var(--text-muted); }
.placeholder { display: flex; flex-direction: column; align-items: center; gap: 16px; padding: 80px; }
.placeholder-icon { font-size: 48px; }
.placeholder-text { color: var(--text-muted); }
</style>


<script setup lang="ts">
import Card from '../components/common/Card.vue'
</script>

<template>
  <div class="rules-page">
    <div class="page-header">
      <h1 class="page-title">Rules</h1>
      <span class="page-subtitle">Detection rule management</span>
    </div>
    <Card>
      <div class="placeholder">
        <span class="placeholder-icon">📝</span>
        <span class="placeholder-text">Rules editor coming soon</span>
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


<!-- Sentinel Preview Component - Phase 4 -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Sparkles, ArrowRight } from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import { useSentinel } from '../../composables/useSentinel'
import InsightCard from '../sentinel/InsightCard.vue'

const router = useRouter()
const { insights, fetchInsights } = useSentinel()

const recentInsights = ref(insights.value.slice(0, 3))

onMounted(async () => {
  await fetchInsights()
  recentInsights.value = insights.value.slice(0, 3)
})

const goToSentinel = () => {
  router.push('/sentinel')
}
</script>

<template>
  <div class="sentinel-preview">
    <div class="preview-header">
      <div class="header-left">
        <Sparkles :size="20" />
        <h3>AI Sentinel</h3>
      </div>
      <button class="view-all-btn" @click="goToSentinel">
        View All
        <ArrowRight :size="14" />
      </button>
    </div>

    <div v-if="recentInsights.length === 0" class="empty-state">
      <Sparkles :size="32" />
      <p>No insights yet</p>
      <p class="subtitle">AI Sentinel is monitoring your system</p>
    </div>

    <div v-else class="insights-preview">
      <InsightCard
        v-for="insight in recentInsights"
        :key="insight.id"
        :insight="insight"
        @action="() => {}"
        @askAI="() => {}"
      />
    </div>
  </div>
</template>

<style scoped>
.sentinel-preview {
  padding: 20px;
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.preview-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.view-all-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.view-all-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-default);
}

.empty-state {
  padding: 60px 20px;
  text-align: center;
  color: var(--text-muted);
}

.empty-state p {
  margin-top: 16px;
  font-size: 14px;
}

.subtitle {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}

.insights-preview {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>


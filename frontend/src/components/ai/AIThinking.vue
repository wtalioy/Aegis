<!-- AI Thinking Animation Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { Sparkles } from 'lucide-vue-next'

const props = defineProps<{
  message?: string
  size?: 'sm' | 'md' | 'lg'
}>()

const sizeClass = computed(() => `size-${props.size || 'md'}`)
</script>

<template>
  <div class="thinking-container" :class="sizeClass">
    <div class="thinking-animation">
      <Sparkles :size="size === 'sm' ? 16 : size === 'lg' ? 24 : 20" class="sparkle sparkle-1" />
      <Sparkles :size="size === 'sm' ? 16 : size === 'lg' ? 24 : 20" class="sparkle sparkle-2" />
      <Sparkles :size="size === 'sm' ? 16 : size === 'lg' ? 24 : 20" class="sparkle sparkle-3" />
    </div>
    <div v-if="message" class="thinking-message">{{ message }}</div>
  </div>
</template>

<style scoped>
.thinking-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 20px;
}

.thinking-animation {
  position: relative;
  width: 60px;
  height: 60px;
}

.sparkle {
  position: absolute;
  color: var(--chart-network);
  opacity: 0;
}

.sparkle-1 {
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  animation: sparkle-1 1.5s ease-in-out infinite;
}

.sparkle-2 {
  bottom: 0;
  left: 0;
  animation: sparkle-2 1.5s ease-in-out infinite;
}

.sparkle-3 {
  bottom: 0;
  right: 0;
  animation: sparkle-3 1.5s ease-in-out infinite;
}

@keyframes sparkle-1 {
  0%, 100% { opacity: 0; transform: translateX(-50%) translateY(0) scale(0.8); }
  50% { opacity: 1; transform: translateX(-50%) translateY(-10px) scale(1); }
}

@keyframes sparkle-2 {
  0%, 100% { opacity: 0; transform: translate(0, 0) scale(0.8); }
  50% { opacity: 1; transform: translate(-10px, 10px) scale(1); }
}

@keyframes sparkle-3 {
  0%, 100% { opacity: 0; transform: translate(0, 0) scale(0.8); }
  50% { opacity: 1; transform: translate(10px, 10px) scale(1); }
}

.thinking-message {
  font-size: 14px;
  color: var(--text-secondary);
  text-align: center;
}

.size-sm .thinking-animation {
  width: 40px;
  height: 40px;
}

.size-lg .thinking-animation {
  width: 80px;
  height: 80px;
}
</style>


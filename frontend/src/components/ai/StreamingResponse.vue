<!-- Streaming Response Component - Phase 4 -->
<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { Loader2 } from 'lucide-vue-next'

const props = defineProps<{
  stream: ReadableStream<Uint8Array> | null
  onComplete?: (text: string) => void
}>()

const emit = defineEmits<{
  complete: [text: string]
}>()

const content = ref('')
const isStreaming = ref(false)

const processStream = async () => {
  if (!props.stream) return

  isStreaming.value = true
  content.value = ''

  try {
    const reader = props.stream.getReader()
    const decoder = new TextDecoder()

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      const chunk = decoder.decode(value, { stream: true })
      content.value += chunk
    }

    isStreaming.value = false
    if (props.onComplete) {
      props.onComplete(content.value)
    }
    emit('complete', content.value)
  } catch (error) {
    console.error('Stream error:', error)
    isStreaming.value = false
  }
}

watch(() => props.stream, (newStream) => {
  if (newStream) {
    processStream()
  }
}, { immediate: true })
</script>

<template>
  <div class="streaming-response">
    <div class="streaming-content">
      <div v-if="content" class="content-text">{{ content }}</div>
      <div v-if="isStreaming" class="streaming-indicator">
        <Loader2 :size="14" class="spin" />
        <span>AI is thinking...</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.streaming-response {
  padding: 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.streaming-content {
  min-height: 40px;
}

.content-text {
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

.streaming-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-muted);
  font-size: 13px;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>


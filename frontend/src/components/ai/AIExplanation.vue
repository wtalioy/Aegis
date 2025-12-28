<!-- AI Explanation Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { MessageSquare, ChevronDown, ChevronUp } from 'lucide-vue-next'
import { marked } from 'marked'

// Configure marked for consistent output
marked.setOptions({
  breaks: true
})

const props = defineProps<{
  explanation: string
  rootCause?: string
  visible?: boolean
  expandable?: boolean
  fit?: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const expanded = defineModel<boolean>('expanded', { default: true })

const normalizeMarkdown = (text: string) => {
  if (!text) return ''
  // Convert leading bullet characters to proper markdown lists
  return text
    .split('\n')
    .map(line => line.replace(/^\s*[•\u2022]\s+/,'- '))
    .join('\n')
}

const html = computed(() => props.explanation ? (marked.parse(normalizeMarkdown(props.explanation)) as string) : '')
const rootHtml = computed(() => props.rootCause ? (marked.parse(normalizeMarkdown(props.rootCause)) as string) : '')
</script>

<template>
  <div class="explanation-container" :class="{ visible: visible !== false, fit }">
    <div class="explanation-header" @click="expandable && (expanded = !expanded)">
      <div class="header-content">
        <MessageSquare :size="16" />
        <span class="header-title">AI Explanation</span>
      </div>
      <div class="header-actions">
        <button v-if="expandable" class="expand-btn" aria-label="Toggle">
          <ChevronDown v-if="!expanded" :size="16" />
          <ChevronUp v-else :size="16" />
        </button>
      </div>
    </div>

    <Transition name="expand">
      <div v-if="!expandable || expanded" class="explanation-body">
        <div class="explanation-text markdown" v-html="html" />
        <div v-if="rootCause" class="root-cause">
          <div class="root-cause-title">Root Cause:</div>
          <div class="root-cause-text markdown" v-html="rootHtml" />
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.explanation-container {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: border-color 0.2s ease, background 0.2s ease;
}

.explanation-container.visible {
  border-color: var(--border-default);
  box-shadow: none;
}

.explanation-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  background: transparent;
  border-bottom: 1px solid var(--border-subtle);
  transition: background 0.15s, border-color 0.15s;
}

.explanation-header:hover {
  background: var(--bg-hover);
}

.header-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.expand-btn,
.close-btn {
  padding: 4px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all 0.15s;
}

.expand-btn:hover,
.close-btn:hover {
  background: var(--bg-void);
  color: var(--text-primary);
}

.explanation-body {
  position: relative;
  padding: 14px 16px 14px 22px; /* slight left pad so bullets don't hug the border */
  border-top: 1px solid var(--border-subtle);
  background: transparent;
  border-bottom-left-radius: var(--radius-lg);
  border-bottom-right-radius: var(--radius-lg);
  max-height: 42vh; /* fixed so the panel below stays visible */
  overflow-y: auto;
  min-height: 0;
}
/* when fit flag is set, allow a bit more height but still constrained */
.explanation-container.fit .explanation-body { max-height: 45vh; }

/* Markdown styles */
.markdown {
  font-size: 11.5px;
  line-height: 1.42;
  color: var(--text-secondary);
}
.markdown h1, .markdown h2, .markdown h3, .markdown h4 {
  color: var(--text-primary);
  margin: 6px 0 4px;
}
.markdown h1 { font-size: 13.5px; font-weight: 700; }
.markdown h2 { font-size: 12.5px; font-weight: 700; }
.markdown h3 { font-size: 12px; font-weight: 700; }
.markdown p { margin: 4px 0; }
.markdown ul, .markdown ol { margin: 10px 0; padding-left: 26px; }
.markdown ul { list-style: disc outside; }
.markdown ol { list-style: decimal outside; }
.markdown ul ul, .markdown ol ol { margin-top: 6px; margin-bottom: 6px; padding-left: 22px; }
.markdown li { margin: 6px 0; }
.markdown li > p { margin: 0; }
.markdown li::marker { color: var(--text-muted); }
.markdown strong { color: var(--text-primary); }
.markdown code { background: var(--bg-void); padding: 2px 6px; border-radius: 4px; font-family: var(--font-mono); }
.markdown pre { background: var(--bg-void); padding: 12px; border-radius: var(--radius-md); overflow: auto; }
.markdown pre code { background: transparent; padding: 0; }

/* Subtle custom scrollbar inside explanation body */
.explanation-body::-webkit-scrollbar { width: 8px; }
.explanation-body::-webkit-scrollbar-thumb { background: var(--border-subtle); border-radius: 8px; }
.explanation-body::-webkit-scrollbar-thumb:hover { background: var(--border-default); }

.root-cause {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.root-cause-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.root-cause-text {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-primary);
  padding: 12px;
  background: var(--bg-void);
  border-radius: var(--radius-md);
}

.expand-enter-active,
.expand-leave-active {
  transition: all 0.2s ease;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
}
</style>


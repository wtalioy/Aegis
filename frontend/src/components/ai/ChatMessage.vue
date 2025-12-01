<script setup lang="ts">
import { computed, watch } from 'vue'
import { User, Sparkles } from 'lucide-vue-next'
import { marked } from 'marked'
import type { ChatMessage } from '../../lib/api'

const props = defineProps<{
    message: ChatMessage
    isStreaming?: boolean
}>()

const emit = defineEmits<{
    (e: 'streaming-update'): void
}>()

const isUser = computed(() => props.message.role === 'user')
const formattedTime = computed(() => {
    const date = new Date(props.message.timestamp)
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
})

// Configure marked for clean output
marked.setOptions({
    breaks: true,
    gfm: true
})

// Render markdown safely
function renderMarkdown(text: string): string {
    if (!text) return ''
    try {
        return marked.parse(text) as string
    } catch {
        return text
    }
}

// Watch content changes during streaming and emit scroll updates
watch(() => props.message.content, () => {
    if (props.isStreaming) {
        emit('streaming-update')
    }
})

const renderedContent = computed(() => {
    if (isUser.value) return props.message.content
    return renderMarkdown(props.message.content)
})
</script>

<template>
    <div class="message" :class="{ user: isUser, assistant: !isUser }">
        <div class="avatar">
            <component :is="isUser ? User : Sparkles" :size="14" />
        </div>
        <div class="content">
            <div class="meta">
                <span class="role">{{ isUser ? 'You' : 'EulerGuard AI' }}</span>
                <span class="time">{{ formattedTime }}</span>
            </div>
            <div v-if="isUser" class="text">{{ message.content }}</div>
            <div v-else class="text markdown" v-html="renderedContent" />
        </div>
    </div>
</template>

<style scoped>
.message {
    display: flex;
    gap: 12px;
}

.message.user {
    flex-direction: row-reverse;
}

.avatar {
    width: 28px;
    height: 28px;
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    color: var(--text-muted);
}

.content {
    max-width: 85%;
    position: relative;
}

.meta {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 4px;
}

.role {
    font-size: 12px;
    font-weight: 500;
    color: var(--text-secondary);
}

.time {
    font-size: 10px;
    color: var(--text-muted);
}

.text {
    padding: 12px 16px;
    border-radius: var(--radius-lg);
    font-size: 14px;
    line-height: 1.7;
}

.message.user .text {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    color: var(--text-primary);
}

.message.assistant .text {
    background: transparent;
    color: var(--text-primary);
    padding: 8px 0;
}

/* Markdown Styles */
.text.markdown :deep(h1),
.text.markdown :deep(h2),
.text.markdown :deep(h3),
.text.markdown :deep(h4) {
    margin: 16px 0 8px;
    color: var(--text-primary);
    font-weight: 600;
    line-height: 1.4;
}

.text.markdown :deep(h1) {
    font-size: 18px;
}

.text.markdown :deep(h2) {
    font-size: 16px;
}

.text.markdown :deep(h3) {
    font-size: 15px;
}

.text.markdown :deep(h4) {
    font-size: 14px;
}

.text.markdown :deep(p) {
    margin: 8px 0;
}

.text.markdown :deep(ul),
.text.markdown :deep(ol) {
    margin: 8px 0;
    padding-left: 24px;
}

.text.markdown :deep(li) {
    margin: 6px 0;
    line-height: 1.6;
}

.text.markdown :deep(li)::marker {
    color: var(--text-muted);
}

.text.markdown :deep(code) {
    background: var(--bg-elevated);
    padding: 2px 6px;
    border-radius: 4px;
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--accent-primary);
}

.text.markdown :deep(pre) {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 12px 16px;
    margin: 12px 0;
    overflow-x: auto;
}

.text.markdown :deep(pre code) {
    background: transparent;
    padding: 0;
    font-size: 12px;
    color: var(--text-primary);
}

.text.markdown :deep(blockquote) {
    border-left: 3px solid var(--accent-primary);
    margin: 12px 0;
    padding: 8px 16px;
    background: var(--bg-elevated);
    border-radius: 0 var(--radius-md) var(--radius-md) 0;
    color: var(--text-secondary);
}

.text.markdown :deep(strong) {
    color: var(--text-primary);
    font-weight: 600;
}

.text.markdown :deep(em) {
    color: var(--text-secondary);
}

.text.markdown :deep(a) {
    color: var(--accent-primary);
    text-decoration: none;
}

.text.markdown :deep(a:hover) {
    text-decoration: underline;
}

.text.markdown :deep(hr) {
    border: none;
    border-top: 1px solid var(--border-subtle);
    margin: 16px 0;
}

.text.markdown :deep(table) {
    width: 100%;
    border-collapse: collapse;
    margin: 12px 0;
    font-size: 13px;
}

.text.markdown :deep(th),
.text.markdown :deep(td) {
    border: 1px solid var(--border-subtle);
    padding: 8px 12px;
    text-align: left;
}

.text.markdown :deep(th) {
    background: var(--bg-elevated);
    font-weight: 600;
}
</style>

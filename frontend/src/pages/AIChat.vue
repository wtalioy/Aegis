<script setup lang="ts">
import { ref, watch, nextTick, onMounted, computed } from 'vue'
import {
    ArrowUp, Trash2, Download, Copy, Check,
    Sparkles, Activity, AlertTriangle, Box, ChevronDown, Zap
} from 'lucide-vue-next'
import { useAIChat } from '../composables/useAIChat'
import { getAIStatus, getWorkloads, type AIStatus, type Workload } from '../lib/api'
import ChatMessage from '../components/ai/ChatMessage.vue'

const {
    messages,
    isLoading,
    error,
    lastContextSummary,
    hasMessages,
    sendMessage,
    clearChat,
    isStreamingMessage
} = useAIChat()

// State
const aiStatus = ref<AIStatus | null>(null)
const workloads = ref<Workload[]>([])
const inputText = ref('')
const messagesContainer = ref<HTMLElement | null>(null)
const selectedWorkload = ref<string>('all')
const workloadDropdownOpen = ref(false)
const copied = ref(false)

// Computed
const isAIReady = computed(() => aiStatus.value?.enabled && aiStatus.value?.status === 'ready')

const selectedWorkloadLabel = computed(() => {
    if (selectedWorkload.value === 'all') return 'All Workloads'
    const w = workloads.value.find(w => w.id === selectedWorkload.value)
    return w ? (w.cgroupPath.split('/').pop() || w.id) : 'All Workloads'
})

onMounted(async () => {
    try {
        const [status, wl] = await Promise.all([
            getAIStatus(),
            getWorkloads()
        ])
        aiStatus.value = status
        workloads.value = wl
    } catch (e) {
        console.error('Failed to fetch initial data:', e)
    }

    // Close dropdown when clicking outside
    document.addEventListener('click', (e) => {
        const target = e.target as HTMLElement
        if (!target.closest('.workload-dropdown')) {
            workloadDropdownOpen.value = false
        }
    })
})

function selectWorkload(id: string) {
    selectedWorkload.value = id
    workloadDropdownOpen.value = false
}

// Scroll to bottom helper
function scrollToBottom() {
    if (messagesContainer.value) {
        messagesContainer.value.scrollTo({
            top: messagesContainer.value.scrollHeight,
            behavior: 'smooth'
        })
    }
}

// Auto-scroll to bottom when new messages arrive
watch(messages, async () => {
    await nextTick()
    scrollToBottom()
}, { deep: true })

async function handleSend() {
    if (!inputText.value.trim() || isLoading.value) return

    const message = inputText.value
    inputText.value = ''
    await sendMessage(message)
}

function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        handleSend()
    }
}

async function exportChat() {
    const content = messages.value.map(m =>
        `[${m.role.toUpperCase()}] ${new Date(m.timestamp).toLocaleString()}\n${m.content}`
    ).join('\n\n---\n\n')

    const blob = new Blob([content], { type: 'text/markdown' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `eulerguard-chat-${Date.now()}.md`
    a.click()
    URL.revokeObjectURL(url)
}

async function copyLastResponse() {
    const lastAssistant = [...messages.value].reverse().find(m => m.role === 'assistant')
    if (lastAssistant) {
        await navigator.clipboard.writeText(lastAssistant.content)
        copied.value = true
        setTimeout(() => copied.value = false, 2000)
    }
}

const quickActions = [
    { text: "System health check", icon: Activity, color: 'green' },
    { text: "Security assessment", icon: AlertTriangle, color: 'orange' },
    { text: "Analyze recent alerts", icon: Zap, color: 'red' },
    { text: "Top resource consumers", icon: Box, color: 'blue' },
]

const suggestionQuestions = [
    { text: "What's causing the high CPU load?", category: 'performance' },
    { text: "Are there any blocked security events?", category: 'security' },
    { text: "Explain the reverse shell alerts", category: 'security' },
    { text: "Which containers are most active?", category: 'workloads' },
    { text: "Summarize network connections", category: 'network' },
    { text: "Any suspicious file access?", category: 'security' },
]
</script>

<template>
    <div class="ai-chat-page">
        <!-- Main Chat Area -->
        <div class="chat-main">
            <!-- Chat Header -->
            <div class="chat-header">
                <div class="header-left">
                    <div class="title-section">
                        <div class="ai-avatar">
                            <Sparkles :size="18" />
                        </div>
                        <div>
                            <h1>EulerGuard AI</h1>
                            <div class="provider-info" v-if="aiStatus?.enabled">
                                <span>{{ aiStatus.provider }}</span>
                                <span class="status-dot" :class="aiStatus.status"></span>
                                <span class="status-text" :class="aiStatus.status">
                                    {{ aiStatus.status === 'ready' ? 'Online' : 'Offline' }}
                                </span>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="header-actions">
                    <!-- Workload Filter Dropdown -->
                    <div class="workload-dropdown" v-if="workloads.length > 0">
                        <button class="dropdown-trigger" @click.stop="workloadDropdownOpen = !workloadDropdownOpen">
                            <Box :size="14" />
                            <span>{{ selectedWorkloadLabel }}</span>
                            <ChevronDown :size="14" :class="{ rotated: workloadDropdownOpen }" />
                        </button>
                        <Transition name="dropdown">
                            <div v-if="workloadDropdownOpen" class="dropdown-menu">
                                <button class="dropdown-item" :class="{ active: selectedWorkload === 'all' }"
                                    @click="selectWorkload('all')">
                                    All Workloads
                                </button>
                                <button v-for="w in workloads" :key="w.id" class="dropdown-item"
                                    :class="{ active: selectedWorkload === w.id }" @click="selectWorkload(w.id)">
                                    {{ w.cgroupPath.split('/').pop() || w.id }}
                                </button>
                            </div>
                        </Transition>
                    </div>

                    <!-- Action Buttons -->
                    <button v-if="hasMessages" class="header-btn" @click="copyLastResponse" title="Copy last response">
                        <component :is="copied ? Check : Copy" :size="16" />
                    </button>
                    <button v-if="hasMessages" class="header-btn" @click="exportChat" title="Export conversation">
                        <Download :size="16" />
                    </button>
                    <button v-if="hasMessages" class="header-btn danger" @click="clearChat" title="New conversation">
                        <Trash2 :size="16" />
                    </button>
                </div>
            </div>

            <!-- Messages Container -->
            <div class="messages-wrapper">
                <div ref="messagesContainer" class="messages-container">
                    <!-- Empty State -->
                    <div v-if="!hasMessages && !isLoading" class="empty-state">
                        <div class="welcome-section">
                            <div class="welcome-icon">
                                <Sparkles :size="32" />
                            </div>
                            <h2>How can I help you today?</h2>
                            <p>I have real-time access to your kernel telemetry via eBPF</p>
                        </div>

                        <!-- Quick Actions -->
                        <div class="quick-actions">
                            <button v-for="action in quickActions" :key="action.text" class="quick-action-btn"
                                :class="action.color" @click="sendMessage(action.text)">
                                <component :is="action.icon" :size="18" />
                                <span>{{ action.text }}</span>
                            </button>
                        </div>

                        <!-- Suggestions Grid -->
                        <div class="suggestions-section">
                            <div class="suggestions-grid">
                                <button v-for="q in suggestionQuestions" :key="q.text" class="suggestion-btn"
                                    @click="sendMessage(q.text)">
                                    <span>{{ q.text }}</span>
                                    <span class="suggestion-arrow">→</span>
                                </button>
                            </div>
                        </div>
                    </div>

                    <!-- Messages -->
                    <template v-if="hasMessages">
                        <ChatMessage v-for="(msg, i) in messages" :key="i" :message="msg"
                            :is-streaming="isStreamingMessage(i)" @streaming-update="scrollToBottom" />
                    </template>

                    <!-- Loading -->
                    <div v-if="isLoading" class="typing-indicator">
                        <div class="typing-avatar">
                            <Sparkles :size="14" />
                        </div>
                        <div class="typing-dots">
                            <span></span>
                            <span></span>
                            <span></span>
                        </div>
                    </div>

                    <!-- Error -->
                    <div v-if="error" class="error-toast">
                        <AlertTriangle :size="16" />
                        <span>{{ error }}</span>
                        <button @click="error = null">×</button>
                    </div>
                </div>
            </div>

            <!-- Input Area -->
            <div class="input-section">
                <div class="input-container">
                    <textarea v-model="inputText" placeholder="Message EulerGuard..." rows="1" @keydown="handleKeydown"
                        :disabled="isLoading || !isAIReady" />
                    <button class="send-btn" @click="handleSend"
                        :disabled="!inputText.trim() || isLoading || !isAIReady">
                        <ArrowUp :size="18" />
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.ai-chat-page {
    display: flex;
    flex-direction: column;
    height: calc(100vh - var(--topbar-height) - var(--footer-height) - 48px);
    margin: -24px;
}

/* Main Chat Area */
.chat-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
    width: 100%;
}

.chat-header {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 24px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
}

.header-left {
    display: flex;
    align-items: center;
    gap: 12px;
}

.title-section {
    display: flex;
    align-items: center;
    gap: 10px;
}

.ai-avatar {
    width: 32px;
    height: 32px;
    border-radius: var(--radius-md);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
}

.title-section h1 {
    font-size: 15px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
}

.provider-info {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: var(--text-muted);
    margin-top: 1px;
}

.status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--text-muted);
}

.status-dot.ready {
    background: var(--status-safe);
}

.status-dot.unavailable {
    background: var(--status-critical);
}

.status-text {
    font-size: 11px;
}

.status-text.ready {
    color: var(--status-safe);
}

.status-text.unavailable {
    color: var(--status-critical);
}

.header-actions {
    display: flex;
    align-items: center;
    gap: 6px;
}

/* Custom Dropdown */
.workload-dropdown {
    position: relative;
}

.dropdown-trigger {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 4px 8px;
    background: transparent;
    border: none;
    border-radius: var(--radius-sm);
    font-size: 12px;
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.15s;
}

.dropdown-trigger:hover {
    background: var(--bg-hover);
    color: var(--text-secondary);
}

.dropdown-trigger svg:last-child {
    transition: transform 0.15s;
}

.dropdown-trigger svg.rotated {
    transform: rotate(180deg);
}

.dropdown-menu {
    position: absolute;
    top: calc(100% + 4px);
    right: 0;
    min-width: 160px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
    z-index: 100;
    overflow: hidden;
}

.dropdown-item {
    display: block;
    width: 100%;
    padding: 8px 12px;
    background: none;
    border: none;
    font-size: 12px;
    color: var(--text-secondary);
    text-align: left;
    cursor: pointer;
    transition: all 0.1s;
}

.dropdown-item:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
}

.dropdown-item.active {
    background: var(--accent-glow);
    color: var(--accent-primary);
}

.dropdown-enter-active,
.dropdown-leave-active {
    transition: all 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
    opacity: 0;
    transform: translateY(-4px);
}

.header-btn {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    border-radius: var(--radius-md);
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.15s;
}

.header-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
}

.header-btn.danger:hover {
    color: var(--status-critical);
}

/* Messages */
.messages-wrapper {
    flex: 1;
    min-height: 0;
    overflow: hidden;
}

.messages-container {
    height: 100%;
    overflow-y: auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.messages-container> :not(.empty-state) {
    max-width: 800px;
    width: 100%;
    margin: 0 auto;
}

/* Empty State */
.empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    max-width: 700px;
    margin: 0 auto;
    padding: 40px 24px;
}

.welcome-section {
    text-align: center;
    margin-bottom: 28px;
}

.welcome-icon {
    width: 48px;
    height: 48px;
    border-radius: var(--radius-lg);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
    margin: 0 auto 14px;
}

.welcome-section h2 {
    font-size: 20px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 6px;
}

.welcome-section p {
    color: var(--text-muted);
    font-size: 13px;
    margin: 0;
}

/* Quick Actions */
.quick-actions {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
    width: 100%;
    max-width: 520px;
    margin-bottom: 24px;
}

.quick-action-btn {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 14px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-secondary);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s;
}

.quick-action-btn:hover {
    background: var(--bg-hover);
    border-color: var(--border-default);
    color: var(--text-primary);
}

.quick-action-btn.green:hover {
    border-color: var(--status-safe);
    color: var(--status-safe);
}

.quick-action-btn.orange:hover {
    border-color: var(--status-warning);
    color: var(--status-warning);
}

.quick-action-btn.red:hover {
    border-color: var(--status-critical);
    color: var(--status-critical);
}

.quick-action-btn.blue:hover {
    border-color: var(--accent-primary);
    color: var(--accent-primary);
}

/* Suggestions */
.suggestions-section {
    width: 100%;
    max-width: 600px;
}

.suggestions-grid {
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.suggestion-btn {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-secondary);
    font-size: 12px;
    text-align: left;
    cursor: pointer;
    transition: all 0.15s;
}

.suggestion-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
    border-color: var(--border-default);
}

.suggestion-arrow {
    opacity: 0;
    transform: translateX(-4px);
    transition: all 0.15s;
    color: var(--text-muted);
}

.suggestion-btn:hover .suggestion-arrow {
    opacity: 1;
    transform: translateX(0);
}

/* Typing Indicator */
.typing-indicator {
    display: flex;
    align-items: flex-start;
    gap: 10px;
}

.typing-avatar {
    width: 28px;
    height: 28px;
    border-radius: var(--radius-md);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
    flex-shrink: 0;
}

.typing-dots {
    display: flex;
    align-items: center;
    gap: 3px;
    padding: 12px 14px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
}

.typing-dots span {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--text-muted);
    animation: typing 1.4s infinite;
}

.typing-dots span:nth-child(2) {
    animation-delay: 0.2s;
}

.typing-dots span:nth-child(3) {
    animation-delay: 0.4s;
}

@keyframes typing {

    0%,
    60%,
    100% {
        transform: translateY(0);
        opacity: 0.4;
    }

    30% {
        transform: translateY(-3px);
        opacity: 1;
    }
}

/* Error Toast */
.error-toast {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 14px;
    background: var(--status-critical-dim);
    border: 1px solid rgba(239, 68, 68, 0.2);
    border-radius: var(--radius-md);
    color: var(--status-critical);
    font-size: 12px;
}

.error-toast button {
    margin-left: auto;
    background: none;
    border: none;
    color: inherit;
    font-size: 16px;
    cursor: pointer;
    line-height: 1;
    opacity: 0.7;
}

.error-toast button:hover {
    opacity: 1;
}

/* Input Section */
.input-section {
    flex-shrink: 0;
    padding: 12px 24px 12px;
}

.input-container {
    display: flex;
    gap: 10px;
    align-items: flex-end;
    max-width: 800px;
    margin: 0 auto;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    padding: 10px 10px 10px 16px;
    transition: border-color 0.15s;
}

.input-container:focus-within {
    border-color: var(--border-default);
}

.input-container textarea {
    flex: 1;
    background: none;
    border: none;
    color: var(--text-primary);
    font-size: 14px;
    resize: none;
    min-height: 22px;
    max-height: 120px;
    line-height: 1.5;
    padding: 6px 0;
}

.input-container textarea:focus {
    outline: none;
}

.input-container textarea::placeholder {
    color: var(--text-muted);
}

.send-btn {
    width: 32px;
    height: 32px;
    background: var(--text-primary);
    border: none;
    border-radius: var(--radius-md);
    color: var(--bg-void);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
    flex-shrink: 0;
}

.send-btn:hover:not(:disabled) {
    opacity: 0.9;
}

.send-btn:disabled {
    background: var(--text-muted);
    opacity: 0.4;
    cursor: not-allowed;
}
</style>

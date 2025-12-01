<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { X, Sparkles, Loader2, AlertTriangle } from 'lucide-vue-next'
import { diagnoseSystem, type DiagnosisResult } from '../../lib/api'
import { marked } from 'marked'

const props = defineProps<{
    visible: boolean
}>()

const emit = defineEmits<{
    (e: 'close'): void
}>()

const loading = ref(false)
const error = ref<string | null>(null)
const result = ref<DiagnosisResult | null>(null)

marked.setOptions({
    gfm: true,
    breaks: true
})

async function runDiagnosis() {
    loading.value = true
    error.value = null

    try {
        result.value = await diagnoseSystem()
    } catch (err) {
        error.value = err instanceof Error ? err.message : 'Unknown error'
    } finally {
        loading.value = false
    }
}

const renderedAnalysis = computed(() => {
    if (!result.value?.analysis) return ''
    try {
        return marked.parse(result.value.analysis) as string
    } catch {
        return result.value.analysis
    }
})

function close() {
    emit('close')
    // Reset state after animation
    setTimeout(() => {
        result.value = null
        error.value = null
    }, 300)
}

// Auto-run diagnosis when modal opens
watch(() => props.visible, (visible) => {
    if (visible && !result.value && !loading.value) {
        runDiagnosis()
    }
})
</script>

<template>
    <Teleport to="body">
        <div v-if="visible" class="modal-overlay" @click.self="close">
            <div class="modal-container">
                <div class="modal-header">
                    <div class="header-title">
                        <Sparkles :size="18" class="header-icon" />
                        <span>System Diagnosis</span>
                    </div>
                    <button class="close-btn" @click="close">
                        <X :size="18" />
                    </button>
                </div>

                <div class="modal-body">
                    <!-- Loading State -->
                    <div v-if="loading" class="loading-state">
                        <Loader2 :size="32" class="spinner" />
                        <p>Analyzing system telemetry...</p>
                    </div>

                    <!-- Error State -->
                    <div v-else-if="error" class="error-state">
                        <AlertTriangle :size="32" class="error-icon" />
                        <p>{{ error }}</p>
                        <button class="retry-btn" @click="runDiagnosis">
                            Retry
                        </button>
                    </div>

                    <!-- Result State -->
                    <div v-else-if="result" class="result-state">
                        <div class="analysis-content" v-html="renderedAnalysis">
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </Teleport>
</template>

<style scoped>
.modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn 0.2s ease;
}

.modal-container {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    width: 90%;
    max-width: 700px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    animation: slideUp 0.3s ease;
}

.modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-subtle);
}

.header-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
}

.header-icon {
    color: var(--text-secondary);
}

.close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 4px;
    border-radius: var(--radius-sm);
}

.close-btn:hover {
    background: var(--bg-elevated);
    color: var(--text-primary);
}

.modal-body {
    flex: 1;
    overflow-y: auto;
    padding: 20px;
}

.loading-state,
.error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    padding: 40px;
    text-align: center;
    color: var(--text-secondary);
}

.spinner {
    animation: spin 1s linear infinite;
    color: var(--accent-primary);
}

.error-icon {
    color: var(--status-critical);
}

.retry-btn {
    padding: 8px 16px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-size: 13px;
    cursor: pointer;
    transition: all 0.15s;
}

.retry-btn:hover {
    background: var(--bg-hover);
    border-color: var(--border-default);
}

.analysis-content {
    line-height: 1.7;
    color: var(--text-primary);
    font-size: 14px;
}

.analysis-content :deep(h1),
.analysis-content :deep(h2),
.analysis-content :deep(h3),
.analysis-content :deep(h4) {
    margin: 16px 0 8px;
    color: var(--text-primary);
    font-weight: 600;
    line-height: 1.4;
}

.analysis-content :deep(h1) {
    font-size: 18px;
}

.analysis-content :deep(h2) {
    font-size: 16px;
}

.analysis-content :deep(h3) {
    font-size: 15px;
}

.analysis-content :deep(h4) {
    font-size: 14px;
}

.analysis-content :deep(p) {
    margin: 8px 0;
}

.analysis-content :deep(ul),
.analysis-content :deep(ol) {
    margin: 8px 0;
    padding-left: 24px;
}

.analysis-content :deep(li) {
    margin: 6px 0;
    line-height: 1.6;
}

.analysis-content :deep(li)::marker {
    color: var(--text-muted);
}

.analysis-content :deep(code) {
    background: var(--bg-elevated);
    padding: 2px 6px;
    border-radius: 4px;
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--accent-primary);
}

.analysis-content :deep(pre) {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 12px 16px;
    margin: 12px 0;
    overflow-x: auto;
}

.analysis-content :deep(pre code) {
    background: transparent;
    padding: 0;
    font-size: 12px;
    color: var(--text-primary);
}

.analysis-content :deep(strong) {
    color: var(--text-primary);
    font-weight: 600;
}

.analysis-content :deep(blockquote) {
    border-left: 3px solid var(--border-default);
    margin: 12px 0;
    padding-left: 16px;
    color: var(--text-secondary);
}

.analysis-content :deep(a) {
    color: var(--accent-primary);
    text-decoration: none;
}

.analysis-content :deep(a:hover) {
    text-decoration: underline;
}

.analysis-content :deep(hr) {
    border: none;
    border-top: 1px solid var(--border-subtle);
    margin: 16px 0;
}

.analysis-content :deep(table) {
    width: 100%;
    border-collapse: collapse;
    margin: 12px 0;
    font-size: 13px;
}

.analysis-content :deep(th),
.analysis-content :deep(td) {
    border: 1px solid var(--border-subtle);
    padding: 8px 12px;
    text-align: left;
}

.analysis-content :deep(th) {
    background: var(--bg-elevated);
    font-weight: 600;
}

@keyframes fadeIn {
    from {
        opacity: 0;
    }

    to {
        opacity: 1;
    }
}

@keyframes slideUp {
    from {
        transform: translateY(20px);
        opacity: 0;
    }

    to {
        transform: translateY(0);
        opacity: 1;
    }
}

@keyframes spin {
    from {
        transform: rotate(0deg);
    }

    to {
        transform: rotate(360deg);
    }
}
</style>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { X, Sparkles, Loader2, AlertTriangle } from 'lucide-vue-next'
import { diagnoseSystem, type DiagnosisResult } from '../../lib/api'

const props = defineProps<{
    visible: boolean
}>()

const emit = defineEmits<{
    (e: 'close'): void
}>()

const loading = ref(false)
const error = ref<string | null>(null)
const result = ref<DiagnosisResult | null>(null)
const userQuery = ref('')

function escapeHtml(text: string): string {
    return text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
}

// Simple markdown renderer with HTML escaping
function renderMarkdown(text: string): string {
    const escaped = escapeHtml(text)
    return escaped
        .replace(/^### (.+)$/gm, '<h3>$1</h3>')
        .replace(/^## (.+)$/gm, '<h2>$1</h2>')
        .replace(/^# (.+)$/gm, '<h1>$1</h1>')
        .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.+?)\*/g, '<em>$1</em>')
        .replace(/`(.+?)`/g, '<code>$1</code>')
        .replace(/^- (.+)$/gm, '<li>$1</li>')
        .replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>')
        .replace(/\n\n/g, '</p><p>')
        .replace(/^/, '<p>')
        .replace(/$/, '</p>')
}

async function runDiagnosis() {
    loading.value = true
    error.value = null
    
    try {
        result.value = await diagnoseSystem(userQuery.value || undefined)
    } catch (err) {
        error.value = err instanceof Error ? err.message : 'Unknown error'
    } finally {
        loading.value = false
    }
}

function close() {
    emit('close')
    // Reset state after animation
    setTimeout(() => {
        result.value = null
        error.value = null
        userQuery.value = ''
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
                        <div class="result-meta">
                            <span class="provider-badge">
                                {{ result.provider }}
                            </span>
                            <span class="duration">{{ result.durationMs }}ms</span>
                            <span class="snapshot">{{ result.snapshotSummary }}</span>
                        </div>
                        
                        <div class="analysis-content" v-html="renderMarkdown(result.analysis)">
                        </div>
                        
                        <!-- Ask follow-up -->
                        <div class="followup-section">
                            <input 
                                v-model="userQuery" 
                                type="text" 
                                placeholder="Ask a follow-up question..."
                                @keyup.enter="runDiagnosis"
                            />
                            <button @click="runDiagnosis" :disabled="loading">
                                Ask
                            </button>
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

.result-meta {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-bottom: 16px;
    flex-wrap: wrap;
}

.provider-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    font-size: 11px;
    color: var(--text-secondary);
}

.duration,
.snapshot {
    font-size: 12px;
    color: var(--text-muted);
}

.analysis-content {
    line-height: 1.7;
    color: var(--text-primary);
}

.analysis-content :deep(h1),
.analysis-content :deep(h2),
.analysis-content :deep(h3) {
    margin: 16px 0 8px;
    color: var(--text-primary);
}

.analysis-content :deep(code) {
    background: var(--bg-elevated);
    padding: 2px 6px;
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: 13px;
}

.analysis-content :deep(ul) {
    padding-left: 20px;
}

.analysis-content :deep(li) {
    margin: 4px 0;
}

.analysis-content :deep(strong) {
    color: var(--text-primary);
    font-weight: 600;
}

.followup-section {
    display: flex;
    gap: 8px;
    margin-top: 20px;
    padding-top: 16px;
    border-top: 1px solid var(--border-subtle);
}

.followup-section input {
    flex: 1;
    padding: 10px 14px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-size: 13px;
}

.followup-section input:focus {
    outline: none;
    border-color: var(--border-default);
}

.followup-section button {
    padding: 10px 18px;
    background: var(--text-primary);
    border: none;
    border-radius: var(--radius-md);
    color: var(--bg-void);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.15s;
}

.followup-section button:hover:not(:disabled) {
    opacity: 0.9;
}

.followup-section button:disabled {
    background: var(--text-muted);
    opacity: 0.4;
    cursor: not-allowed;
}

@keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
}

@keyframes slideUp {
    from { transform: translateY(20px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
}

@keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}
</style>


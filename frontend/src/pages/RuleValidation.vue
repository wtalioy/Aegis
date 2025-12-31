<!-- Rule Validation Dashboard - Phase 5 -->
<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
    ClipboardCheck, Activity, CheckCircle2, Clock, AlertTriangle, X, Edit, Zap
} from 'lucide-vue-next'

import type { Rule } from '../types/rules'
import TestingRuleCard from '../components/rules/TestingRuleCard.vue'

import RuleLifecycleTimeline from '../components/rules/RuleLifecycleTimeline.vue'
import ReadinessChecklist from '../components/rules/ReadinessChecklist.vue'
import { getTestingRules } from '../lib/api'

const route = useRoute()
const router = useRouter()

const testingRules = ref<Rule[]>([])
const selectedRule = ref<Rule | null>(null)
const validationData = ref<any>(null)
const loading = ref(true)
const promoting = ref(false)
const error = ref<string | null>(null)
const newlyDeployedRule = ref<string | null>(null)

// Promotion criteria from config (defaults)
const promotionMinObservationMinutes = ref(1440) // 24 hours in minutes
const promotionMinHits = ref(100)

const fetchTestingRulesList = async () => {
    loading.value = true
    error.value = null
    try {
        // Use new testing endpoint with fallback handled in api.ts
        const data = await getTestingRules()
        console.log('Testing rules response:', data)
        // Backend returns array of TestingRuleResponse with { rule, validation, stats }
        // Extract unique rules by name to avoid duplicates, preserving all rule properties
        const ruleMap = new Map<string, Rule>()
        if (Array.isArray(data)) {
            data.forEach((item: any) => {
                // Handle both lowercase (JSON tag) and uppercase (Go struct) property names
                const rule = item.rule || item.Rule || item
                console.log('Processing rule item:', item, 'extracted rule:', rule)

                // Check for name in both cases
                const ruleName = rule?.name || rule?.Name
                if (rule && ruleName) {
                    // Normalize rule properties to lowercase for consistency
                    const normalizedRule: any = {
                        name: ruleName,
                        description: rule.description || rule.Description || '',
                        severity: rule.severity || rule.Severity || '',
                        action: rule.action || rule.Action || '',
                        state: rule.state || rule.State || 'testing',
                        match: rule.match || rule.Match || {},
                        type: rule.type || rule.Type,
                        ...rule // Keep all other properties
                    }

                    // Merge stats and validation data into rule for display
                    // Backend returns TestingStats with capitalized fields (Hits, ObservationMinutes, etc.)
                    const stats = item.stats || item.Stats || {}
                    console.log('Stats for rule:', ruleName, stats)
                    const enrichedRule: any = {
                        ...normalizedRule,
                        // Preserve validation stats if available - check both cases
                        actual_testing_hits: stats.hits || stats.Hits || stats.matchCount || stats.MatchCount || normalizedRule.actual_testing_hits || 0,
                        observationMinutes: stats.observationMinutes || stats.ObservationMinutes || (stats.observationHours ? Math.round(stats.observationHours * 60) : 0) || (stats.ObservationHours ? Math.round(stats.ObservationHours * 60) : 0) || (normalizedRule.observationHours ? Math.round(normalizedRule.observationHours * 60) : 0) || 0,
                        promotion_score: item.validation?.score || item.validation?.promotionScore || item.Validation?.score || item.Validation?.promotionScore || normalizedRule.promotion_score || 0,
                    }
                    console.log('Enriched rule:', ruleName, 'hits:', enrichedRule.actual_testing_hits)
                    ruleMap.set(ruleName, enrichedRule)
                } else {
                    console.warn('Skipping invalid rule item - missing name:', item, 'rule:', rule)
                }
            })
        } else {
            console.warn('Response is not an array:', data)
        }
        testingRules.value = Array.from(ruleMap.values())
        console.log('Final testing rules:', testingRules.value)
        if (testingRules.value.length > 0 && !selectedRule.value) {
            selectedRule.value = testingRules.value[0]
            await loadValidation(testingRules.value[0])
        }
    } catch (e) {
        console.error('Failed to fetch testing rules:', e)
        error.value = 'Failed to load testing rules'
    } finally {
        loading.value = false
    }
}

const loadValidation = async (rule: Rule) => {
    if (!rule || !rule.name) {
        error.value = 'Invalid rule'
        return
    }
    try {
        const encodedName = encodeURIComponent(rule.name)
        const response = await fetch(`/api/rules/validation/${encodedName}`)
        if (!response.ok) {
            const errorText = await response.text()
            throw new Error(errorText || `HTTP ${response.status}`)
        }
        validationData.value = await response.json()
        error.value = null // Clear any previous errors
    } catch (e) {
        console.error('Failed to load validation data:', e)
        error.value = `Failed to load validation data: ${e instanceof Error ? e.message : 'Unknown error'}`
    }
}

// Auto-refresh validation data every 5 seconds when a rule is selected
let refreshInterval: ReturnType<typeof setInterval> | null = null

const startAutoRefresh = () => {
    if (refreshInterval) {
        clearInterval(refreshInterval)
    }
    refreshInterval = setInterval(async () => {
        // Refresh both the testing rules list (to update hit counts in cards) and validation data
        await fetchTestingRulesList()
        if (selectedRule.value) {
            // Update selected rule from refreshed list
            const refreshed = testingRules.value.find(r => r.name === selectedRule.value!.name)
            if (refreshed) {
                selectedRule.value = refreshed
            }
            await loadValidation(selectedRule.value)
        }
    }, 5000) // Refresh every 5 seconds
}

const stopAutoRefresh = () => {
    if (refreshInterval) {
        clearInterval(refreshInterval)
        refreshInterval = null
    }
}

const handleSelectRule = async (rule: Rule) => {
    selectedRule.value = rule
    await loadValidation(rule)
    startAutoRefresh() // Start auto-refresh when rule is selected
}

// router already declared above
const handleAdjustRule = () => {
    if (!selectedRule.value) return
    router.push({ path: '/policy-studio', query: { rule: selectedRule.value.name } })
}


const handlePromote = async (force: boolean = false) => {
    if (!selectedRule.value) return

    promoting.value = true
    error.value = null
    try {
        const response = await fetch(`/api/rules/validation/${selectedRule.value.name}/promote`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ force })
        })
        if (!response.ok) {
            const errorText = await response.text()
            throw new Error(errorText || 'Failed to promote rule')
        }
        // Refresh data
        await fetchTestingRulesList()
        if (selectedRule.value) {
            await loadValidation(selectedRule.value)
        }
    } catch (e) {
        console.error('Failed to promote rule:', e)
        error.value = e instanceof Error ? e.message : 'Failed to promote rule'
    } finally {
        promoting.value = false
    }
}

const isReady = computed(() => {
    const stats = validationData.value?.stats
    const validation = validationData.value?.validation
    if (!stats && !validation) return false
    // Prefer explicit criteria if available
    const obsMinutes = stats?.observationMinutes ?? stats?.ObservationMinutes ?? 0
    const matches = stats?.hits ?? stats?.matchCount ?? stats?.matches ?? 0
    const criteriaReady = obsMinutes >= promotionMinObservationMinutes.value && matches >= promotionMinHits.value
    // Fallback to backend readiness flag if present
    return typeof validation?.is_ready === 'boolean' ? validation.is_ready : criteriaReady
})

const stats = computed(() => ({
    totalTesting: testingRules.value.length,
    readyToPromote: testingRules.value.filter(r => {
        return (r as any).promotionScore >= 0.7
    }).length,
    avgObservationTime: testingRules.value.length > 0
        ? testingRules.value.reduce((sum, r) => {
            const minutes = (r as any).observationMinutes || ((r as any).observationHours ? Math.round((r as any).observationHours * 60) : 0) || 0
            return sum + minutes
        }, 0) / testingRules.value.length / 60 // Convert to hours for display
        : 0,
}))

onMounted(async () => {
    // Fetch promotion config
    try {
        const response = await fetch('/api/settings')
        if (response.ok) {
            const settings = await response.json()
            if (settings.promotion?.minObservationMinutes) {
                promotionMinObservationMinutes.value = settings.promotion.minObservationMinutes
            }
            if (settings.promotion?.minHits) {
                promotionMinHits.value = settings.promotion.minHits
            }
        }
    } catch (e) {
        console.warn('Failed to fetch promotion config, using defaults:', e)
    }

    // Small delay to ensure backend has reloaded rules after deployment
    const qp = (route as any)?.query || {}
    const fromDeploy = typeof qp.from === 'string' && qp.from === 'deploy'
    if (fromDeploy) {
        // Wait a bit for backend to finish reloading
        await new Promise(resolve => setTimeout(resolve, 500))
    }

    await fetchTestingRulesList()
    // Deep-link: preselect a rule from ?rule=
    const ruleParam = typeof qp.rule === 'string' ? qp.rule : null

    if (ruleParam && testingRules.value.length > 0) {
        const match = testingRules.value.find(r => r.name === ruleParam)
        if (match) {
            await handleSelectRule(match)
            // Mark as newly deployed if coming from Policy Studio
            if (fromDeploy) {
                newlyDeployedRule.value = ruleParam
                // Auto-dismiss banner after 8 seconds
                setTimeout(() => {
                    newlyDeployedRule.value = null
                }, 8000)
            }
        }
    } else if (fromDeploy && testingRules.value.length === 0) {
        // If we just deployed but no rules found, try again after a short delay
        setTimeout(async () => {
            await fetchTestingRulesList()
            if (ruleParam && testingRules.value.length > 0) {
                const match = testingRules.value.find(r => r.name === ruleParam)
                if (match) {
                    await handleSelectRule(match)
                    newlyDeployedRule.value = ruleParam
                    setTimeout(() => {
                        newlyDeployedRule.value = null
                    }, 8000)
                }
            }
        }, 1000)
    }
})

onBeforeUnmount(() => {
    stopAutoRefresh()
})
</script>

<template>
    <div class="rule-validation-page">
        <!-- Header -->
        <div class="page-header">
            <div class="header-content">
                <h1 class="page-title">
                    Rule Validation
                </h1>
                <span class="page-subtitle">Test and promote rules to production</span>
            </div>

            <div class="header-stats">
                <div class="stat-card">
                    <Activity :size="16" class="stat-icon" />
                    <div class="stat-info">
                        <div class="stat-label">Testing Rules</div>
                        <div class="stat-value">{{ stats.totalTesting }}</div>
                    </div>
                </div>
                <div class="stat-card ready">
                    <CheckCircle2 :size="16" class="stat-icon" />
                    <div class="stat-info">
                        <div class="stat-label">Ready to Promote</div>
                        <div class="stat-value">{{ stats.readyToPromote }}</div>
                    </div>
                </div>
                <div class="stat-card">
                    <Clock :size="16" class="stat-icon" />
                    <div class="stat-info">
                        <div class="stat-label">Avg Observation</div>
                        <div class="stat-value">{{ stats.avgObservationTime.toFixed(1) }}h</div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Deployment Success Banner -->
        <div v-if="newlyDeployedRule" class="success-banner">
            <CheckCircle2 :size="16" class="banner-icon" />
            <div class="banner-content">
                <span class="banner-title">Rule deployed to Testing</span>
                <span class="banner-text">You just deployed <strong>{{ newlyDeployedRule }}</strong> to testing. We'll
                    compute readiness as data accumulates.</span>
            </div>
            <button class="banner-close" @click="newlyDeployedRule = null">
                <X :size="16" />
            </button>
        </div>

        <!-- Error Message -->
        <div v-if="error" class="error-banner">
            <AlertTriangle :size="16" />
            <span>{{ error }}</span>
        </div>

        <!-- Main Content -->
        <div class="validation-content">
            <!-- Left Panel: Testing Rules List -->
            <div class="rules-panel">
                <div class="panel-header">
                    <h2>Testing Rules</h2>
                    <span class="rule-count">{{ testingRules.length }}</span>
                </div>

                <div v-if="loading" class="loading-state">
                    <div class="spinner"></div>
                    <span>Loading rules...</span>
                </div>

                <div v-else-if="testingRules.length === 0" class="empty-state">
                    <ClipboardCheck :size="32" class="empty-icon" />
                    <p>No testing rules yet</p>
                    <span class="empty-hint">Create a rule and deploy it to testing to start validation</span>
                </div>

                <div v-else class="rules-list">
                    <TestingRuleCard v-for="rule in testingRules" :key="rule.name" :rule="rule"
                        :selected="selectedRule?.name === rule.name" @select="() => handleSelectRule(rule)" />
                </div>
            </div>

            <!-- Right Panel: Validation Details -->
            <div class="details-panel">
                <div v-if="!selectedRule" class="empty-state">
                    <ClipboardCheck :size="32" class="empty-icon" />
                    <p>Select a testing rule to view validation details</p>
                </div>

                <div v-else-if="validationData" class="validation-details">
                    <!-- Rule Info -->
                    <div class="rule-header">
                        <div>
                            <h3>{{ selectedRule.name }}</h3>
                            <p class="rule-description">{{ selectedRule.description }}</p>
                        </div>
                        <div class="rule-meta">
                            <span class="severity" :class="selectedRule.severity">
                                {{ selectedRule.severity.toUpperCase() }}
                            </span>
                        </div>
                    </div>

                    <!-- Lifecycle Timeline -->
                    <RuleLifecycleTimeline :rule="selectedRule" />

                    <!-- Promotion Readiness -->
                    <ReadinessChecklist :validation-data="validationData" />


                    <!-- Actions -->
                    <div class="action-buttons">
                        <button class="btn btn-primary" :disabled="!isReady || promoting"
                            @click="() => handlePromote(false)">
                            <CheckCircle2 :size="16" />
                            {{ promoting ? 'Promoting...' : 'Promote to Production' }}
                        </button>
                        <button v-if="!isReady" class="btn btn-warning" :disabled="promoting"
                            @click="() => handlePromote(true)" title="Force promote even if requirements are not met">
                            <Zap :size="16" />
                            Force Promote
                        </button>
                        <button class="btn btn-secondary" @click="handleAdjustRule">
                            <Edit :size="16" />
                            Adjust Rule
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.rule-validation-page {
    display: flex;
    flex-direction: column;
    gap: 24px;
    padding: 24px;
    height: calc(100vh - var(--topbar-height, 60px) - var(--footer-height, 0px) - 48px);
}

.page-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 16px;
}

.header-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.page-title {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 24px;
    font-weight: 600; /* Softened from 700 */
    color: var(--text-primary);
    margin: 0;
}

.title-icon {
    color: var(--accent-primary);
}

.page-subtitle {
    font-size: 14px;
    color: var(--text-muted);
}

.header-stats {
    display: flex;
    gap: 12px;
}

.stat-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
}

.stat-card.ready {
    border-color: var(--status-safe);
    background: var(--status-safe-dim);
}

.stat-icon {
    color: var(--text-secondary);
}

.stat-card.ready .stat-icon {
    color: var(--status-safe);
}

.stat-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
}

.stat-label {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.stat-value {
    font-size: 18px;
    font-weight: 600; /* Softened from 700 */
    color: var(--text-primary);
}

.success-banner {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: var(--status-safe-dim);
    border: 1px solid var(--status-safe);
    border-radius: var(--radius-md);
    color: var(--status-safe);
    font-size: 13px;
}

.banner-icon {
    flex-shrink: 0;
    color: var(--status-safe);
}

.banner-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
}

.banner-title {
    font-weight: 600;
    color: var(--status-safe);
}

.banner-text {
    font-size: 12px;
    color: var(--text-secondary);
}

.banner-close {
    flex-shrink: 0;
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 4px;
    display: flex;
    align-items: center;
    transition: color 0.2s;
}

.banner-close:hover {
    color: var(--text-primary);
}

.error-banner {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    background: var(--status-critical-dim);
    border: 1px solid var(--status-critical);
    border-radius: var(--radius-md);
    color: var(--status-critical);
    font-size: 13px;
}

.validation-content {
    display: grid;
    grid-template-columns: 350px 1fr;
    gap: 20px;
    flex: 1;
    min-height: 0;
}

.rules-panel,
.details-panel {
    display: flex;
    flex-direction: column;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    overflow: hidden;
}

.panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px;
    border-bottom: 1px solid var(--border-subtle);
}

.panel-header h2 {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
}

.rule-count {
    padding: 4px 8px;
    background: var(--bg-overlay);
    border-radius: var(--radius-sm);
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
}

.rules-list {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.details-panel {
    padding: 20px;
    overflow-y: auto;
}

.validation-details {
    display: flex;
    flex-direction: column;
    gap: 24px;
}

.rule-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
}

.rule-header h3 {
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 4px 0;
}

.rule-description {
    font-size: 13px;
    color: var(--text-secondary);
    margin: 0;
}

.rule-meta {
    display: flex;
    gap: 8px;
}

.severity {
    padding: 4px 8px;
    border-radius: var(--radius-sm);
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.severity.critical {
    background: var(--status-critical-dim);
    color: var(--status-critical);
}

.severity.high {
    background: var(--status-warning-dim);
    color: var(--status-warning);
}

.severity.warning {
    background: var(--status-warning-dim);
    color: var(--status-warning);
}

.severity.info {
    background: var(--status-info-dim);
    color: var(--status-info);
}

.action-buttons {
    display: flex;
    gap: 12px;
    padding-top: 16px;
    border-top: 1px solid var(--border-subtle);
    flex-wrap: wrap;
}

.btn {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 16px;
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    background: var(--bg-surface);
    color: var(--text-secondary);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all var(--transition-fast);
}

.btn:hover:not(:disabled) {
    background: var(--bg-hover);
    color: var(--text-primary);
    border-color: var(--border-default);
}

.btn.btn-primary {
    background: var(--accent-primary);
    color: white;
    border-color: var(--accent-primary);
}

.btn.btn-primary:hover:not(:disabled) {
    background: var(--accent-primary-hover);
    border-color: var(--accent-primary-hover);
}

.btn.btn-warning {
    background: var(--status-warning-dim);
    color: var(--status-warning);
    border-color: var(--status-warning);
}

.btn.btn-warning:hover:not(:disabled) {
    background: var(--status-warning-dim);
    border-color: var(--status-warning);
    filter: brightness(0.95);
}

.btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    gap: 12px;
    text-align: center;
    color: var(--text-muted);
}

.empty-icon {
    color: var(--text-muted);
    opacity: 0.5;
}

.empty-state p {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
    margin: 0;
}

.empty-hint {
    font-size: 12px;
    color: var(--text-muted);
}

.loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    gap: 12px;
    color: var(--text-muted);
}

.spinner {
    width: 24px;
    height: 24px;
    border: 2px solid var(--border-subtle);
    border-top-color: var(--accent-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

@media (max-width: 1200px) {
    .validation-content {
        grid-template-columns: 1fr;
    }

    .rules-panel {
        max-height: 300px;
    }
}
</style>
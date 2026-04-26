<!-- Rule Validation Dashboard - Phase 5 -->
<script setup lang="ts">
import {
    ClipboardCheck, Activity, CheckCircle2, Clock, AlertTriangle, X, Edit, Zap
} from 'lucide-vue-next'

import TestingRuleCard from '../components/rules/TestingRuleCard.vue'
import RuleLifecycleTimeline from '../components/rules/RuleLifecycleTimeline.vue'
import ReadinessChecklist from '../components/rules/ReadinessChecklist.vue'
import { useRuleValidation } from '../composables/useRuleValidation'

const {
    testingRules,
    selectedRule,
    validationData,
    loading,
    promoting,
    error,
    newlyDeployedRule,
    isReady,
    stats,
    handleSelectRule,
    handleAdjustRule,
    handlePromote
} = useRuleValidation()
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

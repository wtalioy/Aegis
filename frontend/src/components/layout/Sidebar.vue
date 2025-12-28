<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  Target,
  Wand2,
  Search,
  Bot,
  Settings,
  Shield,
  CheckCircle2,
  ClipboardCheck
} from 'lucide-vue-next'

interface NavItem {
  icon: any
  label: string
  route: string
  section: 'core' | 'system'
  badge?: number
  badgeType?: 'default' | 'critical'
}

const route = useRoute()

const navItems: NavItem[] = [
  { icon: Target, label: 'Observatory', route: '/observatory', section: 'core' },
  { icon: Wand2, label: 'Policy Studio', route: '/policy-studio', section: 'core' },
  { icon: ClipboardCheck, label: 'Rule Validation', route: '/rule-validation', section: 'core' },
  { icon: Search, label: 'Investigation', route: '/investigation', section: 'core' },
  { icon: Bot, label: 'Sentinel', route: '/sentinel', section: 'core' },
  { icon: Settings, label: 'Settings', route: '/settings', section: 'system' },
]

const coreItems = computed(() => navItems.filter(item => item.section === 'core'))
const systemItems = computed(() => navItems.filter(item => item.section === 'system'))

const isActive = (path: string) => {
  // Handle Observatory - / redirects to /observatory
  if (path === '/observatory') {
    return route.path === '/observatory' || route.path === '/'
  }
  return route.path === path
}
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <Shield class="logo-icon" />
      <span class="logo-text">Aegis</span>
    </div>

    <nav class="sidebar-nav">
      <div class="nav-section">
        <span class="nav-section-label">CORE</span>
        <router-link v-for="item in coreItems" :key="item.route" :to="item.route" class="nav-item"
          :class="{ active: isActive(item.route) }">
          <component :is="item.icon" class="nav-icon" :size="20" />
          <span class="nav-label">{{ item.label }}</span>
          <span v-if="item.badge" class="nav-badge" :class="{ critical: item.badgeType === 'critical' }">
            {{ item.badge }}
          </span>
        </router-link>
      </div>

      <div class="nav-section">
        <span class="nav-section-label">SYSTEM</span>
        <router-link v-for="item in systemItems" :key="item.route" :to="item.route" class="nav-item"
          :class="{ active: isActive(item.route) }">
          <component :is="item.icon" class="nav-icon" :size="20" />
          <span class="nav-label">{{ item.label }}</span>
        </router-link>
      </div>
    </nav>
  </aside>
</template>

<style scoped>
.sidebar {
  width: var(--sidebar-width);
  background: var(--bg-surface);
  border-right: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  height: var(--topbar-height);
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.logo-icon {
  color: var(--accent-primary);
  flex-shrink: 0;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.sidebar-nav {
  flex: 1;
  padding: 16px 0;
  overflow-y: auto;
}

.nav-section {
  margin-bottom: 24px;
}

.nav-section-label {
  display: block;
  padding: 0 16px;
  margin-bottom: 8px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  margin: 5px 8px;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-decoration: none;
  transition: all var(--transition-fast);
}

.nav-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.nav-item.active {
  background: var(--accent-glow);
  color: var(--accent-primary);
}

.nav-icon {
  flex-shrink: 0;
}

.nav-label {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
}

.nav-badge {
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-size: 11px;
  font-weight: 600;
  background: var(--bg-overlay);
  color: var(--text-secondary);
}

.nav-badge.critical {
  background: var(--status-critical-dim);
  color: var(--status-critical);
}
</style>

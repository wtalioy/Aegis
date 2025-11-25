<script setup lang="ts">
import { computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent } from 'echarts/components'

use([CanvasRenderer, PieChart, TooltipComponent, LegendComponent])

const props = defineProps<{
  high: number
  warning: number
  info: number
}>()

const total = computed(() => props.high + props.warning + props.info)

const option = computed(() => ({
  tooltip: {
    trigger: 'item',
    backgroundColor: 'rgba(18, 18, 26, 0.95)',
    borderColor: 'rgba(255, 255, 255, 0.1)',
    textStyle: {
      color: '#f1f5f9'
    },
    formatter: '{b}: {c} ({d}%)'
  },
  series: [
    {
      type: 'pie',
      radius: ['55%', '80%'],
      center: ['50%', '50%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 6,
        borderColor: '#12121a',
        borderWidth: 3
      },
      label: {
        show: false
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 14,
          fontWeight: 'bold',
          color: '#f1f5f9'
        },
        itemStyle: {
          shadowBlur: 20,
          shadowColor: 'rgba(0, 0, 0, 0.5)'
        }
      },
      labelLine: {
        show: false
      },
      data: [
        { 
          value: props.high, 
          name: 'High', 
          itemStyle: { color: '#ef4444' }
        },
        { 
          value: props.warning, 
          name: 'Warning', 
          itemStyle: { color: '#f59e0b' }
        },
        { 
          value: props.info, 
          name: 'Info', 
          itemStyle: { color: '#3b82f6' }
        }
      ].filter(d => d.value > 0)
    }
  ],
  graphic: total.value > 0 ? [
    {
      type: 'text',
      left: 'center',
      top: '45%',
      style: {
        text: total.value.toString(),
        fontSize: 28,
        fontWeight: 'bold',
        fontFamily: 'JetBrains Mono, monospace',
        fill: '#f1f5f9'
      }
    },
    {
      type: 'text',
      left: 'center',
      top: '58%',
      style: {
        text: 'ALERTS',
        fontSize: 11,
        fill: '#64748b'
      }
    }
  ] : []
}))
</script>

<template>
  <div class="severity-pie">
    <v-chart v-if="total > 0" :option="option" autoresize />
    <div v-else class="no-data">
      <span class="no-data-value">0</span>
      <span class="no-data-label">No Alerts</span>
    </div>
  </div>
</template>

<style scoped>
.severity-pie {
  width: 100%;
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.no-data {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.no-data-value {
  font-size: 32px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--status-safe);
}

.no-data-label {
  font-size: 12px;
  color: var(--text-muted);
}
</style>


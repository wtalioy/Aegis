<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { subscribeToEventRates, type EventRates } from '../../lib/api'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent])

const MAX_POINTS = 60

const execData = ref<number[]>(Array(MAX_POINTS).fill(0))
const networkData = ref<number[]>(Array(MAX_POINTS).fill(0))
const fileData = ref<number[]>(Array(MAX_POINTS).fill(0))
const timeLabels = ref<string[]>(Array(MAX_POINTS).fill(''))

let unsubscribe: (() => void) | null = null

const option = computed(() => ({
  tooltip: {
    trigger: 'axis',
    backgroundColor: 'rgba(18, 18, 26, 0.95)',
    borderColor: 'rgba(255, 255, 255, 0.1)',
    textStyle: { color: '#f1f5f9' }
  },
  legend: {
    data: ['Exec', 'Network', 'File'],
    textStyle: { color: '#94a3b8' },
    top: 0,
    right: 0
  },
  grid: { left: 50, right: 20, top: 40, bottom: 30 },
  xAxis: {
    type: 'category',
    data: timeLabels.value,
    axisLine: { lineStyle: { color: 'rgba(255, 255, 255, 0.1)' } },
    axisLabel: { color: '#64748b', fontSize: 10 },
    axisTick: { show: false }
  },
  yAxis: {
    type: 'value',
    axisLine: { show: false },
    axisLabel: { color: '#64748b', fontSize: 10 },
    splitLine: { lineStyle: { color: 'rgba(255, 255, 255, 0.05)' } }
  },
  series: [
    {
      name: 'Exec',
      type: 'line',
      data: execData.value,
      smooth: true,
      symbol: 'none',
      lineStyle: { color: '#60a5fa', width: 2 },
      areaStyle: {
        color: {
          type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(96, 165, 250, 0.3)' },
            { offset: 1, color: 'rgba(96, 165, 250, 0)' }
          ]
        }
      }
    },
    {
      name: 'Network',
      type: 'line',
      data: networkData.value,
      smooth: true,
      symbol: 'none',
      lineStyle: { color: '#f59e0b', width: 2 },
      areaStyle: {
        color: {
          type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(245, 158, 11, 0.3)' },
            { offset: 1, color: 'rgba(245, 158, 11, 0)' }
          ]
        }
      }
    },
    {
      name: 'File',
      type: 'line',
      data: fileData.value,
      smooth: true,
      symbol: 'none',
      lineStyle: { color: '#10b981', width: 2 },
      areaStyle: {
        color: {
          type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(16, 185, 129, 0.3)' },
            { offset: 1, color: 'rgba(16, 185, 129, 0)' }
          ]
        }
      }
    }
  ]
}))

const updateData = (rates: EventRates) => {
  execData.value.shift()
  execData.value.push(rates.exec)
  networkData.value.shift()
  networkData.value.push(rates.network)
  fileData.value.shift()
  fileData.value.push(rates.file)

  timeLabels.value.shift()
  timeLabels.value.push(new Date().toLocaleTimeString('en-US', { 
    hour12: false, minute: '2-digit', second: '2-digit' 
  }))
}

onMounted(() => {
  unsubscribe = subscribeToEventRates(updateData)
})

onUnmounted(() => {
  unsubscribe?.()
})
</script>

<template>
  <div class="events-chart">
    <v-chart :option="option" autoresize />
  </div>
</template>

<style scoped>
.events-chart {
  width: 100%;
  height: 280px;
}
</style>

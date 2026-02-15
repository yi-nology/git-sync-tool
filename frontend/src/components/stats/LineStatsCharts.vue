<template>
  <div class="line-stats-charts" v-if="lineStatsData && lineStatsData.status === 'ready'">
    <el-row :gutter="16">
      <el-col :span="10">
        <el-card shadow="hover">
          <template #header><span class="chart-title">代码组成</span></template>
          <div ref="compositionChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card shadow="hover">
          <template #header><span class="chart-title">语言分布</span></template>
          <div ref="langChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useEcharts } from '@/composables/useEcharts'
import type { LineStatsResponse } from '@/types/stats'

const props = defineProps<{ lineStatsData: LineStatsResponse | null }>()

const compositionChartRef = ref<HTMLElement | null>(null)
const langChartRef = ref<HTMLElement | null>(null)

const compositionChart = useEcharts(compositionChartRef)
const langChart = useEcharts(langChartRef)

function updateCharts() {
  if (!props.lineStatsData || props.lineStatsData.status !== 'ready') return
  nextTick(() => {
    renderCompositionChart()
    renderLangChart()
  })
}

function renderCompositionChart() {
  const d = props.lineStatsData!
  const data = [
    { name: '代码行', value: d.code_lines, itemStyle: { color: '#409EFF' } },
    { name: '注释行', value: d.comment_lines, itemStyle: { color: '#67C23A' } },
    { name: '空白行', value: d.blank_lines, itemStyle: { color: '#C0C4CC' } },
  ].filter(item => item.value > 0)

  compositionChart.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} 行 ({d}%)',
    },
    legend: { bottom: 0 },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['50%', '45%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 8, borderColor: '#fff', borderWidth: 2 },
      label: {
        formatter: '{b}\n{d}%',
      },
      emphasis: {
        label: { show: true, fontSize: 16, fontWeight: 'bold' },
      },
      data,
    }],
  })
}

function renderLangChart() {
  const langs = [...(props.lineStatsData?.languages || [])]
    .sort((a, b) => (b.code + b.comment + b.blank) - (a.code + a.comment + a.blank))
    .slice(0, 10)
    .reverse()

  if (!langs.length) return

  const names = langs.map(l => l.name)

  langChart.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
    },
    legend: { bottom: 0 },
    grid: { left: 10, right: 40, top: 10, bottom: 40, containLabel: true },
    xAxis: { type: 'value' },
    yAxis: {
      type: 'category',
      data: names,
      axisLabel: { width: 80, overflow: 'truncate' },
    },
    series: [
      {
        name: '代码行',
        type: 'bar',
        stack: 'total',
        data: langs.map(l => l.code),
        itemStyle: { color: '#409EFF' },
        barMaxWidth: 24,
      },
      {
        name: '注释行',
        type: 'bar',
        stack: 'total',
        data: langs.map(l => l.comment),
        itemStyle: { color: '#67C23A' },
        barMaxWidth: 24,
      },
      {
        name: '空白行',
        type: 'bar',
        stack: 'total',
        data: langs.map(l => l.blank),
        itemStyle: { color: '#C0C4CC' },
        barMaxWidth: 24,
      },
    ],
  })
}

watch(() => props.lineStatsData, () => updateCharts(), { deep: true })
</script>

<style scoped>
.chart-container {
  width: 100%;
  height: 350px;
}
.chart-title {
  font-weight: 600;
  font-size: 14px;
}
</style>

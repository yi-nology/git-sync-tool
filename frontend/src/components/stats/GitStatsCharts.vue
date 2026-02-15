<template>
  <div class="git-stats-charts" v-if="statsData && statsData.authors?.length">
    <!-- Row 1: Contributor ranking + Time trend -->
    <el-row :gutter="16" class="mb-4">
      <el-col :span="10">
        <el-card shadow="hover">
          <template #header><span class="chart-title">贡献者排行</span></template>
          <div ref="authorChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card shadow="hover">
          <template #header><span class="chart-title">提交趋势</span></template>
          <div ref="trendChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Row 2: File type distribution -->
    <el-row class="mb-4">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header><span class="chart-title">文件类型分布</span></template>
          <div ref="fileTypeChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Row 3: Contributor detail list -->
    <el-card shadow="hover">
      <template #header>
        <div class="contributor-header">
          <span class="chart-title">贡献者详情</span>
          <el-tag size="small" type="info">共 {{ sortedAuthors.length }} 人</el-tag>
        </div>
      </template>
      <div class="contributor-list">
        <div v-for="(author, idx) in sortedAuthors" :key="author.email" class="contributor-item">
          <div class="contributor-rank">#{{ idx + 1 }}</div>
          <div class="contributor-avatar" :style="{ background: CHART_COLORS[idx % CHART_COLORS.length] }">
            {{ author.name.charAt(0).toUpperCase() }}
          </div>
          <div class="contributor-info">
            <div class="contributor-name-row">
              <span class="contributor-name">{{ author.name }}</span>
              <el-text type="info" size="small">{{ author.email }}</el-text>
            </div>
            <div class="contributor-bar-row">
              <el-progress
                :percentage="maxLines > 0 ? Math.round(author.total_lines / maxLines * 100) : 0"
                :stroke-width="14"
                :color="CHART_COLORS[idx % CHART_COLORS.length]"
                :show-text="false"
              />
            </div>
            <div class="contributor-meta-row">
              <span class="contributor-lines">{{ author.total_lines.toLocaleString() }} 行</span>
              <span class="contributor-percent">{{ totalLines > 0 ? ((author.total_lines / totalLines) * 100).toFixed(1) : 0 }}%</span>
              <span class="contributor-tags">
                <el-tag
                  v-for="ft in getTopFileTypes(author, 3)"
                  :key="ft.name"
                  size="small"
                  type="info"
                  effect="plain"
                  round
                >{{ ft.name }} {{ ft.lines }}</el-tag>
              </span>
            </div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useEcharts, CHART_COLORS } from '@/composables/useEcharts'
import type { StatsResponse, AuthorStat } from '@/types/stats'

const props = defineProps<{ statsData: StatsResponse | null }>()

const authorChartRef = ref<HTMLElement | null>(null)
const trendChartRef = ref<HTMLElement | null>(null)
const fileTypeChartRef = ref<HTMLElement | null>(null)

const authorChart = useEcharts(authorChartRef)
const trendChart = useEcharts(trendChartRef)
const fileTypeChart = useEcharts(fileTypeChartRef)

const sortedAuthors = computed(() =>
  [...(props.statsData?.authors || [])].sort((a, b) => b.total_lines - a.total_lines)
)

const maxLines = computed(() => sortedAuthors.value[0]?.total_lines || 0)
const totalLines = computed(() => props.statsData?.total_lines || 0)

function getTopFileTypes(author: AuthorStat, count: number) {
  if (!author.file_types) return []
  return Object.entries(author.file_types)
    .sort((a, b) => b[1] - a[1])
    .slice(0, count)
    .map(([name, lines]) => ({ name, lines }))
}

function updateCharts() {
  if (!props.statsData?.authors?.length) return
  nextTick(() => {
    renderAuthorChart()
    renderTrendChart()
    renderFileTypeChart()
  })
}

function renderAuthorChart() {
  const authors = sortedAuthors.value.slice(0, 10).reverse()

  authorChart.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: unknown) => {
        const p = (params as Array<{ name: string; value: number; marker: string }>)[0]
        if (!p) return ''
        const author = props.statsData?.authors?.find(a => a.name === p.name)
        const pct = totalLines.value > 0 ? ((p.value / totalLines.value) * 100).toFixed(1) : '0'
        let html = `<b>${p.name}</b>`
        if (author?.email) html += `<br/><span style="color:#909399">${author.email}</span>`
        html += `<br/>${p.marker} ${p.value.toLocaleString()} 行 (${pct}%)`
        return html
      },
    },
    grid: { left: 10, right: 60, top: 10, bottom: 10, containLabel: true },
    xAxis: { type: 'value' },
    yAxis: {
      type: 'category',
      data: authors.map(a => a.name),
      axisLabel: { width: 80, overflow: 'truncate' },
    },
    series: [{
      type: 'bar',
      data: authors.map((a, i) => ({
        value: a.total_lines,
        itemStyle: {
          color: CHART_COLORS[(9 - i) % CHART_COLORS.length],
          borderRadius: [0, 4, 4, 0],
        },
      })),
      label: { show: true, position: 'right', formatter: '{c}' },
      barMaxWidth: 28,
    }],
  })
}

function renderTrendChart() {
  const authors = props.statsData?.authors || []
  const timeSet = new Set<string>()
  authors.forEach(a => {
    if (a.time_trend) Object.keys(a.time_trend).forEach(k => timeSet.add(k))
  })
  const times = [...timeSet].sort()
  if (!times.length) return

  const topAuthors = sortedAuthors.value.slice(0, 5)

  const totalData = times.map(t => {
    let sum = 0
    authors.forEach(a => { if (a.time_trend?.[t]) sum += a.time_trend[t] })
    return sum
  })

  const series: Record<string, unknown>[] = [{
    name: '总计',
    type: 'line',
    data: totalData,
    smooth: true,
    lineStyle: { width: 3 },
    areaStyle: { opacity: 0.15 },
    itemStyle: { color: '#409EFF' },
  }]

  topAuthors.forEach((a, i) => {
    series.push({
      name: a.name,
      type: 'line',
      data: times.map(t => a.time_trend?.[t] || 0),
      smooth: true,
      lineStyle: { width: 2, type: 'dashed' as const },
      itemStyle: { color: CHART_COLORS[(i + 1) % CHART_COLORS.length] },
    })
  })

  trendChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: {
      data: ['总计', ...topAuthors.map(a => a.name)],
      bottom: 0,
      type: 'scroll',
    },
    grid: { left: 10, right: 20, top: 10, bottom: 40, containLabel: true },
    xAxis: { type: 'category', data: times, boundaryGap: false },
    yAxis: { type: 'value' },
    series,
  })
}

function renderFileTypeChart() {
  const authors = props.statsData?.authors || []
  const typeMap: Record<string, number> = {}
  authors.forEach(a => {
    if (a.file_types) {
      Object.entries(a.file_types).forEach(([k, v]) => {
        typeMap[k] = (typeMap[k] || 0) + v
      })
    }
  })

  const sorted = Object.entries(typeMap).sort((a, b) => b[1] - a[1])
  const top = sorted.slice(0, 10)
  const othersVal = sorted.slice(10).reduce((s, [, v]) => s + v, 0)
  const data = top.map(([name, value]) => ({ name, value }))
  if (othersVal > 0) data.push({ name: '其他', value: othersVal })

  fileTypeChart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} 行 ({d}%)' },
    legend: { orient: 'vertical', right: 20, top: 'center', type: 'scroll' },
    color: CHART_COLORS,
    series: [{
      type: 'pie',
      radius: ['35%', '65%'],
      center: ['40%', '50%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { formatter: '{b}\n{d}%' },
      data,
    }],
  })
}

watch(() => props.statsData, () => updateCharts(), { deep: true })
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
.mb-4 {
  margin-bottom: 16px;
}

/* Contributor header */
.contributor-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* Contributor list */
.contributor-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.contributor-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  transition: background 0.2s;
}
.contributor-item:hover {
  background: #f5f7fa;
}
.contributor-rank {
  width: 28px;
  font-size: 13px;
  font-weight: 700;
  color: #909399;
  text-align: center;
  flex-shrink: 0;
}
.contributor-item:nth-child(1) .contributor-rank { color: #E6A23C; }
.contributor-item:nth-child(2) .contributor-rank { color: #909399; }
.contributor-item:nth-child(3) .contributor-rank { color: #B87333; }

.contributor-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 700;
  font-size: 16px;
  flex-shrink: 0;
}
.contributor-info {
  flex: 1;
  min-width: 0;
}
.contributor-name-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 4px;
}
.contributor-name {
  font-weight: 600;
  font-size: 14px;
  color: #303133;
}
.contributor-bar-row {
  margin-bottom: 4px;
}
.contributor-meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.contributor-lines {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}
.contributor-percent {
  font-size: 12px;
  color: #909399;
}
.contributor-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
.contributor-tags .el-tag {
  font-size: 11px;
}
</style>

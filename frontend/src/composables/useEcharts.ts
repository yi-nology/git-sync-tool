import { shallowRef, onMounted, onBeforeUnmount, type Ref } from 'vue'
import * as echarts from 'echarts/core'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { EChartsOption } from 'echarts'

echarts.use([
  BarChart, LineChart, PieChart,
  TitleComponent, TooltipComponent, LegendComponent, GridComponent,
  CanvasRenderer,
])

export const CHART_COLORS = [
  '#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399',
  '#9B59B6', '#1ABC9C', '#3498DB', '#E74C3C', '#2ECC71',
]

export function useEcharts(domRef: Ref<HTMLElement | null>) {
  const instance = shallowRef<echarts.ECharts | null>(null)

  function init() {
    if (domRef.value && !instance.value) {
      instance.value = echarts.init(domRef.value)
    }
  }

  function setOption(option: EChartsOption) {
    if (!instance.value) init()
    instance.value?.setOption(option, true)
  }

  function handleResize() {
    instance.value?.resize()
  }

  onMounted(() => {
    init()
    window.addEventListener('resize', handleResize)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('resize', handleResize)
    instance.value?.dispose()
    instance.value = null
  })

  return { instance, setOption, resize: handleResize }
}

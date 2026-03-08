import { ref, computed, watch, type Ref, type ComputedRef } from 'vue'

export interface PaginationOptions {
  pageSize?: number
  pageSizes?: number[]
}

/**
 * 通用分页 composable
 */
export function usePagination<T>(
  dataSource: Ref<T[]> | ComputedRef<T[]>,
  options: PaginationOptions = {}
) {
  const {
    pageSize = 10,
    pageSizes = [10, 20, 50, 100]
  } = options

  const currentPage = ref(1)
  const currentPageSize = ref(pageSize)
  const availablePageSizes = pageSizes

  // 计算总条数
  const total = computed(() => dataSource.value.length)

  // 计算总页数
  const totalPages = computed(() =>
    Math.ceil(total.value / currentPageSize.value)
  )

  // 分页后的数据
  const paginatedData = computed(() => {
    const start = (currentPage.value - 1) * currentPageSize.value
    const end = start + currentPageSize.value
    return dataSource.value.slice(start, end)
  })

  // 页码变化
  const handlePageChange = (page: number) => {
    currentPage.value = page
  }

  // 每页条数变化
  const handleSizeChange = (size: number) => {
    currentPageSize.value = size
    currentPage.value = 1 // 重置到第一页
  }

  // 重置分页
  const resetPagination = () => {
    currentPage.value = 1
    currentPageSize.value = pageSize
  }

  // 当数据源变化时，如果当前页超出范围，重置到最后一页
  watch(total, (newTotal) => {
    const maxPage = Math.ceil(newTotal / currentPageSize.value)
    if (currentPage.value > maxPage && maxPage > 0) {
      currentPage.value = maxPage
    }
  })

  return {
    currentPage,
    currentPageSize,
    availablePageSizes,
    total,
    totalPages,
    paginatedData,
    handlePageChange,
    handleSizeChange,
    resetPagination,
  }
}

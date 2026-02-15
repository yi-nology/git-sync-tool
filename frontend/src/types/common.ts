export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
}

export interface PaginationParams {
  page?: number
  page_size?: number
  search?: string
}

export interface PaginationResponse<T> {
  total: number
  list: T[]
}

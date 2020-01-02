export interface Time {
  create_time: string
  update_time: string
}

export interface SashResponse<T> {
  page_num: number
  page_size: number
  total: number
  data: T
}

export interface DependencyBasic {
  service_name: string
  dependencies: string[]
}

export interface Dependency extends DependencyBasic, Time {
}

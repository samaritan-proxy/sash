import { SashResponse, Dependency, DependencyBasic } from "../models/models";
import { ajaxGet, ajaxPut, ajaxPost, ajaxDelete } from './ajax'

const APIPrefix = ""

// note page start from 0
export function GetDependencies(page: number, service: string): Promise<SashResponse<Dependency[]>> {
  if (service === "") {
    return ajaxGet(`${APIPrefix}/dependencies?page_num=${page}`)
  } 
  return ajaxGet(`${APIPrefix}/dependencies?service_name=${escape(service)}`)
}

export function PutDependency(dependency: DependencyBasic) {
  return ajaxPut(`${APIPrefix}/dependencies/${escape(dependency.service_name)}`, { dependencies: dependency.dependencies })
}

export function PostDependency(dependency: DependencyBasic) {
  return ajaxPost(`${APIPrefix}/dependencies`, dependency)
}
export function DeleteDependency(service: string) {
  return ajaxDelete(`${APIPrefix}/dependencies/${escape(service)}`)
}
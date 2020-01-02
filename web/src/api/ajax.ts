import axios, { AxiosError, AxiosResponse } from 'axios'
import { notification, message } from 'antd'

/**
 * ajax get请求
 * @param url         服务地址
 * @param fn          回调函数，无传入undefined
 * @param config      get请求参数体
 * @param isCatch     是否捕获异常
 * @param isReturnData  是否返回后端返回数据
 */
async function ajaxGet(
  url: string,
  fn?: (data: any) => void,
  config?: object,
  isCatch: boolean = true,
  isReturnData: boolean = true,
  isFilterNull: boolean = false
): Promise<any> {
  let res: void | AxiosResponse
  if (!isCatch) {
    res = await axios.get(`${url}`, config)
  } else {
    res = await axios.get(`${url}`, config).catch(errorHandler)
  }
  if (res) {
    if (!validateResStatus(res)) {
      return void 0
    }
    if (!isReturnData) {
      return true
    }
    if (fn) {
      return fn(res.data)
    }
    return isFilterNull ? res.data.fliter((d: any) => d) : res.data
  }
  return res
}


async function ajaxPost(
  url: string,
  data: any,
  returnData: boolean = false,
  config?: object
): Promise<any> {
  let doCall = async () => {
    const res = await axios.post(`${url}`, data, config).catch(errorHandler)
    if (res) {
      if (!validateResStatus(res)) {
        return false
      }
      return returnData ? res.data : true
    }
    return false
  }
  doCallWithMessage(doCall)
}


async function ajaxPut(
  url: string,
  data: any,
  config?: object,
  returnData: boolean = false
): Promise<any> {
  let doCall = async () => {
    const res = await axios.put(`${url}`, data, config).catch(errorHandler)
    if (res) {
      if (!validateResStatus(res)) {
        return false
      }
      return returnData ? res.data : true
    }
    return false
  }
  doCallWithMessage(doCall)
}
/* 
 * 删除数据body体放入到config参数中，如：
 * const data = { id: id }
 * const res = await ajaxDelete(url, { data: data }) 
 */
async function ajaxDelete(
  url: string,
  config?: object,
  returnData: boolean = false
): Promise<any> {
  let doCall = async () => {
    const res = await axios.delete(`${url}`, config).catch(errorHandler)
    if (res) {
      if (!validateResStatus(res)) {
        return false
      }
      return returnData ? res.data : true
    }
    return false
  }
  doCallWithMessage(doCall)
}

async function ajaxPatch(
  url: string,
  data: any,
  config?: object,
  returnData: boolean = false
): Promise<any> {
  let doCall = async () => {
    const res = await axios.patch(`${url}`, data, config).catch(errorHandler)
    if (res) {
      if (!validateResStatus(res)) {
        return false
      }
      return returnData ? res.data : true
    }
    return false
  }
  doCallWithMessage(doCall)
}

function validateResStatus(res: AxiosResponse) {
  if (res.status !== 200 && res.status !== 201 && res.status !== 202) {
    notification.error({
      message: 'Error',
      description: res.data
    })
    return false
  }
  return true
}
  
function errorHandler(err: AxiosError) {
  notification.error({
    message: 'Error',
    description: err.message
  })
}

// patchWithNotice same as ajaxPatch but will prompt notice when done
function doCallWithMessage(doCall: any) {
  message.loading("请求发送中")
  doCall().then(
    (res: any) => {
      message.destroy()
      if (res) {
        message.success("请求成功")
      } else {
        message.error("请求失败")
      }
      return res
    }
  )
}

export { ajaxGet, ajaxPost, ajaxPut, ajaxDelete, ajaxPatch }
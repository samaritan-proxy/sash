// Copyright 2019 Samaritan Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import axios, {AxiosError, AxiosResponse} from 'axios'
import {message, notification} from 'antd'

async function ajaxGet(
    url: string,
    fn?: (data: any) => void,
    config?: object,
    isCatch: boolean = true,
    isReturnData: boolean = true,
    isFilterNull: boolean = false
): Promise<any> {
    let res: void | AxiosResponse;
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
        const res = await axios.post(`${url}`, data, config).catch(errorHandler);
        if (res) {
            if (!validateResStatus(res)) {
                return false
            }
            return returnData ? res.data : true
        }
        return false
    };
    doCallWithMessage(doCall)
}

async function ajaxPut(
    url: string,
    data: any,
    config?: object,
    returnData: boolean = false
): Promise<any> {
    let doCall = async () => {
        const res = await axios.put(`${url}`, data, config).catch(errorHandler);
        if (res) {
            if (!validateResStatus(res)) {
                return false
            }
            return returnData ? res.data : true
        }
        return false
    };
    doCallWithMessage(doCall)
}

async function ajaxDelete(
    url: string,
    config?: object,
    returnData: boolean = false
): Promise<any> {
    let doCall = async () => {
        const res = await axios.delete(`${url}`, config).catch(errorHandler);
        if (res) {
            if (!validateResStatus(res)) {
                return false
            }
            return returnData ? res.data : true
        }
        return false
    };
    doCallWithMessage(doCall)
}

async function ajaxPatch(
    url: string,
    data: any,
    config?: object,
    returnData: boolean = false
): Promise<any> {
    let doCall = async () => {
        const res = await axios.patch(`${url}`, data, config).catch(errorHandler);
        if (res) {
            if (!validateResStatus(res)) {
                return false
            }
            return returnData ? res.data : true
        }
        return false
    };
    doCallWithMessage(doCall)
}

function validateResStatus(res: AxiosResponse) {
    if (res.status !== 200 && res.status !== 201 && res.status !== 202) {
        notification.error({
            message: 'Error',
            description: res.data
        });
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
    message.loading("loading...");
    doCall().then(
        (res: any) => {
            message.destroy();
            if (res) {
                message.success("success")
            } else {
                message.error("failed")
            }
            return res
        }
    )
}

export {ajaxGet, ajaxPost, ajaxPut, ajaxDelete, ajaxPatch}
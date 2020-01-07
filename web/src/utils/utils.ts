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

import * as R from 'ramda';

export function validateJSONFormat(rule: any, value: any, callback: any, source?: any, options?: any): any {
    try {
        JSON.parse(value)
    } catch (e) {
        callback(e)
    }
    callback()
}

export interface Entry {
    Key: any
    Value: any
}

export function Object2Array(obj: object, keyMap?: Map<string, string>): Entry[] {
    let res: Entry[] = [];
    R.forEachObjIndexed((v, k) => {
        res = R.append({
            Key: keyMap && keyMap.has(k) ? keyMap.get(k) : k,
            Value: v
        } as Entry, res)
    }, obj);
    return res
}

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

import {LbPolicy, ProxyConfig} from "../../models/proxy-config";
import React from "react";
import {Divider, Form, Input, InputNumber, Select} from "antd";
import {FormComponentProps} from "antd/es/form";
import {keys, map} from "ramda";

interface ProxyConfigProps extends FormComponentProps {
    proxyConfig?: ProxyConfig
}

interface ProxyConfigState {

}

class ProxyConfigDetail extends React.Component<ProxyConfigProps, ProxyConfigState> {
    render() {
        const formItemLayout = {
            labelCol: {span: 4},
            wrapperCol: {span: 8},
        };
        const options = map(v => (
            <Select.Option key={v}> {v}</Select.Option>
        ), keys(LbPolicy));
        return (

            <Form layout="horizontal" style={{textAlign: 'left'}}>
                <Divider orientation="left">Listener</Divider>
                <Form.Item label="Address" {...formItemLayout}>
                    <Input
                        allowClear
                    />
                </Form.Item>
                <Form.Item label="Connection Limit" {...formItemLayout}>
                    <InputNumber/>
                </Form.Item>
                <Divider orientation="left">Load Balance</Divider>
                <Form.Item label="Policy" {...formItemLayout}>
                    <Select
                        defaultValue={LbPolicy.LEAST_CONNECTION}
                        dropdownRender={menu => {
                            return options
                        }}
                    />
                </Form.Item>
            </Form>

        )
    }
}

export default Form.create<ProxyConfigProps>({})(ProxyConfigDetail);
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

import React, {FormEvent, Fragment} from "react";
import {FormComponentProps} from "antd/es/form";
import {Button, Form, Input} from "antd";
import TextArea from "antd/es/input/TextArea";
import {PostProxyConfig} from "../../api/api";
import {validateJSONFormat} from "../../utils/utils";

interface CreateProxyConfigState {
}

interface CreateProxyConfigProps extends FormComponentProps {
}

class CreateProxyConfig extends React.Component<CreateProxyConfigProps, CreateProxyConfigState> {

    constructor(props: CreateProxyConfigProps) {
        super(props);
        this.handleSubmit = this.handleSubmit.bind(this)
    }

    handleSubmit(e: FormEvent<HTMLFormElement>) {
        e.preventDefault();
        this.props.form.validateFields((err: boolean, fieldsValue: any) => {
            if (err) {
                return;
            }
            PostProxyConfig({
                service_name: fieldsValue.service_name,
                config: JSON.parse(fieldsValue.config)
            })
        })
    }


    render() {
        const {form} = this.props;
        const {getFieldDecorator} = form;
        return (
            <Fragment>
                <Form onSubmit={this.handleSubmit} labelCol={{span: 6}} wrapperCol={{span: 12}}>
                    <Form.Item required={true} label="Service:">
                        {
                            getFieldDecorator("service_name", {
                                rules: [
                                    {required: true, message: "should not be empty"}
                                ]
                            })(<Input/>)
                        }
                    </Form.Item>
                    <Form.Item required={true} label="Proxy Config:">
                        {getFieldDecorator('config', {
                            rules: [
                                {required: true, message: "should not be empty"},
                                {validator: validateJSONFormat}
                            ]
                        })(<TextArea autoSize={{minRows: 10, maxRows: 15}}/>)}
                    </Form.Item>
                    <Form.Item wrapperCol={{offset: 6}} label="">
                        <Button type="primary" htmlType="submit">Submit</Button>
                    </Form.Item>
                </Form>
            </Fragment>
        )
    }
}

const WrappedCreateProxyConfig = Form.create({name: "create_proxy_config"})(CreateProxyConfig);
export default WrappedCreateProxyConfig
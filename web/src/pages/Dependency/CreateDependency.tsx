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

import React, {Fragment} from 'react'
import {Button, Form, Input, Select} from 'antd'
import {PostDependency} from '../../api/api'
import {FormComponentProps} from "antd/es/form";

interface CreateDependencyState {
}

interface CreateDependencyProps extends FormComponentProps {
}

class CreateDependency extends React.Component<CreateDependencyProps, CreateDependencyState> {

    constructor(props: any) {
        super(props);
        this.handleSubmit = this.handleSubmit.bind(this)
    }

    handleSubmit(e: any) {
        e.preventDefault();
        this.props.form.validateFields((err: boolean, fieldsValue: any) => {
            console.log('Received values of form: ', fieldsValue);
            if (err) {
                return;
            }
            let res = PostDependency(fieldsValue);
            console.warn(res)
        })
    }

    validateNotStartWithNumber(rule: any, value: string, callback: any) {
        if (value === "") {
            callback("should not be empty");
            return
        }
        let startWithNumber = value.match("^[0-9]");
        if (startWithNumber) {
            callback("should not start with number")
        } else {
            callback()
        }
    }

    render() {
        const {form} = this.props;
        const {getFieldDecorator} = form;
        return (
            <Fragment>
                <Form onSubmit={this.handleSubmit} labelCol={{span: 6}} wrapperCol={{span: 12}}>
                    <Form.Item required={true} label="Service:">
                        {getFieldDecorator('service_name', {
                            rules: [
                                {
                                    required: true,
                                    message: 'should not be empty',
                                },
                                {
                                    validator: this.validateNotStartWithNumber,
                                }
                            ]
                        })(<Input/>)}
                    </Form.Item>
                    <Form.Item required={true} label="Dependencies:">
                        {getFieldDecorator('dependencies', {})(<Select mode="tags"/>)}
                    </Form.Item>
                    <Form.Item wrapperCol={{offset: 6}} label="">
                        <Button type="primary" htmlType="submit">Submit</Button>
                    </Form.Item>
                </Form>
            </Fragment>
        )
    }
}

const WrappedForm = Form.create({name: 'create_dependency'})(CreateDependency);
export default WrappedForm

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
import {GetDependency, PutDependency} from '../../api/api'
import {RouteComponentProps} from "react-router-dom";
import {FormComponentProps} from "antd/es/form";
import {Dependency} from "../../models/dependency";
import * as queryString from "querystring";
import {slice, startsWith} from "ramda";
import {mockDependencies} from "../../dev/mocks";

interface UpdateDependencyState {
    dependency: Dependency
}

interface UpdateDependencyProps extends RouteComponentProps, FormComponentProps {

}

class UpdateDependency extends React.Component<UpdateDependencyProps, UpdateDependencyState> {
    state = {
        dependency: mockDependencies[0],
    };

    constructor(props: any) {
        super(props);
        this.handleSubmit = this.handleSubmit.bind(this);
        const {search} = this.props.location;
        const parsed = queryString.parse(startsWith('?', search) ? slice(1, Infinity, search) : search);
        const service = parsed['service'];
        if (!service) {
            this.props.history.push("/error/404");
            return
        }
        GetDependency(service.toString()).then(res => {
            if (!res) {
                this.props.history.push("/error/404");
                return
            }
            this.setState({dependency: res})
        });
    }

    componentDidMount(): void {

    }

    handleSubmit(e: any) {
        e.preventDefault();
        this.props.form.validateFields((err: boolean, fieldsValue: any) => {
            console.log('Received values of form: ', fieldsValue);
            if (err) {
                return;
            }
            let res = PutDependency(fieldsValue);
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
        const {dependency} = this.state;
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
                            ],
                            initialValue: dependency.service_name,
                        })(<Input disabled={true}/>)}
                    </Form.Item>
                    <Form.Item required={false} label="Dependencies:">
                        {getFieldDecorator('dependencies', {initialValue: dependency.dependencies})(<Select
                            mode="tags"/>)}
                    </Form.Item>
                    <Form.Item wrapperCol={{offset: 6}} label="">
                        <Button type="primary" htmlType="submit">Submit</Button>
                    </Form.Item>
                </Form>
            </Fragment>
        )
    }
}

const WrappedForm = Form.create({
    name: 'create_dependency'
})(UpdateDependency);
export default WrappedForm

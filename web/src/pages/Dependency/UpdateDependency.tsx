import React, { Fragment } from 'react'
import { Form, Input, Button, Select } from 'antd'
import { PutDependency } from '../../api/api'
import { DependencyBasic } from '../../models/models'

interface UpdateDependencyState {
}

interface UpdateDependencyProps {
  value: DependencyBasic
  form: any
}

class UpdateDependency extends React.Component<UpdateDependencyProps, UpdateDependencyState> {
  
  constructor(props: any) {
    super(props)
    this.handleSubmit = this.handleSubmit.bind(this)
  }

  handleSubmit(e: any) {
    e.preventDefault()
    this.props.form.validateFields((err: boolean, fieldsValue: any) => {
      console.log('Received values of form: ', fieldsValue);
      if (err) {
        return;
      }
      let res = PutDependency(fieldsValue)
      console.warn(res)
    })
  }

  validateNotStartWithNumber(rule: any, value: string, callback: any) {
    if (value === "") {
      callback("should not be empty")
      return
    }
    let startWithNumber = value.match("^[0-9]")
    if (startWithNumber) {
      callback("should not start with number")
    } else {
      callback()
    }
  }
      
  render() {
    const { form } = this.props
    const { getFieldDecorator } = form
    return (
      <Fragment>
        <Form onSubmit={this.handleSubmit} labelCol={{ span: 6 }} wrapperCol={{ span: 12 }}>
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
            })(<Input disabled={true} />)}
          </Form.Item> 
          <Form.Item required={false} label="Dependencies:">
            {getFieldDecorator('dependencies', {
            })(<Select mode="tags"/>)}
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
  name: 'create_dependency',
  mapPropsToFields: (props: UpdateDependencyProps) => {
    const value = props.value
    return {
      ["service_name"]: Form.createFormField({ value: value.service_name }),
      ["dependencies"]: Form.createFormField({ value: value.dependencies }),
    }
  }
})(UpdateDependency);
export default WrappedForm

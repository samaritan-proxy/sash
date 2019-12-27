import React, { Fragment } from 'react'
import { Form, Input, Button, Select } from 'antd'
import { PostDependency } from '../../api/api'

interface CreateDependencyState {
}

interface CreateDependencyProps {
  form: any
}

class CreateDependency extends React.Component<CreateDependencyProps, CreateDependencyState> {
  
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
      let res = PostDependency(fieldsValue)
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
            })(<Input />)}
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

const WrappedForm = Form.create({ name: 'create_dependency' })(CreateDependency);
export default WrappedForm

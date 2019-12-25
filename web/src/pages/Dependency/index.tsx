import React, { Fragment } from 'react'
import { Form, Select, Input, Button, Row, Col, Table, Icon } from 'antd'

interface DependencyState {
}

interface DependencyProps {
    
}

class Dependency extends React.Component<DependencyProps, DependencyState> {
  columns = [
    {
      title: 'Service Name',
      key: 'service_name',
      dataIndex: 'service_name',
    },
    {
      title: 'Create Time',
      key: 'create_time',
      dataIndex: 'create_time',
    },
    {
      title: 'Update Time',
      key: 'update_time',
      dataIndex: 'update_time',
    },
    {
      title: 'Dependency',
      key: 'dependency',
      dataIndex: 'dependency',
    },
    {
      title: 'Operation',
      render: (record: any) => (<Button key={record.id}>Update</Button>),
    }
  ]

  componentWillMount() {

  }

  render() {
    return (
      <Fragment>
        <Row>
          <Col span={12}>
            <Form labelCol={{ span: 6 }} wrapperCol={{ span: 10 }}>
              <Form.Item  label="Service:">
                <Select>
                  {/* add options */}
                </Select>
              </Form.Item>
              <Form.Item label="Instance ID:">
                <Input style={{ width: '65%', marginRight: '3%' }} />
                <Button style={{ width: '32%' }} type="primary">Search</Button>
              </Form.Item>
            </Form>
          </Col>
          <Col span={6}/>
          <Col span={6}>
            <Button style={{ alignSelf: "right" }} type="primary">
              <Icon type="plus" />
              Create
            </Button>
          </Col>
        </Row>
        <Table columns={this.columns} />
      </Fragment>
    )
  }
}
export default Dependency

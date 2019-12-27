import React, { Fragment } from 'react'
import { Form, Input, Button, Row, Col, Table, Icon, Modal } from 'antd'
import { Dependency } from '../../models/models'
import { mockDepedencies } from "../../dev/mocks";
import { RenderStringArrayAsTags } from "../../renders/renders";
import { GetDependencies, DeleteDependency } from '../../api/api'
import CreateDependency from './CreateDependency'
import UpdateDependency from './UpdateDependency'

interface DependencyPageState {
  loading: boolean
  page: number
  pageSize: number
  total: number
  searchService: string
  selectedDependency: Dependency
  dependencies: Dependency[]
  showUpdateModal: boolean
  showCreateModal: boolean
}

interface DependencyPageProps {
    
}

class DependencyPage extends React.Component<DependencyPageProps, DependencyPageState> {
  state = {
    page: 0,
    pageSize: 10,
    loading: true,
    total: 0,
    searchService: "",
    selectedDependency: {} as Dependency,
    dependencies: mockDepedencies,
    showUpdateModal: false,
    showCreateModal: false,
  }
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
      title: 'Dependencies',
      key: 'dependencies',
      dataIndex: 'dependencies',
      render: RenderStringArrayAsTags,
    },
    {
      title: 'Operation',
      render: (record: any) => (
        <div>
          <Button type="primary" onClick={this.onClickUpdateButton(record)} key={`${record.id}+update`}>Update</Button>
          <Button type="danger" onClick={this.onClickDeleteButton(record.service_name)} key={`${record.id}+delete`}>Delete</Button>
        </div>
      ),
    }
  ]

  constructor(props: any) {
    super(props)
    this.bindCallbacks()
  }

  bindCallbacks() {
    this.onClickUpdateButton = this.onClickUpdateButton.bind(this)
    this.onClickCreateButton = this.onClickCreateButton.bind(this)
    this.onCloseUpdateModal = this.onCloseUpdateModal.bind(this)
    this.onCloseCreateModal = this.onCloseCreateModal.bind(this)
    this.onInputServiceSearch = this.onInputServiceSearch.bind(this)
    this.onSearchService = this.onSearchService.bind(this)
    this.onClickDeleteButton = this.onClickDeleteButton.bind(this)
    this.onChangePage = this.onChangePage.bind(this)
  }

  onClickUpdateButton(record: Dependency) {
    return () => {
      this.setState({showUpdateModal: true, selectedDependency: record})
    }
  }

  onClickDeleteButton(service_name: string) {
    return async () => {
      await DeleteDependency(service_name)
      this.reloadPage()
    }
  }

  onClickCreateButton() {
    this.setState({showCreateModal: true})
  }

  onCloseUpdateModal() {
    this.setState({ showUpdateModal: false })
    this.reloadPage()
  }

  onCloseCreateModal() {
    this.setState({ showCreateModal: false })
    this.reloadPage()
  }

  onInputServiceSearch(e: any) {
    this.setState({ searchService: e.target.value })
  }

  onSearchService() {
    this.reloadPage()
  }

  onChangePage(pageNum: number) {
    this.setState({page: pageNum-1}, this.reloadPage)
  }

  componentDidMount() {
    this.reloadPage()
  }

  reloadPage() {
    this.setState({ loading: true}, this.getDependencies)
  }

  async getDependencies() {
    let res = await GetDependencies(this.state.page, this.state.searchService)
    if (res) {
      this.setState({
        page: res.page_num,
        pageSize: res.page_size,
        total: res.total,
        dependencies: res.data,
        loading: false
      })
    }
  }

  render() {
    const {page, pageSize, total, loading, dependencies, showCreateModal, showUpdateModal} = this.state
    return (
      <Fragment>
        <Row>
          <Col span={12}>
            <Form labelCol={{ span: 6 }} wrapperCol={{ span: 10 }}>
              <Form.Item label="Service:">
                <Input onChange={this.onInputServiceSearch} style={{ width: '65%', marginRight: '3%' }} />
                <Button onClick={this.onSearchService} style={{ width: '32%' }} type="primary">Search</Button>
              </Form.Item>
            </Form>
          </Col>
          <Col span={6}/>
          <Col span={6}>
            <Button onClick={this.onClickCreateButton} style={{ alignSelf: "right" }} type="primary">
              <Icon type="plus" />
              Create
            </Button>
          </Col>
        </Row>
        <Table
          rowKey="service_name"
          columns={this.columns}
          dataSource={dependencies}
          loading={loading}
          pagination={{
            pageSize: pageSize,
            current: page+1,
            total: total,
            onChange: this.onChangePage,
          }}
        />
        <Modal title="Create New Dependency" footer={null} visible={showCreateModal} onCancel={this.onCloseCreateModal}>
          <CreateDependency />
        </Modal>
        <Modal footer={null} visible={showUpdateModal} onCancel={this.onCloseUpdateModal}>
          <UpdateDependency value={this.state.selectedDependency}/>
        </Modal>
      </Fragment>
    )
  }
}
export default DependencyPage

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
import {Button, Col, Divider, Icon, Modal, Row, Table} from 'antd'
import {Dependency} from '../../models/dependency'
import {mockDependencies} from "../../dev/mocks";
import {RenderStringArrayAsTags} from "../../renders/renders";
import {DeleteDependency, GetDependencies} from '../../api/api'
import {RouteComponentProps} from "react-router-dom";
import Search from "antd/es/input/Search";

interface DependencyPageState {
    loading: boolean
    page: number
    pageSize: number
    total: number
    searchService: string
    selectedDependency: Dependency
    dependencies: Dependency[]
    showDeleteConfirm: boolean
}

interface DependencyPageProps extends RouteComponentProps {

}

class DependencyPage extends React.Component<DependencyPageProps, DependencyPageState> {
    state = {
        page: 0,
        pageSize: 10,
        loading: true,
        total: 0,
        searchService: "",
        selectedDependency: {} as Dependency,
        dependencies: mockDependencies,
        showDeleteConfirm: false
    };

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
                    <Button type="link" size="small"
                            onClick={this.onClickUpdateButton(record)}>Update</Button>
                    <Divider type="vertical"/>
                    <Button type="link" size="small"
                            onClick={this.onClickDeleteButton(record.service_name)}>Delete</Button>
                </div>
            ),
        }
    ];

    constructor(props: any) {
        super(props);
        this.bindCallbacks()
    }

    bindCallbacks() {
        this.onClickUpdateButton = this.onClickUpdateButton.bind(this);
        this.onClickCreateButton = this.onClickCreateButton.bind(this);
        this.onInputServiceSearch = this.onInputServiceSearch.bind(this);
        this.onSearchService = this.onSearchService.bind(this);
        this.onClickDeleteButton = this.onClickDeleteButton.bind(this);
        this.onChangePage = this.onChangePage.bind(this)
    }

    onClickUpdateButton(record: Dependency) {
        return () => {
            this.props.history.push(`/dependency/update?service=${record.service_name}`)
        }
    }

    onClickDeleteButton(service_name: string) {
        return () => {
            const reloadPage = this.reloadPage.bind(this);
            Modal.confirm({
                title: `Do you want to delete ${service_name}?`,
                onOk: () => {
                    DeleteDependency(service_name);
                    return new Promise((resolve) => {
                        setTimeout(() => {
                            reloadPage();
                            resolve()
                        }, 1000);
                    });
                },
            });
        };
    }

    onClickCreateButton() {
        this.props.history.push(`/dependency/new`)
    }

    onInputServiceSearch(e: any) {
        this.setState({searchService: e.target.value})
    }

    onSearchService() {
        this.reloadPage()
    }

    onChangePage(pageNum: number) {
        this.setState({page: pageNum - 1}, this.reloadPage)
    }

    componentDidMount() {
        this.reloadPage()
    }

    reloadPage() {
        this.setState({loading: true}, this.getDependencies)
    }

    async getDependencies() {
        let res = await GetDependencies(this.state.page, this.state.searchService);
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
        const {page, pageSize, total, loading, dependencies} = this.state;
        return (
            <Fragment>
                <Row style={{marginBottom: 15}}>
                    <Col span={6}>
                        <Search
                            placeholder="Service Name"
                            enterButton
                            onSearch={value => {
                                this.setState({searchService: value}, this.reloadPage)
                            }}
                            allowClear={true}
                        />
                    </Col>
                    <Col span={2} offset={16}>
                        <Button
                            onClick={this.onClickCreateButton}
                            style={{textAlign: "right"}}
                            type="primary">
                            <Icon type="plus"/>
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
                        current: page + 1,
                        total: total,
                        onChange: this.onChangePage,
                    }}
                />
            </Fragment>

        )
    }
}

export default DependencyPage

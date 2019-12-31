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

import {ProxyConfig} from "../../models/proxy-config";
import React, {Component, Fragment} from "react";
import {mockProxyConfigs} from "../../dev/mocks";
import {ColumnProps} from "antd/es/table";
import {DeleteProxyConfig, GetProxyConfigs} from "../../api/api";
import {Button, Col, Divider, Icon, Modal, Row, Table} from "antd";
import {RouteComponentProps} from "react-router-dom";
import Search from "antd/es/input/Search";

interface ProxyConfigPageState {
    loading: boolean
    page: number
    pageSize: number
    total: number
    searchService: string
    proxyConfigs: ProxyConfig[]
    showCreateModal: boolean,
}

interface ProxyConfigPageProps extends RouteComponentProps {
}

class ProxyConfigPage extends Component<ProxyConfigPageProps, ProxyConfigPageState> {
    state = {
        page: 0,
        pageSize: 10,
        loading: true,
        total: 0,
        searchService: "",
        proxyConfigs: mockProxyConfigs,
        showCreateModal: false,
    };

    columns: ColumnProps<ProxyConfig>[] = [
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
            title: 'Protocol',
            key: 'config.protocol',
            dataIndex: 'config.protocol'
        },
        {
            title: 'Operation',
            render: (record: ProxyConfig) => (
                <div>
                    <a onClick={() => {
                        this.props.history.push(`/proxy-configs/${record.service_name}`)
                    }}>Update</a>
                    <Divider type="vertical"/>
                    <a onClick={this.onClickDeleteButton(record.service_name)}>Delete</a>
                </div>
            ),
        }
    ];

    constructor(props: ProxyConfigPageProps) {
        super(props);
        this.onClickDeleteButton = this.onClickDeleteButton.bind(this)
    }

    componentDidMount(): void {
        this.reloadPage()
    }

    onClickDeleteButton(service_name: string) {
        return () => {
            const reloadPage = this.reloadPage.bind(this);
            Modal.confirm({
                title: `Do you want to delete ${service_name}?`,
                onOk: () => {
                    DeleteProxyConfig(service_name);
                    return new Promise((resolve) => {
                        setTimeout(() => {
                            reloadPage();
                            resolve()
                        }, 1500);
                    });
                },
            });
        };
    }

    reloadPage() {
        this.setState({loading: true}, this.getProxyConfigs)
    }

    async getProxyConfigs() {
        let res = await GetProxyConfigs(this.state.page, this.state.searchService);
        if (!res) {
            return
        }
        this.setState({
            page: res.page_num,
            pageSize: res.page_size,
            total: res.total,
            proxyConfigs: res.data,
            loading: false
        })
    }

    render() {
        const {page, pageSize, total, loading, proxyConfigs} = this.state
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
                            onClick={() => {
                                this.props.history.push("/proxy-configs/new")
                            }}
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
                    dataSource={proxyConfigs}
                    loading={loading}
                    pagination={{
                        pageSize: pageSize,
                        current: page + 1,
                        total: total,
                        onChange: (pageNum) => {
                            this.setState({page: pageNum - 1}, this.reloadPage)
                        },
                    }}
                />
            </Fragment>
        )
    }
}

export default ProxyConfigPage
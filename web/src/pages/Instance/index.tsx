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

import React, {FormEvent, Fragment} from 'react'
import {Button, Form, Icon, Input, Row, Table} from 'antd'
import {Instance} from '../../models/instance';
import {mockInstances} from '../../dev/mocks';
import {ColumnProps} from "antd/lib/table/interface";
import {GetInstances} from "../../api/api";
import {Link, RouteComponentProps} from "react-router-dom";

interface InstancePageState {
    loading: boolean
    page: number
    pageSize: number
    total: number
    instances: Instance[],
    searchID: string
    searchBelongService: string
}

interface InstancePageProps extends RouteComponentProps {
}

class InstancePage extends React.Component<InstancePageProps, InstancePageState> {
    state = {
        page: 0,
        pageSize: 10,
        loading: true,
        total: 0,
        instances: mockInstances,
        searchID: "",
        searchBelongService: ""
    };

    columns: ColumnProps<Instance>[] = [
        {
            title: 'Instance ID',
            key: 'id',
            dataIndex: 'id',
            render: (text, record) => {
                return (
                    <Link to={`/instance/${record.id}`}>{text}</Link>
                )
            }
        }
        ,
        {
            title: 'Hostname',
            key: 'hostname',
            dataIndex: 'hostname',
        },
        {
            title: 'IP',
            key: 'ip',
            dataIndex: 'ip',
        },
        {
            title: 'Port',
            key: 'port',
            dataIndex: 'port'
        },
        {
            title: 'Version',
            key: 'version',
            dataIndex: 'version',
        },
        {
            title: 'Belong Service',
            key: "belong_service",
            dataIndex: "belong_service",
        },
    ];

    constructor(props: any) {
        super(props);
        this.bindCallbacks()
    }

    componentDidMount() {
        this.reloadPage()
    }

    bindCallbacks() {
        this.onChangePage = this.onChangePage.bind(this);
        this.onSearch = this.onSearch.bind(this);
    }

    onChangePage(pageNum: number) {
        this.setState({page: pageNum - 1}, this.reloadPage)
    }

    onSearch(e: FormEvent<HTMLFormElement>) {
        e.preventDefault();
        this.reloadPage();
    }

    reloadPage() {
        this.setState({loading: true}, this.getInstances)
    }

    async getInstances() {
        let res = await GetInstances(this.state.page, this.state.searchID, this.state.searchBelongService);
        if (!res) {
            return
        }
        this.setState({
            page: res.page_num,
            pageSize: res.page_size,
            total: res.total,
            instances: res.data,
            loading: false
        })
    }

    render() {
        const {page, pageSize, total, loading, instances} = this.state;
        return (
            <Fragment>
                <Form layout="inline" style={{textAlign: "left", marginBottom: 15}} onSubmit={this.onSearch}>
                    <Form.Item>
                        <Input placeholder="Instance ID" allowClear={true} onChange={event => {
                            this.setState({searchID: event.target.value})
                        }}/>
                    </Form.Item>
                    <Form.Item>
                        <Input placeholder="Belong Service" allowClear={true} onChange={event => {
                            this.setState({searchBelongService: event.target.value})
                        }}/>
                    </Form.Item>
                    <Form.Item>
                        <Button type="primary" htmlType="submit">
                            <Icon type="search"/>
                            Search
                        </Button>
                    </Form.Item>
                </Form>


                <Row>
                    <Table
                        rowKey="id"
                        columns={this.columns}
                        dataSource={instances}
                        loading={loading}
                        pagination={{
                            pageSize: pageSize,
                            current: page + 1,
                            total: total,
                            onChange: this.onChangePage
                        }}
                    />
                </Row>
            </Fragment>
        )
    }
}

export default InstancePage
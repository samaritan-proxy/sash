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

import React, {Fragment} from "react";
import {Instance} from "../../models/instance";
import {RouteComponentProps} from "react-router-dom";
import {GetInstance} from "../../api/api";
import {Entry, Object2Array} from "../../utils/utils";
import {ColumnProps} from "antd/lib/table/interface";
import {Table} from "antd";

interface InstanceDetailPageState {
    loading: boolean
    id: string
    instance: Instance
}

interface InstanceDetailPageProps<T> extends RouteComponentProps<T> {
}

interface RouteParams {
    id: string
}

class InstanceDetailPage extends React.Component<InstanceDetailPageProps<RouteParams>, InstanceDetailPageState> {
    state = {
        loading: true,
        id: "",
        instance: {} as Instance
    };

    columns: ColumnProps<Entry>[] = [
        {
            key: 'Key',
            dataIndex: 'Key',
            render: (text) => {
                return (
                    <b>{text}</b>
                )
            }
        },
        {
            key: 'Value',
            dataIndex: 'Value'
        }
    ];

    keyMap: Map<string, string> = new Map<string, string>([
        ["create_time", "Create Time"],
        ["update_time", "Update Time"],
        ["id", "Instance ID"],
        ["hostname", "Hostname"],
        ["ip", "IP"],
        ["port", "Port"],
        ["version", "Version"],
        ["belong_service", "Belong Service"]
    ]);

    componentDidMount() {
        this.setState({
            loading: true,
            id: this.props.match.params.id,
        }, this.getInstance)
    }

    async getInstance() {
        let instance = await GetInstance(this.state.id);
        if (!instance) {
            return
        }
        this.setState({
            loading: false,
            instance: instance,
        })
    }

    render() {
        const {instance, loading} = this.state;
        return (
            <Fragment>
                <Table
                    columns={this.columns}
                    rowKey={undefined}
                    dataSource={Object2Array(instance, this.keyMap)}
                    loading={loading}
                    showHeader={false}
                    pagination={false}
                />
            </Fragment>
        )
    }
}

export default InstanceDetailPage

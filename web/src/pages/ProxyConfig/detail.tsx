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
import {RouteComponentProps} from "react-router-dom";
import {ProxyConfig} from "../../models/proxy-config";
import {GetProxyConfig, PutProxyConfig} from "../../api/api";
import {Button, Form, Row, Spin} from "antd";
import TextArea from "antd/es/input/TextArea";

interface ProxyConfigDetailPageState {
    loading: boolean
    service: string
    proxyConfig: ProxyConfig
    proxyConfigValue: string
}

interface ProxyConfigDetailPageProps<T> extends RouteComponentProps<T> {
}

interface RouteParams {
    service: string
}

class ProxyConfigDetailPage extends React.Component<ProxyConfigDetailPageProps<RouteParams>, ProxyConfigDetailPageState> {
    state = {
        loading: true,
        service: "",
        proxyConfig: {} as ProxyConfig,
        proxyConfigValue: ""
    };

    componentDidMount(): void {
        this.setState({
            loading: true,
            service: this.props.match.params.service
        }, this.getProxyConfig)
    }

    async getProxyConfig() {
        let cfg = await GetProxyConfig(this.state.service);
        if (!cfg) {
            return
        }
        this.setState({
            loading: false,
            proxyConfig: cfg,
            proxyConfigValue: JSON.stringify(cfg.config, null, 2)
        })
    }

    render() {
        const {loading} = this.state;
        return (
            <Fragment>
                <Spin spinning={loading}>
                    <Form>
                        <Row>
                            <TextArea
                                autoSize={{minRows: 20, maxRows: 20}}
                                onChange={(e) => {
                                    this.setState({proxyConfigValue: e.target.value})
                                }}
                                value={this.state.proxyConfigValue}/>
                        </Row>
                        <Row>
                            <Button type="primary" block size="large" onClick={() => {
                                this.setState({
                                    proxyConfig: {
                                        service_name: this.state.service,
                                        config: JSON.parse(this.state.proxyConfigValue)
                                    }
                                }, () => {
                                    PutProxyConfig(this.state.proxyConfig)
                                });
                            }}>Update</Button>
                        </Row>
                    </Form>
                </Spin>
            </Fragment>
        );
    }
}

export default ProxyConfigDetailPage;
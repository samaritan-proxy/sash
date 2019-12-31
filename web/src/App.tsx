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

import React from 'react';
import Home from './pages/Home'
import './App.css';
import {Route, Router, Switch} from 'react-router';
import {createBrowserHistory} from 'history'
import {Icon, Layout, Menu} from 'antd';
import {Link} from 'react-router-dom';
import DependencyPage from './pages/Dependency'
import NewDependencyPage from './pages/Dependency/CreateDependency';
import UpdateDependencyPage from './pages/Dependency/UpdateDependency'
import InstancePage from "./pages/Instance";
import ProxyConfigPage from "./pages/ProxyConfig";
import ProxyConfigDetailPage from "./pages/ProxyConfig/detail";
import InstanceDetailPage from "./pages/Instance/detail";
import NotFoundPage from "./pages/Errors/404";
import CreateProxyConfigPage from "./pages/ProxyConfig/CreateProxyConfig";

const {Header, Content, Footer, Sider} = Layout;

const URLkey: any = {
    "/": "1",
    "/dependency": "2",
    "/instance": "3",
    "/proxy-configs": "4",
};

class App extends React.Component {
    state = {
        collapsed: false,
    };

    render() {
        const {collapsed} = this.state;
        const history = createBrowserHistory();
        const path = history.location.pathname;
        let currentKey = URLkey[path];
        return (
            <div className="App">
                <Router history={history}>
                    <Layout style={{minHeight: '100vh'}}>
                        <Sider collapsible collapsed={collapsed}
                               onCollapse={(collapsed: boolean) => this.setState({collapsed})}>
                            <div className='logo'>
                                <img style={{width: "100%"}} src={'/logo.png'} alt=""/>
                            </div>
                            <Menu className="menu" theme="dark" mode="inline" defaultSelectedKeys={[currentKey]}>
                                <Menu.Item key="1">
                                    <Link to="/">
                                        <Icon type="home"/>
                                        <span>Home</span>
                                    </Link>
                                </Menu.Item>
                                <Menu.Item key="2">
                                    <Link to="/dependency">
                                        <Icon type="apartment"/>
                                        <span>Dependency</span>
                                    </Link>
                                </Menu.Item>
                                <Menu.Item key="3">
                                    <Link to="/instance">
                                        <Icon type="deployment-unit"/>
                                        <span>Instance</span>
                                    </Link>
                                </Menu.Item>
                                <Menu.Item key="4">
                                    <Link to="/proxy-configs">
                                        <Icon type="setting"/>
                                        <span>Proxy Config</span>
                                    </Link>
                                </Menu.Item>
                            </Menu>
                        </Sider>
                        <Layout>
                            <Header style={{background: '#fff', padding: 0}}/>
                            <Content style={{margin: '24px 16px 0'}}>
                                <div style={{padding: 24, background: '#fff', minHeight: 360}}>
                                    <Switch>
                                        <Route exact={true} path="/" component={Home}/>
                                        <Route path="/dependency/new" component={NewDependencyPage}/>
                                        <Route path="/dependency/update" component={UpdateDependencyPage}/>
                                        <Route path="/dependency" component={DependencyPage}/>
                                        <Route path="/instance/:id" component={InstanceDetailPage}/>
                                        <Route path="/instance" component={InstancePage}/>
                                        <Route path="/proxy-configs/new" component={CreateProxyConfigPage}/>
                                        <Route path="/proxy-configs/:service" component={ProxyConfigDetailPage}/>
                                        <Route path="/proxy-configs" component={ProxyConfigPage}/>
                                        <Route path="/error/404" component={NotFoundPage}/>

                                        <Route component={NotFoundPage}/>
                                    </Switch>
                                </div>
                            </Content>
                            <Footer style={{textAlign: 'center'}}>Sash ©2019 Samaritan Proxy</Footer>
                        </Layout>
                    </Layout>
                </Router>
            </div>
        );
    }
}

export default App;
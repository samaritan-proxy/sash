import React from 'react';
import Home from './pages/Home'
import Test from './pages/Dependency'
import './App.css';

import { Router, Route, Switch } from 'react-router';
import { createBrowserHistory } from 'history'
import { Layout, Menu, Icon } from 'antd';
import { Link } from 'react-router-dom';


const { Header, Content, Footer, Sider } = Layout;

const URLkey: any = {
  "/": "1",
  "/dependency": "2",
}

class App extends React.Component {
  state = {
    collapsed: false,
  }
  render() {
    const { collapsed } = this.state
    const history = createBrowserHistory()
    const path = history.location.pathname
    let currentKey = URLkey[path]
    return (
      <div className="App">
        <Router history={history} >
          <Layout style={{ minHeight: '100vh' }}>
            <Sider collapsible collapsed={collapsed} onCollapse={(collapsed: boolean) => this.setState({collapsed})}>
              <div className='logo'>
                <img style={{width: "100%"}} src={'/logo.png'} alt="" />
              </div>
              <Menu theme="dark" mode="inline" defaultSelectedKeys={[currentKey]}>
                <Menu.Item key="1">
                  <Link to="/">
                    <Icon type="home" />
                    <span>Home</span>
                  </Link>
                </Menu.Item>
                <Menu.Item key="2" >
                  <Link to="/dependency">
                    <Icon type="user" />
                    <span>Dependency</span>
                  </Link>
                </Menu.Item>
              </Menu>
            </Sider>
            <Layout>
              <Header style={{ background: '#fff', padding: 0 }} />
              <Content style={{ margin: '24px 16px 0' }}>
                <div style={{ padding: 24, background: '#fff', minHeight: 360 }}>
                  <Switch>
                    <Route exact={true} path="/" component={Home} />
                    <Route path="/dependency" component={Test} />
                  </Switch>
                </div>
              </Content>
              <Footer style={{ textAlign: 'center' }}>Sash ©2019 SamProxy</Footer>
            </Layout>
          </Layout>
        </Router>
      </div>
    );
  }
}

export default App;

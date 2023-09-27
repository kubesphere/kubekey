import React from 'react';
import {Column, Columns} from "@kube-design/components";
import logo from './assets/kubekey-logo.svg';
import Install from "./Components/Install/Install";
import {Route, Switch} from "react-router-dom";
import Cluster from "./Components/Cluster/Cluster";
import DeleteNode from "./Components/DeleteNode/DeleteNode";
import {ClusterTableProvider} from "./context/ClusterTableContext";
import {DeleteNodeFormProvider} from "./context/DeleteNodeFormContext";
import AddNode from "./Components/AddNode/AddNode";
import {AddNodeFormProvider} from "./context/AddNodeFormContext";
import {InstallFormProvider} from "./context/InstallFormContext";
import UpgradeCluster from "./Components/UpgradeCluster/UpgradeCluster";
import {GlobalProvider} from "./context/GlobalContext";
import {DeleteClusterProvider} from "./context/DeleteClusterContext";
import DeleteCluster from "./Components/DeleteCluster/DeleteCluster";
import {UpgradeClusterFormProvider} from "./context/UpgradeClusterFormContext";
const App = () => {
  return (
      <div>
          <GlobalProvider>
          <Columns>
              <Column className={'is-1'}></Column>
              <Column className={'is-2'}>
                  <img src={logo} alt='logo' style={{width:'70%',height:'70%',marginTop: '10px'}}></img>
              </Column>
          </Columns>
          <Switch>
              <Route exact path="/">
                  <Cluster/>
              </Route>
              <Route path="/install">
                  <InstallFormProvider>
                    <Install/>
                  </InstallFormProvider>
              </Route>
              <Route path="/DeleteCluster/:clusterName">
                  <ClusterTableProvider>
                      <DeleteClusterProvider>
                          <DeleteCluster/>
                      </DeleteClusterProvider>
                  </ClusterTableProvider>
              </Route>
              <Route path="/DeleteNode/:clusterName">
                  <ClusterTableProvider>
                      <DeleteNodeFormProvider>
                          <DeleteNode/>
                      </DeleteNodeFormProvider>
                  </ClusterTableProvider>
              </Route>
              <Route path="/AddNode/:clusterName">
                  <ClusterTableProvider>
                      <AddNodeFormProvider>
                          <AddNode/>
                      </AddNodeFormProvider>
                  </ClusterTableProvider>
              </Route>
              <Route path="/UpgradeCluster/:clusterName">
                  <ClusterTableProvider>
                      <UpgradeClusterFormProvider>
                          <UpgradeCluster/>
                      </UpgradeClusterFormProvider>
                  </ClusterTableProvider>
              </Route>
              {/* TODO 增加用户输入/AddNode/xxx错误集群名时的报错处理*/}
              <Route path="*">
                <div>路径错误</div>
              </Route>
          </Switch>
          </GlobalProvider>
      </div>
  )
}
export default App;

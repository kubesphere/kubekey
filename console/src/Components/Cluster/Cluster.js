import React from 'react';
import {Button, Column, Columns} from "@kube-design/components";
import {Link} from "react-router-dom";
import ClusterTable from "./ClusterTable";
import {ClusterTableProvider} from "../../context/ClusterTableContext";

const Cluster = () => {
    return (
        <div>
            <Columns>
                <Column className={'is-1'}></Column>
                <Column className={'is-2'}>
                    <h2>集群列表</h2>
                </Column>
                <Column className={'is-8'}>
                    <Columns>
                        <Column className={'is-10'}>
                        </Column>
                        <Column>
                            <Link to='/install'>
                                <Button>新建集群</Button>
                            </Link>
                        </Column>
                    </Columns>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-1'}></Column>
                <Column className={'is-10'}>
                    <ClusterTableProvider>
                        <ClusterTable/>
                    </ClusterTableProvider>
                </Column>
            </Columns>
        </div>
    );
};

export default Cluster;

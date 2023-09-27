import React, {useEffect} from 'react';
import {Link, useParams} from "react-router-dom";
import useClusterTableContext from "../../hooks/useClusterTableContext";
import {Button, Column, Columns} from "@kube-design/components";
import useDeleteClusterContext from "../../hooks/useDeleteClusterContext";
import DeleteClusterProgressBar from "./DeleteClusterProgressBar";
import DeleteClusterForm from "./DeleteClusterForm";

const DeleteCluster = () => {
    const {clusterName} = useParams()
    const {clusterData} = useClusterTableContext()
    const {setCurCluster,canToHome} = useDeleteClusterContext()
    useEffect(() => {
        if (clusterData.length > 0) {
            setCurCluster(clusterData.find(item=>item.metadata.name===clusterName))
        }
    }, [clusterData]);
    return (
        <>
            <Columns>
                <Column className="is-1"></Column>
                <Column className="is-2">
                    <h2>删除集群</h2>
                </Column>
                <Column className={'is-8'}>
                    <Columns>
                        <Column className={'is-10'}>
                        </Column>
                        <Column>
                            {canToHome ? (
                                <Link to='/'>
                                    <Button disabled={!canToHome}>集群列表</Button>
                                </Link>
                            ) : (
                                <Button disabled={!canToHome}>集群列表</Button>
                            )}
                        </Column>
                    </Columns>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-1'}></Column>
                <Column className={'is-2'}>
                    <DeleteClusterProgressBar></DeleteClusterProgressBar>
                </Column>
                <Column className={'is-8'}>
                    <DeleteClusterForm/>
                </Column>
            </Columns>
        </>
    );
};

export default DeleteCluster;

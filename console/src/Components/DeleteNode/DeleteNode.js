import React, {useEffect} from 'react';
import {Link, useParams} from "react-router-dom";
import {Button, Column, Columns} from "@kube-design/components";
import DeleteNodeProgressBar from "./DeleteNodeProgressBar";
import DeleteNodeForm from "./DeleteNodeForm";
import useDeleteNodeFormContext from "../../hooks/useDeleteNodeFormContext";
import useClusterTableContext from "../../hooks/useClusterTableContext";

const DeleteNode = () => {
    const {clusterName} = useParams()
    const {clusterData} = useClusterTableContext()
    const {setCurCluster,canToHome} = useDeleteNodeFormContext()
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
                        <h2>删除节点</h2>
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
                        <DeleteNodeProgressBar></DeleteNodeProgressBar>
                    </Column>
                    <Column className={'is-8'}>
                        <DeleteNodeForm/>
                    </Column>
                </Columns>
            </>
    );
};

export default DeleteNode;

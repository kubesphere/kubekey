import React, {useEffect} from 'react';
import {Link, useParams} from "react-router-dom";
import {Button, Column, Columns} from "@kube-design/components";
import AddNodeProgressBar from "./AddNodeProgressBar";
import AddNodeForm from "./AddNodeForm";
import useClusterTableContext from "../../hooks/useClusterTableContext";
import useAddNodeFormContext from "../../hooks/useAddNodeFormContext";

const AddNode = () => {
    const {clusterName} = useParams()
    const {clusterData} = useClusterTableContext()
    const {setCurCluster,canToHome} = useAddNodeFormContext()
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
                    <h2>新增节点</h2>
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
                    <AddNodeProgressBar></AddNodeProgressBar>
                </Column>
                <Column className={'is-8'}>
                    <AddNodeForm/>
                </Column>
            </Columns>
        </>
    );
};

export default AddNode;

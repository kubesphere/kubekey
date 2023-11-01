import { Button, Column, Columns } from "@kube-design/components";
import React, { useEffect } from "react";
import { Link, useParams } from "react-router-dom";
import useClusterTableContext from "../../hooks/useClusterTableContext";
import useDeleteNodeFormContext from "../../hooks/useDeleteNodeFormContext";
import DeleteNodeForm from "./DeleteNodeForm";
import DeleteNodeProgressBar from "./DeleteNodeProgressBar";

const DeleteNode = () => {
  const { clusterName } = useParams();
  const { clusterData } = useClusterTableContext();
  const { setCurCluster, canToHome } = useDeleteNodeFormContext();
  useEffect(() => {
    if (clusterData.length > 0) {
      setCurCluster(
        clusterData.find((item) => item.metadata.name === clusterName)
      );
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [clusterData]);
  return (
    <>
      <Columns>
        <Column className="is-1"></Column>
        <Column className="is-2">
          <h2>删除节点</h2>
        </Column>
        <Column className={"is-8"}>
          <Columns>
            <Column className={"is-10"}></Column>
            <Column>
              {canToHome ? (
                <Link to="/">
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
        <Column className={"is-1"}></Column>
        <Column className={"is-2"}>
          <DeleteNodeProgressBar></DeleteNodeProgressBar>
        </Column>
        <Column className={"is-8"}>
          <DeleteNodeForm />
        </Column>
      </Columns>
    </>
  );
};

export default DeleteNode;

import { Button, Column, Columns } from "@kube-design/components";
import React, { useEffect } from "react";
import { Link, useParams } from "react-router-dom";
import useClusterTableContext from "../../hooks/useClusterTableContext";
import useUpgradeClusterFormContext from "../../hooks/useUpgradeClusterFormContext";
import UpgradeClusterForm from "./UpgradeClusterForm";
import UpgradeClusterProgressBar from "./UpgradeClusterProgressBar";

const UpgradeCluster = () => {
  const { clusterName } = useParams();
  const { clusterData } = useClusterTableContext();
  const { canToHome, setCurCluster, setOriginalClusterVersion } =
    useUpgradeClusterFormContext();
  useEffect(() => {
    if (clusterData.length > 0) {
      const newV = clusterData.find(
        (item) => item.metadata.name === clusterName
      );
      setOriginalClusterVersion(newV.spec.kubernetes.version);
      newV.spec.kubernetes.version = "";
      setCurCluster(newV);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [clusterData]);
  return (
    <>
      <Columns>
        <Column className="is-1"></Column>
        <Column className="is-2">
          <h2>升级集群</h2>
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
          <UpgradeClusterProgressBar></UpgradeClusterProgressBar>
        </Column>
        <Column className={"is-8"}>
          <UpgradeClusterForm />
        </Column>
      </Columns>
    </>
  );
};

export default UpgradeCluster;

import {
  Column,
  Columns,
  Input,
  Radio,
  Select,
  Toggle,
  Tooltip,
} from "@kube-design/components";
import React, { useEffect, useState } from "react";
import useGlobalContext from "../../../hooks/useGlobalContext";
import useUpgradeClusterFormContext from "../../../hooks/useUpgradeClusterFormContext";

const UpgradeClusterSetting = () => {
  const [clusterVersionOptions, setClusterVersionOptions] = useState([]);
  const { curCluster, handleChange, originalClusterVersion } =
    useUpgradeClusterFormContext();
  const { backendIP } = useGlobalContext();
  const changeClusterVersionHandler = (e) => {
    handleChange("spec.kubernetes.version", e);
    handleChange("spec.kubernetes.containerManager", "");
  };
  const changeClusterNameHandler = (e) => {
    handleChange("spec.kubernetes.clusterName", e.target.value);
    handleChange("metadata.name", e.target.value);
  };
  const changeAutoRenewHandler = (e) => {
    handleChange("spec.kubernetes.autoRenewCerts", e);
  };
  const changeContainerManagerHandler = (e) => {
    handleChange("spec.kubernetes.containerManager", e.target.name);
  };
  const compareVersion = (a, b) => {
    // 去除"v"并按点拆分版本字符串
    const partsA = a.substring(1).split(".").map(Number);
    const partsB = b.substring(1).split(".").map(Number);

    // 找到最长的版本长度
    const maxLength = Math.max(partsA.length, partsB.length);

    // 遍历并比较
    for (let i = 0; i < maxLength; i++) {
      const partA = partsA[i] || 0;
      const partB = partsB[i] || 0;

      if (partA > partB) {
        return true;
      } else if (partA < partB) {
        return false;
      }
    }
    // 如果到这里，说明版本完全相同
    return false;
  };

  useEffect(() => {
    if (Object.keys(curCluster).length > 0 && backendIP !== "") {
      fetch(`http://${backendIP}:8082/clusterVersionOptions`)
        .then((res) => {
          return res.json();
        })
        .then((data) => {
          const targetClusterVersionList = data.clusterVersionOptions.filter(
            (item) => compareVersion(item, originalClusterVersion)
          );
          setClusterVersionOptions(
            targetClusterVersionList.map((item) => ({
              value: item,
              label: item,
            }))
          );
        })
        .catch(() => {});
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [curCluster, backendIP]);

  return (
    <div>
      <Columns>
        <Column className={"is-2"}>Kubernetes目标升级版本：</Column>
        <Column>
          <Select
            value={
              Object.keys(curCluster).length > 0
                ? curCluster.spec.kubernetes.version
                : ""
            }
            options={clusterVersionOptions}
            onChange={changeClusterVersionHandler}
          />
        </Column>
      </Columns>
      <Columns>
        <Column className={"is-2"}>Kubernetes集群名称:</Column>
        <Column>
          <Input
            onChange={changeClusterNameHandler}
            value={
              Object.keys(curCluster).length > 0 ? curCluster.metadata.name : ""
            }
            placeholder="请输入要创建的Kubernetes集群名称"
          />
        </Column>
      </Columns>
      <Columns>
        <Column className={"is-2"}>是否自动续费证书:</Column>
        <Column>
          <Toggle
            checked={
              Object.keys(curCluster).length > 0
                ? curCluster.spec.kubernetes.autoRenewCerts
                : false
            }
            onChange={changeAutoRenewHandler}
            onText="开启"
            offText="关闭"
          />
        </Column>
      </Columns>
      <Columns>
        <Column className={"is-2"}>容器运行时：</Column>
        <Column>
          <Tooltip content={"v1.24.0及以上版本集群不支持docker作为容器运行时"}>
            <Radio
              name="docker"
              checked={
                Object.keys(curCluster).length > 0
                  ? curCluster.spec.kubernetes.containerManager === "docker"
                  : false
              }
              onChange={changeContainerManagerHandler}
              disabled={
                Object.keys(curCluster).length > 0
                  ? curCluster.spec.kubernetes.version >= "v1.24.0"
                  : true
              }
            >
              Docker
            </Radio>
          </Tooltip>
          <Radio
            name="containerd"
            checked={
              Object.keys(curCluster).length > 0
                ? curCluster.spec.kubernetes.containerManager === "containerd"
                : false
            }
            onChange={changeContainerManagerHandler}
          >
            Containerd
          </Radio>
        </Column>
      </Columns>
    </div>
  );
};

export default UpgradeClusterSetting;

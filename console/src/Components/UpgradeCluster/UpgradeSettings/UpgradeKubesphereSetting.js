import React from 'react';
import {Column, Columns, Select, Toggle, Tooltip} from "@kube-design/components";
import useUpgradeClusterFormContext from "../../../hooks/useUpgradeClusterFormContext";

const UpgradeKubesphereSetting = () => {
    const { ksEnable, setKsEnable, ksVersion, setKsVersion} = useUpgradeClusterFormContext()
    const KubesphereVersionOptions = [{label:"v3.4.0",value:"v3.4.0"}]
    const changeInstallKubesphereHandler = (e) => {
        setKsVersion('')
        setKsEnable(e)
    }
    const changeKubesphereVersionHandler = e => {
        setKsVersion(e)
    }

    return (
        <div>
            <Columns>
                <Column className={'is-2'}>是否安装Kubesphere:</Column>
                <Column>
                    <Tooltip content="升级k8s必须同时安装v3.4.0版本KubeSphere" placement="right" >
                        <Toggle checked={ksEnable} onChange={changeInstallKubesphereHandler} onText="开启" offText="关闭" disabled={true}/>
                    </Tooltip>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    Kubesphere版本：
                </Column>
                <Column>
                    <Select placeholder="Kubesphere可选版本与K8s集群版本有关" value={ksVersion} options={KubesphereVersionOptions} disabled={true} onChange={changeKubesphereVersionHandler} />
                </Column>
            </Columns>
        </div>
    );
};

export default UpgradeKubesphereSetting;

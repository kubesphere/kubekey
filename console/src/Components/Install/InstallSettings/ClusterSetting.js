import React, {useEffect, useState} from 'react';
import useInstallFormContext from "../../../hooks/useInstallFormContext";
import {Column, Input, Columns, Select, Toggle, RadioGroup, RadioButton} from "@kube-design/components";
import useGlobalContext from "../../../hooks/useGlobalContext";

const ClusterSetting = () => {
    const {backendIP} = useGlobalContext();
    const [clusterVersionOptions,setClusterVersionOptions] = useState([])

    const { data, handleChange} = useInstallFormContext()
    // const { data, handleChange, KubekeyNamespace, setKubekeyNamespace} = useInstallFormContext()
    const changeClusterVersionHandler = e => {
        handleChange('spec.kubernetes.version',e)
        handleChange('spec.kubernetes.containerManager','')
    }
    const changeClusterNameHandler = e => {
        handleChange('spec.kubernetes.clusterName',e.target.value)
        handleChange('metadata.name',e.target.value)
    }
    const changeAutoRenewHandler = e => {
        handleChange('spec.kubernetes.autoRenewCerts',e)
    }
    const changeContainerManagerHandler = e => {
        handleChange('spec.kubernetes.containerManager',e)
    }
    // const changeKubekeyNamespaceHandler = e => {
    //     setKubekeyNamespace(e.target.value)
    // }
    useEffect(()=>{
        if(backendIP!=='') {
            fetch(`http://${backendIP}:8082/clusterVersionOptions`)
                // fetch('http://139.196.14.61:8082/clusterVersionOptions')
                .then((res)=>{
                    return res.json()
                }).then(data => {
                setClusterVersionOptions(data.clusterVersionOptions.map(item => ({ value: item, label: item })))
            }).catch(()=>{

            })
        }
    },[backendIP])

    const containerManagerOptions = [
        {
            label: 'Docker',
            value: 'docker'
        },
        {
            label: 'Containerd',
            value: 'containerd'
        }
    ]



    return (
        <div>
            <Columns>
                <Column className={'is-2'}>集群名称:</Column>
                <Column >
                    <Input onChange={changeClusterNameHandler} value={data.metadata.name} placeholder="请输入要创建的 Kubernetes 集群名称" />
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    Kubernetes 版本：
                </Column>
                <Column>
                    <Select value={data.spec.kubernetes.version} options={clusterVersionOptions} onChange={changeClusterVersionHandler} />
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>证书自动更新:</Column>
                <Column>
                    <Toggle checked={data.spec.kubernetes.autoRenewCerts} onChange={changeAutoRenewHandler} onText="开启" offText="关闭" />
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    容器运行时：
                </Column>
                <Column>
                    <RadioGroup buttonWidth={100} wrapClassName="radio-group-button"onChange={changeContainerManagerHandler} defaultValue={data.spec.kubernetes.containerManager}>
                        {containerManagerOptions.map(option => <RadioButton key={option.value} value={option.value}>
                            {option.label}
                        </RadioButton>)}
                    </RadioGroup>
                </Column>
            </Columns>
        </div>
    )
};

export default ClusterSetting;

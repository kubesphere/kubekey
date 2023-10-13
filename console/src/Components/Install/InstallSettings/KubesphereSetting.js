import React, {useEffect, useState} from 'react';
import {Column, Columns, Select, Toggle, Tooltip} from "@kube-design/components";
import useInstallFormContext from "../../../hooks/useInstallFormContext";
import useGlobalContext from "../../../hooks/useGlobalContext";

const KubesphereSetting = () => {
    const {backendIP} = useGlobalContext();
    const { data, ksEnable, setKsEnable, ksVersion, setKsVersion} = useInstallFormContext()
    const [KubesphereVersionOptions,setKubesphereVersionOptions] = useState([])

    useEffect(()=>{
        if(backendIP!=='') {
            fetch(`http://${backendIP}:8082/ksVersionOptions/${data.spec.kubernetes.version}`)
                .then((res)=>{
                    return res.json()
                }).then(data => {
                setKubesphereVersionOptions(data.ksVersionOptions.map(item => ({ value: item, label: item })))
            }).catch(()=>{

            })
        }
    },[backendIP])
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
                <Column className={'is-2'}>是否安装 KubeSphere:</Column>
                <Column>
                    <Tooltip content="安装 KubeSphere 需要在存储设置中开启本地存储" placement="right" >
                        <Toggle checked={ksEnable} onChange={changeInstallKubesphereHandler}
                                disabled={!(data.spec.storage
                                    && data.spec.storage.openebs
                                    && data.spec.storage.openebs.basePath
                                    && data.spec.storage.openebs.basePath!=='')}
                                onText="开启" offText="关闭"/>
                    </Tooltip>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    KubeSphere 版本：
                </Column>
                <Column>
                    <Select placeholder="KubeSphere 可选版本与 K8s 集群版本有关" value={ksVersion} options={KubesphereVersionOptions} disabled={!ksEnable} onChange={changeKubesphereVersionHandler} />
                </Column>
            </Columns>
        </div>
    );
};

export default KubesphereSetting;

import React from 'react';
import useInstallFormContext from "../../../hooks/useInstallFormContext";
import {Column, Input, RadioGroup, Columns} from "@kube-design/components";

const NetworkSetting = () => {
    const networkPluginOptions = [
        {
            value:'calico',
            label:'calico'
        },
        {
            value:'flannel',
            label:'flannel'
        },
        {
            value:'cilium',
            label:'cilium'
        },
        {
            value:'hybridnet',
            label:'hybridnet'
        },
        {
            value:'Kube-OVN',
            label:'kube-ovn'
        },
        {
            value:'',
            label:'不启用'
        }
    ]

    const { data, handleChange } = useInstallFormContext()
    const networkPluginChangeHandler = (e) => {
        handleChange('spec.network.plugin',e)
        if(e==='none') {
            handleChange('enableMultusCNI',false)
        }
    }

    const  kubePodsCIDRChangeHandler = (e) => {
            handleChange('spec.network.kubePodsCIDR',e.target.value)
    }


    const  kubeServiceCIDRChangeHandler = (e) => {
            handleChange('spec.network.kubeServiceCIDR',e.target.value)
    }

    // const changEnableMultusCNIHandler = e => {
    //     handleChange('enableMultusCNI',e)
    // }

    return (

        <div>
            <Columns>
                <Column className={'is-2'}>网络插件：</Column>
                <Column>
                    <RadioGroup value={data.spec.network.plugin} options={networkPluginOptions} onChange={networkPluginChangeHandler}>
                    </RadioGroup>
                </Column>
            </Columns>
            <Columns >
                {/*TODO 要改*/}
                <Column className={'is-2'}>kubePodsCIDR:</Column>
                <Column>
                    <Input placeholder={'10.233.64.0/18'} name='kubePodsCIDRPrefix' value={data.spec.network.kubePodsCIDR} onChange={kubePodsCIDRChangeHandler}></Input>
                </Column>

            </Columns>
            <Columns >
                <Column className={'is-2'}>kubeServiceCIDR:</Column>
                <Column>
                    <Input placeholder={'10.233.0.0/18'} name='kubeServiceCIDRPrefix' value={data.spec.network.kubeServiceCIDR} onChange={kubeServiceCIDRChangeHandler}></Input>
                </Column>
            </Columns>
            {/*<Columns>*/}
            {/*    <Column className={'is-2'}>是否开启Multus CNI:</Column>*/}
            {/*    <Column>*/}
            {/*        <Tooltip content="Multus 不能独立部署。它总是需要至少一个传统的 CNI 插件，以满足 Kubernetes 集群的网络要求。该 CNI 插件成为 Multus 的默认插件，并将被用来为所有的 pod 提供主接口。">*/}
            {/*
            {/*            <Toggle checked={data.enableMultusCNI} onChange={changEnableMultusCNIHandler} onText="开启" offText="关闭" disabled={data.spec.network.plugin==='none' || data.spec.network.plugin===''}/>*/}
            {/*        </Tooltip>*/}
            {/*    </Column>*/}
            {/*</Columns>*/}
        </div>
    );
};

export default NetworkSetting;

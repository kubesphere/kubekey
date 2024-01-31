import React from 'react';
import useInstallFormContext from "../../../hooks/useInstallFormContext";
import {Column, Columns, RadioGroup} from "@kube-design/components";
import {Select, Tag} from "@kubed/components"
const EtcdSetting = () => {
    const { data, handleChange } = useInstallFormContext()

    const initEtcdNodes = () => {
        let etcdNodes = []
        data.spec.hosts.map(host=>{
            if (data.spec.roleGroups.master.includes(host.name)) etcdNodes.push(host.name)
        })
        return etcdNodes
    }

    const ETCDTypeOptions = [{
        value: 'kubekey',
        label: 'kubekey'
    }]

    const ETCDChangeHandler = (e) => {
        handleChange('spec.roleGroups.etcd',e)
    }
    const ETCDTypeChangeHandler = e => {
        handleChange('spec.etcd.type',e)
    }
    const ETCDOptionContent = (item) => {
        return (
            <Select.Option key={item.name} value={item.name} label={item.name}>
                <div style={{display:`flex`}}>
                    <div style={{width:"200px"}}>{item.name}</div>
                    <div style={{display:`flex`}}>
                        {data.spec.roleGroups.master.includes(item.name) && <Tag style={{marginRight:"10px"}} color="error">MASTER</Tag>}
                        {data.spec.roleGroups.worker.includes(item.name) && <Tag color="secondary">WORKER</Tag>}
                    </div>
                </div>
            </Select.Option>
        )
    }

    return (
        <div>
            <Columns>
                <Column className={'is-2'}>ETCD 部署节点：</Column>
                <Column>
                    <div style={{display:`flex`}}>
                        <Select style={{minWidth:'400px'}} defaultValue={initEtcdNodes} onChange={ETCDChangeHandler} placeholder="请选择ETCD部署节点" mode="multiple" showSearch allowClear showArrow optionLabelProp="label">
                            {data.spec.hosts.map(host=>ETCDOptionContent(host))}
                        </Select>
                    </div>
                    {/*<Select options={ETCDOptions} value={data.ETCD} onChange={ETCDChangeHandler} searchable multi />*/}

                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>ETCD 部署类型：</Column>
                <Column >
                    <RadioGroup options={ETCDTypeOptions} value={data.spec.etcd.type} onChange={ETCDTypeChangeHandler} />
                </Column>
            </Columns>
        </div>
    )
};

export default EtcdSetting;

import React from 'react';
import {Column, Columns, RadioGroup} from "@kube-design/components";
import {Select, Tag} from "@kubed/components"
import useAddNodeFormContext from "../../../hooks/useAddNodeFormContext";
const AddNodeEtcdSetting = () => {
    const { curCluster, setCurCluster } = useAddNodeFormContext()

    const ETCDTypeOptions = [{
        value: 'kubekey',
        label: 'kubekey'
    }]

    const ETCDChangeHandler = (e) => {
        setCurCluster(prev=>{
            const newCluster = {...prev}
            newCluster.spec.roleGroups.etcd = e
            return newCluster
        })
    }
    const ETCDTypeChangeHandler = e => {
        setCurCluster(prev=>{
            const newCluster = {...prev}
            newCluster.spec.etcd.type = e
            return newCluster
        })
    }
    const ETCDOptionContent = (item) => {
        return (
            <Select.Option key={item.name} value={item.name} label={item.name}>
                <div style={{display:`flex`}}>
                    <div style={{width:"200px"}}>{item.name}</div>
                    <div style={{display:`flex`}}>
                        {curCluster.spec.roleGroups.master.includes(item.name) && <Tag style={{marginRight:"10px"}} color="error">MASTER</Tag>}
                        {curCluster.spec.roleGroups.worker.includes(item.name) && <Tag color="secondary">WORKER</Tag>}
                    </div>
                </div>
            </Select.Option>
        )
    }

    return (
        <div>
            <Columns>
                <Column className={'is-2'}>ETCD部署节点：</Column>
                <Column>
                    <div style={{display:`flex`}}>
                        <Select style={{minWidth:'400px'}} value={curCluster.spec.roleGroups.etcd} onChange={ETCDChangeHandler} placeholder="请选择ETCD部署节点" mode="multiple" showSearch allowClear showArrow optionLabelProp="label">
                            {curCluster.spec.hosts.map(host=>ETCDOptionContent(host))}
                        </Select>
                    </div>

                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>ETCD类型：</Column>
                <Column >
                    <RadioGroup options={ETCDTypeOptions} value={curCluster.spec.etcd.type} onChange={ETCDTypeChangeHandler} />
                </Column>
            </Columns>
        </div>
    )
};

export default AddNodeEtcdSetting;

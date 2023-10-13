import React, {useState} from 'react';
import useInstallFormContext from "../../hooks/useInstallFormContext";
import {CheckboxGroup, Column, Columns, Input, InputPassword} from "@kube-design/components";
import {Modal,Button} from "@kubed/components";

const HostEditModal = ({record}) => {
    const recordCopy = record

    const { data, handleChange } = useInstallFormContext()
    const roleCopy = []
    if (data.spec.roleGroups.master.includes(record.name)) roleCopy.push('Master');
    if (data.spec.roleGroups.worker.includes(record.name)) roleCopy.push('Worker');
    const [curRole,setCurRole] = useState(roleCopy)
    const [visible, setVisible] = React.useState(false);

    const [curHost,setCurHost] = useState(record)

    const ref = React.createRef();
    const openModal = () => {
        setVisible(true);
    };

    const closeModal = () => {
        setCurHost(recordCopy)
        // 可以不要这行
        setCurRole(roleCopy)
        setVisible(false);
    };
    const roleOptions = [
        {
            value:'Master',
            label:'Master'
        },
        {
            value:'Worker',
            label:'Worker'
        }
    ]
    const onChangeHandler = e => {
        if(Array.isArray(e)) {
            setCurRole(e)
        } else {
            setCurHost(prevState => {
                return ({...prevState,[e.target.name]:e.target.value})
            })
        }
    }
    const onOKHandler = () => {
        const newHosts = data.spec.hosts.map(host => {
            if (host.name === recordCopy.name) {
                return curHost;
            } else {
                return host;
            }
        });
        handleChange('spec.hosts',newHosts)
        // 无论改没改名，都在master和worker中删掉原名
        const otherMasters = data.spec.roleGroups.master.filter(name => name!==recordCopy.name)
        const otherWorkers = data.spec.roleGroups.worker.filter(name => name!==recordCopy.name)
        // 再加回去
        if(curRole.length===2){
            handleChange("spec.roleGroups.master",[...otherMasters,curHost.name])
            handleChange("spec.roleGroups.worker",[...otherWorkers,curHost.name])
        }
        else if(curRole[0]==='Master') {
            handleChange("spec.roleGroups.master",[...otherMasters,curHost.name])
            handleChange("spec.roleGroups.worker",[...otherWorkers])
        }
        else if(curRole[0]==='Worker') {
            handleChange("spec.roleGroups.worker",[...otherWorkers,curHost.name])
            handleChange("spec.roleGroups.master",[...otherMasters])
        }
        setVisible(false);
    }
    const modalContent = (
        <div>
            <Columns>
                <Column className={'is-2'}>
                    主机名：
                </Column>
                <Column>
                    <Input name='name' value={curHost.name} onChange={onChangeHandler}></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    Address：
                </Column>
                <Column>
                    <Input name='address' value={curHost.address} onChange={onChangeHandler}></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    InternalAddress：
                </Column>
                <Column>
                    <Input name='internalAddress' value={curHost.internalAddress} onChange={onChangeHandler}></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    角色：
                </Column>
                <Column>
                    <CheckboxGroup name='role' value={curRole} options={roleOptions} onChange={onChangeHandler} ></CheckboxGroup>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    用户名：
                </Column>
                <Column>
                    <Input name='user' value={curHost.user} onChange={onChangeHandler}></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    密码：
                </Column>
                <Column>
                    <InputPassword name='password' value={curHost.password} onChange={onChangeHandler}></InputPassword>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    id_rsa路径：
                </Column>
                <Column>
                    <Input name='privateKeyPath' value={curHost.privateKeyPath} onChange={onChangeHandler}></Input>
                </Column>
            </Columns>

        </div>
    )

    return (
        <>
            <Button variant="link" style={{marginRight:'20px'}} onClick={openModal}>编辑</Button>
            <Modal
                ref={ref}
                visible={visible}
                title="编辑节点"
                onCancel={closeModal}
                onOk={onOKHandler}
                okText={"保存"}
                cancelText={"取消"}
            >
                {modalContent}
            </Modal>
        </>
    );
}

export default HostEditModal;

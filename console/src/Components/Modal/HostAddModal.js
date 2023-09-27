import React, {useState} from 'react';
import {Modal} from "@kubed/components";
import { CheckboxGroup, Column, Columns, Input, InputPassword, Button} from "@kube-design/components";
import useInstallFormContext from "../../hooks/useInstallFormContext";
const HostAddModal = () => {

    const { data, handleChange } = useInstallFormContext()

    const [visible, setVisible] = useState(false);

    const [curRole, setCurRole] = useState([]);

    const [newHost,setNewHost] = useState({
        name : '',
        address : '',
        internalAddress : '',
        user : '',
        password : '',
        privateKeyPath : ''
    })

    const ref = React.createRef();
    const openModal = () => {
        setVisible(true);
    };

    const closeModal = () => {
        setNewHost({
            name : '',
            address : '',
            internalAddress : '',
            user : '',
            password : '',
            privateKeyPath : ''
        })
        setCurRole([])
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
            setNewHost(prevState => {
                return ({...prevState,[e.target.name]:e.target.value})
            })
        }
    }
    const onOKHandler = () => {
        handleChange('spec.hosts',[...data.spec.hosts,newHost])
        if(curRole.length===2){
            handleChange("spec.roleGroups.master",[...data.spec.roleGroups.master,newHost.name])
            handleChange("spec.roleGroups.worker",[...data.spec.roleGroups.worker,newHost.name])
        }
        else if(curRole[0]==='Master') {
            handleChange("spec.roleGroups.master",[...data.spec.roleGroups.master,newHost.name])
        }
        else if(curRole[0]==='Worker') {
            handleChange("spec.roleGroups.worker",[...data.spec.roleGroups.worker,newHost.name])
        }
        setNewHost({
            name : '',
            address : '',
            internalAddress : '',
            user : '',
            password : '',
            privateKeyPath : ''
        })
        setCurRole([])
        setVisible(false);
    }
    const modalContent = (
        <div>
            <Columns>
                <Column className={'is-2'}>
                    主机名：
                </Column>
                <Column>
                    <Input name='name' value={newHost.name} onChange={onChangeHandler}></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    Address：
                </Column>
                <Column>
                    <Input name='address' value={newHost.address} onChange={onChangeHandler}></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    InternalAddress：
                </Column>
                <Column>
                    <Input name='internalAddress' value={newHost.internalAddress} onChange={onChangeHandler}></Input>
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
                    <Input name='user' value={newHost.user} onChange={onChangeHandler}></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    密码：
                </Column>
                <Column>
                    <InputPassword name='password' value={newHost.password} onChange={onChangeHandler}></InputPassword>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    id_rsa路径：
                </Column>
                <Column>
                    <Input name='privateKeyPath' value={newHost.privateKeyPath} onChange={onChangeHandler}></Input>
                </Column>
            </Columns>

        </div>
    )

    return (
        <>
            <Button onClick={openModal}>添加节点</Button>
            <Modal
                ref={ref}
                visible={visible}
                title="添加节点"
                onCancel={closeModal}
                onOk={onOKHandler}
            >
                {modalContent}
            </Modal>
        </>
    );
}

export default HostAddModal;

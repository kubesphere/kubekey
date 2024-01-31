import React, {useState} from 'react';
import {Modal} from "@kubed/components";
import {Column, Columns, Input, InputPassword, Button, RadioGroup, RadioButton, Select} from "@kube-design/components";
import useInstallFormContext from "../../hooks/useInstallFormContext";
import {MinusSquare, PlusSquare} from "@kubed/icons";
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
        privateKeyPath : '',
        arch: 'amd64',
        labels: {}
    })

    const [sshAuthenticationType,setSshAuthenticationType] = useState('password')

    const [labels, setLabels] = useState([new Map()]);

    const onChangeLabelsHandler = e => {
        if (e.target.name.includes('key-')) {
            const index = e.target.name.split('-')[1]
            const newLabels = [...labels]
            newLabels[index] = new Map([[e.target.value, newLabels[index].values().next().value]])

            setLabels(newLabels)
        }

        if (e.target.name.includes('value-')) {
            const index = e.target.name.split('-')[1]
            const newLabels = [...labels]
            newLabels[index]= new Map([[newLabels[index].keys().next().value, e.target.value]])
            setLabels(newLabels)
        }
    }

    const addLabel = () => {
        setLabels([...labels, new Map()]);
    }

    const removeLabel = (index) => {
        const updetaLabels = [...labels];
        updetaLabels.splice(index, 1);
        setLabels(updetaLabels);
    }

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
            privateKeyPath : '',
            arch: 'amd64',
            labels: {}
        })
        setCurRole([])
        setVisible(false);
    };

    const roleOptions = [
        {
            value: 'master',
            label: 'master'
        },
        {
            value: 'worker',
            label: 'worker'
        }
    ]

    const archOptions = [
        {
            value: 'amd64',
            label:'amd64'
        },
        {
            value: 'arm64',
            label:'arm64'
        }
    ]

    const sshAuthenticationOptions = [
        {
            value: 'password',
            label:'密码'
        },
        {
            value: 'privateKey',
            label:'密钥'
        },
        {
            value: 'privateKeyPath',
            label:'密钥路径'
        }
    ]

    const onChangeRoleHandler = e => {
        if(Array.isArray(e)) {
            setCurRole(e)
        }
    }

    const onChangeHandler = e => {
        if (e === 'amd64' || e === 'arm64') {
            setNewHost(prevState => {
                return ({...prevState, arch: e})
            })
        } else {
            setNewHost(prevState => {
                return ({...prevState, [e.target.name]: e.target.value})
            })
        }
    }

    const onChangeSshAuthenticationTypeHandler = e => {
        if (e === 'password') {
            setSshAuthenticationType('password')
        } else if (e === 'privateKey') {
            setSshAuthenticationType('privateKey')
        } else if (e === 'privateKeyPath') {
            setSshAuthenticationType('privateKeyPath')
        }
    }

    const onOKHandler = () => {
        const labelsMap = () => {
            const labels = {}
            for (let i = 0; i < labels.length; i++) {
                labels[Array.from(labels[i].keys())[0]] = Array.from(labels[i].values())[0]
            }
            return labels
        }
        newHost.labels = labelsMap()
        handleChange('spec.hosts',[...data.spec.hosts,newHost])
        if(curRole.length===2){
            handleChange("spec.roleGroups.master",[...data.spec.roleGroups.master,newHost.name])
            handleChange("spec.roleGroups.worker",[...data.spec.roleGroups.worker,newHost.name])
        }
        else if(curRole[0]==='master') {
            handleChange("spec.roleGroups.master",[...data.spec.roleGroups.master,newHost.name])
        }
        else if(curRole[0]==='worker') {
            handleChange("spec.roleGroups.worker",[...data.spec.roleGroups.worker,newHost.name])
        }
        setNewHost({
            name : '',
            address : '',
            internalAddress : '',
            user : '',
            password : '',
            privateKeyPath : '',
            arch: 'amd64',
            labels: {}
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
                    <Input name='name' value={newHost.name} onChange={onChangeHandler} placeholder='必填，设置节点名称'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    SSH 地址：
                </Column>
                <Column>
                    <Input name='address' value={newHost.address} onChange={onChangeHandler} placeholder='必填，节点 SSH 链接地址'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    节点 IP：
                </Column>
                <Column>
                    <Input name='internalAddress' value={newHost.internalAddress} onChange={onChangeHandler} placeholder='必填，节点 IP 地址'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    节点角色：
                </Column>
                <Column>
                    <Select multi name='role' value={curRole} options={roleOptions} onChange={onChangeRoleHandler} placeholder='选择节点角色' ></Select>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    CPU 架构：
                </Column>
                <Column>
                    <RadioGroup buttonWidth={100} wrapClassName="radio-group-button" onChange={onChangeHandler} defaultValue='amd64'>
                        {
                            archOptions.map(option => <RadioButton key={option.value}  value={option.value}>{option.label}</RadioButton>)
                        }
                    </RadioGroup>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    SSH 用户：
                </Column>
                <Column>
                    <Input name='user' value={newHost.user} onChange={onChangeHandler} placeholder='必填，用户应具有 sudo 权限'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    SSH 认证：
                </Column>
                <Column>
                    <RadioGroup buttonWidth={120} wrapClassName="radio-group-button" onChange={onChangeSshAuthenticationTypeHandler} defaultValue='password'>
                        {
                            sshAuthenticationOptions.map(option => <RadioButton key={option.value}  value={option.value}>{option.label}</RadioButton>)
                        }
                    </RadioGroup>
                </Column>
            </Columns>
            {
                sshAuthenticationType === 'password' && (
                    <>
                        <Columns>
                            <Column className={'is-2'}>
                                SSH 密码：
                            </Column>
                            <Column>
                                <InputPassword name='password' value={newHost.password} onChange={onChangeHandler} placeholder='SSH认证方式任选其一，节点 SSH 密码' disabled={sshAuthenticationType === 'password' ? false : true}></InputPassword>
                            </Column>
                        </Columns>
                    </>
                )
            }
            {
                sshAuthenticationType === 'privateKey' && (
                    <>
                        <Columns>
                            <Column className={'is-2'}>
                                SSH 密钥：
                            </Column>
                            <Column>
                                <InputPassword name='privateKey' value={newHost.privateKey} onChange={onChangeHandler} placeholder='SSH认证方式任选其一，节点 SSH 密钥' disabled={sshAuthenticationType === 'privateKey' ? false : true}></InputPassword>
                            </Column>
                        </Columns>
                    </>
                )
            }
            {
                sshAuthenticationType === 'privateKeyPath' && (
                    <>
                        <Columns>
                            <Column className={'is-2'}>
                                密钥文件：
                            </Column>
                            <Column>
                                <Input name='privateKeyPath' value={newHost.privateKeyPath} onChange={onChangeHandler} placeholder='SSH认证方式任选其一，节点 SSH 密钥文件路径' disabled={sshAuthenticationType === 'privateKeyPath' ? false : true}></Input>
                            </Column>
                        </Columns>
                    </>
                )
            }
            {
                    labels.map((label, index) => (
                            <>
                                <Columns>
                                    <Column className={'is-2'}>
                                        {
                                            index === 0 ? '标签：' : ''
                                        }
                                    </Column>
                                    <Column>
                                        {
                                            index === 0 ? (
                                                <>
                                                    <Input style={{width: '188px'}} name={'key-'+index} value={label.size === 0 ? '' : Array.from(label.keys())[0]} onChange={onChangeLabelsHandler} placeholder='键'></Input>
                                                    <Input style={{width: '188px'}} name={'value-'+index} value={label.size === 0 ? '' : Array.from(label.values())[0]} onChange={onChangeLabelsHandler} placeholder='值'></Input>
                                                    <PlusSquare onClick={addLabel} style={{marginLeft: '10px'}} />
                                                </>
                                            ):(
                                                <>
                                                    <Input style={{width: '188px'}} name={'key-'+index} defaultValue={label.size === 0 ? '' : Array.from(label.keys())[0]} onChange={onChangeLabelsHandler} placeholder='键'></Input>
                                                    <Input style={{width: '188px'}} name={'value-'+index} defaultValue={label.size === 0 ? '' : Array.from(label.values())[0]} onChange={onChangeLabelsHandler} placeholder='值'></Input>
                                                    <PlusSquare onClick={addLabel} style={{marginLeft: '10px'}} />
                                                    <MinusSquare onClick={() => removeLabel(index)} />
                                                </>
                                            )
                                        }

                                    </Column>
                                </Columns>
                            </>
                        )

                )
            }
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
                <div style={{ paddingTop: '20px', paddingLeft: '30px', paddingBottom: '20px', paddingRight: '0px' }}>
                    {modalContent}
                </div>
            </Modal>
        </>
    );
}

export default HostAddModal;

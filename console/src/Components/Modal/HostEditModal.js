import React, {useState} from 'react';
import useInstallFormContext from "../../hooks/useInstallFormContext";
import {
    Column,
    Columns,
    Input,
    InputPassword,
    RadioButton,
    RadioGroup,
    Select
} from "@kube-design/components";
import {Modal,Button} from "@kubed/components";
import {MinusSquare, PlusSquare} from "@kubed/icons";

const HostEditModal = ({record}) => {
    const recordCopy = record

    const { data, handleChange } = useInstallFormContext()
    const roleCopy = []
    if (data.spec.roleGroups.master.includes(record.name)) roleCopy.push('master');
    if (data.spec.roleGroups.worker.includes(record.name)) roleCopy.push('worker');
    const [curRole,setCurRole] = useState(roleCopy)
    const [visible, setVisible] = React.useState(false);

    const [curHost,setCurHost] = useState(record)

    const [sshAuthenticationType,setSshAuthenticationType] = useState(() => {
        console.log(record)
        if (record.password) {
            return 'password'
        } else if (record.privateKey) {
            return 'privateKey'
        } else if (record.privateKeyPath) {
            return 'privateKeyPath'
        } else {
            return 'password'
        }
    })

    const [labels, setLabels] = useState(() => {
        if (record.labels === undefined || record.labels.length === 0) {
            return [new Map()]
        } else {
            const labels = record.labels
            const labesMap = Object.entries(labels).map(([key, value]) => {
                return new Map([[key, value]])
            })

            return labesMap
        }
    });


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
        setCurHost(recordCopy)
        // 可以不要这行
        setCurRole(roleCopy)
        setVisible(false);
    };
    const roleOptions = [
        {
            value: 'master',
            label:'master'
        },
        {
            value: 'worker',
            label:'worker'
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
            setCurHost(prevState => {
                return ({...prevState, arch: e})
            })
        } else {
            setCurHost(prevState => {
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
            const updateLabels = {}
            for (let i = 0; i < labels.length; i++) {
                updateLabels[Array.from(labels[i].keys())[0]] = Array.from(labels[i].values())[0]
            }
            return updateLabels
        }
        const newHosts = data.spec.hosts.map(host => {
            if (host.name === recordCopy.name) {
                curHost.labels = labelsMap()
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
        else if(curRole[0]==='master') {
            handleChange("spec.roleGroups.master",[...otherMasters,curHost.name])
            handleChange("spec.roleGroups.worker",[...otherWorkers])
        }
        else if(curRole[0]==='worker') {
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
                    <Input name='name' value={curHost.name} onChange={onChangeHandler} placeholder='必填，设置节点名称'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    SSH 地址：
                </Column>
                <Column>
                    <Input name='address' value={curHost.address} onChange={onChangeHandler} placeholder='必填，节点 SSH 链接地址'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    节点 IP：
                </Column>
                <Column>
                    <Input name='internalAddress' value={curHost.internalAddress} onChange={onChangeHandler} placeholder='必填，节点 IP 地址'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    节点角色：
                </Column>
                <Column>
                    <Select multi name='role' value={curRole}  options={roleOptions} onChange={onChangeRoleHandler} placeholder='选择节点角色' ></Select>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    CPU 架构：
                </Column>
                <Column>
                    <RadioGroup buttonWidth={100} wrapClassName="radio-group-button" onChange={onChangeHandler} value={curHost.arch}>
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
                    <Input name='user' value={curHost.user} onChange={onChangeHandler} placeholder='用户应具有 sudo 权限'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    SSH 认证：
                </Column>
                <Column>
                    <RadioGroup buttonWidth={120} wrapClassName="radio-group-button" onChange={onChangeSshAuthenticationTypeHandler} defaultValue={sshAuthenticationType}>
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
                                <InputPassword name='password' value={curHost.password} onChange={onChangeHandler} placeholder='SSH认证方式任选其一，节点 SSH 密码' disabled={sshAuthenticationType === 'password' ? false : true}></InputPassword>
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
                                <InputPassword name='privateKey' value={curHost.privateKey} onChange={onChangeHandler} placeholder='SSH认证方式任选其一，节点 SSH 密钥' disabled={sshAuthenticationType === 'privateKey' ? false : true}></InputPassword>
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
                                <Input name='privateKeyPath' value={curHost.privateKeyPath} onChange={onChangeHandler} placeholder='SSH认证方式任选其一，节点 SSH 密钥文件路径' disabled={sshAuthenticationType === 'privateKeyPath' ? false : true}></Input>
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
                <div style={{ paddingTop: '20px', paddingLeft: '30px', paddingBottom: '20px', paddingRight: '0px' }}>
                    {modalContent}
                </div>
            </Modal>
        </>
    );
}

export default HostEditModal;

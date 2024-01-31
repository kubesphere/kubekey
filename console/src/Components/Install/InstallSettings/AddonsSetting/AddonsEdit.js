import React, {useState} from 'react';
import useInstallFormContext from "../../../../hooks/useInstallFormContext";
import {Column, Columns, Input, InputPassword, RadioButton, RadioGroup, Select} from "@kube-design/components";
import {Modal, Button} from "@kubed/components";
import {MinusSquare, PlusSquare} from "@kubed/icons";

const AddonsEdit = ({record}) => {
    const { data, handleChange } = useInstallFormContext()

    const recordCopy = () => {
        let copy = record
        if (!copy.sources.hasOwnProperty('yaml')) {
            copy.sources.yaml = {path: []}
        }
        if (!copy.sources.hasOwnProperty('chart')) {
            copy.sources.chart = {name: '', repo: '', path: '', valuesFile: ''}
        }
        return copy
    }

    const [curAddon,setCurAddon] = useState(recordCopy)

    const typeCopy = () => {
        let type = 'Helm'

        if (curAddon.sources.hasOwnProperty('yaml')  && curAddon.sources.yaml.path.length > 0) {
            type='Yaml'
        } else if (curAddon.sources.hasOwnProperty('chart') && curAddon.sources.chart.name !== '') {
            type='Helm'
        }
        return type
    }

    const [curType, setCurType] = useState(typeCopy)
    const [visible, setVisible] = useState(false);

    const initPaths = () => {
        let paths = []
        if (curAddon.sources.hasOwnProperty('yaml')  && curAddon.sources.yaml.path.length > 0) {
            paths = curAddon.sources.yaml.path
        }
        return paths
    }

    const [curPaths, setCurPaths] = useState(initPaths);

    const ref = React.createRef();
    const openModal = () => {
        setVisible(true);
    };

    const closeModal = () => {
        setCurAddon(recordCopy)
        setCurType(typeCopy)
        setVisible(false);
    };

    const typeOptions = [
        {
            label: 'Helm',
            value: 'Helm'
        },
        {
            label: 'Yaml',
            value: 'Yaml'
        }
    ]

    const onChangeHandler = e => {
        setCurAddon(prevState => {
            if (e.target.name === 'name' || e.target.name === 'namespace') {
                return {
                    ...prevState,
                    [e.target.name]: e.target.value
                }
            } else if (e.target.name === 'chartName') {
                return {
                    ...prevState,
                    [prevState.sources.chart.name]: e.target.value
                }
            } else if (e.target.name === 'chartRepo') {
                return {
                    ...prevState,
                    [prevState.sources.chart.repo]: e.target.value
                }
            } else if (e.target.name === 'chartPath') {
                return {
                    ...prevState,
                    [prevState.sources.chart.path]: e.target.value
                }
            } else if (e.target.name === 'chartValuesFile') {
                return {
                    ...prevState,
                    [prevState.sources.chart.valuesFile]: e.target.value
                }
            } else if (Number.isInteger(e.target.name)) {
                return {
                    ...prevState,
                    [prevState.sources.yaml.path[e.target.name]]: e.target.value
                }
            }
        })
    }

    const onOKHandler = () => {
        handleChange('addons', [...data.addons, curAddon])
        setCurAddon({
            name: '',
            namespace: '',
            sources: {
                chart: {
                    name: '',
                    repo: '',
                    path: '',
                    valuesFile: '',
                },
                yaml: {
                    path: [],
                }

            }
        });
        setVisible(false);
    }

    const addPath = () => {
        setCurPaths([...curPaths, '']);
    }

    const removePath = (index) => {
        const updetaPath = [...curPaths];
        updetaPath.splice(index, 1);
        setCurPaths(updetaPath);
    }

    const modalContent = (
        <div>
            <Columns>
                <Column className={'is-2'}>
                    组件名称：
                </Column>
                <Column>
                    <Input name='name' value={curAddon.name} onChange={onChangeHandler} placeholder='必填，设置扩展组件名称'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    命名空间：
                </Column>
                <Column>
                    <Input name='namespace' value={curAddon.namespace} onChange={onChangeHandler} placeholder='必填，设置扩展组件命名空间'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    部署类型：
                </Column>
                <Column>
                    <RadioGroup buttonWidth={100} wrapClassName="radio-group-button" onChange={setCurType} defaultValue={curType}>
                        {
                            typeOptions.map(option => <RadioButton key={option.value}  value={option.value}>{option.label}</RadioButton>)
                        }
                    </RadioGroup>
                </Column>
            </Columns>
            {
                curType === 'Yaml' && (
                    <>
                        { curPaths.map((item, index) => (
                            <div key={{index}}>
                                <Columns>
                                    <Column className={'is-2'}>
                                        文件路径：
                                    </Column>
                                    <Column>
                                        { index === 0 ? (
                                            <>
                                                <Input
                                                    name={index}
                                                    onChange={onChangeHandler}
                                                    placeholder='必填，设置扩展组件Yaml路径'
                                                    defaultValue={item}
                                                    value={curAddon.sources.yaml.path[index]}
                                                ></Input>
                                                <PlusSquare onClick={addPath} style={{marginLeft: '10px'}} />
                                            </>
                                        ) : (
                                            <>
                                                <Input
                                                    name={index}
                                                    onChange={onChangeHandler}
                                                    placeholder='必填，设置扩展组件Yaml路径'
                                                    defaultValue={item}
                                                    value={curAddon.sources.yaml.path[index]}
                                                ></Input>
                                                <PlusSquare onClick={addPath} style={{marginLeft: '10px'}} />
                                                <MinusSquare onClick={() => removePath(index)} />
                                            </>
                                        )
                                        }

                                    </Column>
                                </Columns>
                            </div>
                        ))}
                    </>
                )
            }
            {
                curType === 'Helm' && (
                    <>
                        <Columns>
                            <Column className={'is-2'}>
                                Chart 名称：
                            </Column>
                            <Column>
                                <Input
                                    name='chartName'
                                    value={curAddon.sources.chart.name}
                                    onChange={onChangeHandler}
                                    placeholder='必填，设置扩展组件命名空间'
                                ></Input>
                            </Column>
                        </Columns>
                        <Columns>
                            <Column className={'is-2'}>
                                Chart 仓库：
                            </Column>
                            <Column>
                                <Input
                                    name='chartRepo'
                                    value={curAddon.sources.chart.repo}
                                    onChange={onChangeHandler}
                                    placeholder='与 Chart 路径二选一，设置扩展组件 Chart 所在仓库'
                                ></Input>
                            </Column>
                        </Columns>
                        <Columns>
                            <Column className={'is-2'}>
                                Chart 路径：
                            </Column>
                            <Column>
                                <Input
                                    name='chartPath'
                                    value={curAddon.sources.chart.path}
                                    onChange={onChangeHandler}
                                    placeholder='与 Chart 仓库二选一，设置扩展组件 Chart 所在路径'
                                ></Input>
                            </Column>
                        </Columns>
                        <Columns>
                            <Column className={'is-2'}>
                                Values 路径：
                            </Column>
                            <Column>
                                <Input
                                    name='chartValuesFile'
                                    value={curAddon.sources.chart.valuesFile}
                                    onChange={onChangeHandler}
                                    placeholder='可选，设置扩展组件 Values 文件路径'
                                ></Input>
                            </Column>
                        </Columns>
                    </>
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
                title="编辑扩展组件"
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

export default AddonsEdit;

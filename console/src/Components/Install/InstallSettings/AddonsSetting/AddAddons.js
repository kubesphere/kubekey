import React, {useState} from 'react'
import {
    Button,
    Column,
    Columns,
    Form,
    Input,
    InputPassword,
    RadioButton,
    RadioGroup,
    Select,
    MinusCircleOutlined
} from "@kube-design/components";
import {Modal} from "@kubed/components";
import {MinusSquare, PlusSquare} from "@kubed/icons";
import useInstallFormContext from "../../../../hooks/useInstallFormContext";

const AddAddons = () => {
    const { data, handleChange } = useInstallFormContext()
    const [visible, setVisible] = useState(false);
    const [deployType, setDeployType] = useState('Helm');
    const [newAddon, setNewAddon] = useState({
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
    const ref = React.createRef();

    const [paths, setPaths] = useState(['']);

    const addPath = () => {
        setPaths([...paths, '']);
    }

    const removePath = (index) => {
        const updetaPath = [...paths];
        updetaPath.splice(index, 1);
        setPaths(updetaPath);
    }

    const openModal = () => {
        setVisible(true);
    };

    const closeModal = () => {
        setVisible(false);
    };

    const onOKHandler = () => {
        handleChange('addons', [...data.addons, newAddon])
        setNewAddon({
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
        setNewAddon(prevState => {
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

    const modalContent = (
        <div>
            <Columns>
                <Column className={'is-2'}>
                    组件名称：
                </Column>
                <Column>
                    <Input name='name' value={newAddon.name} onChange={onChangeHandler} placeholder='必填，设置扩展组件名称'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    命名空间：
                </Column>
                <Column>
                    <Input name='namespace' value={newAddon.namespace} onChange={onChangeHandler} placeholder='必填，设置扩展组件命名空间'></Input>
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>
                    部署类型：
                </Column>
                <Column>
                    <RadioGroup buttonWidth={100} wrapClassName="radio-group-button" onChange={setDeployType} defaultValue='Helm'>
                        {
                            typeOptions.map(option => <RadioButton key={option.value}  value={option.value}>{option.label}</RadioButton>)
                        }
                    </RadioGroup>
                </Column>
            </Columns>
            {
                deployType === 'Yaml' && (
                    <>
                        { paths.map((item, index) => (
                            <div key={{index}}>
                                <Columns>
                                    <Column className={'is-2'}>
                                        文件路径：
                                    </Column>
                                    <Column>
                                        { index === 0 ? (
                                            <>
                                                <Input name={index} value={newAddon.sources.yaml.path[index]} onChange={onChangeHandler} placeholder='必填，设置扩展组件Yaml路径'></Input>
                                                <PlusSquare onClick={addPath} style={{marginLeft: '10px'}} />
                                            </>
                                        ) : (
                                            <>
                                                <Input name={index} value={newAddon.sources.yaml.path[index]} onChange={onChangeHandler} placeholder='必填，设置扩展组件Yaml路径'></Input>
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
                deployType === 'Helm' && (
                    <>
                        <Columns>
                            <Column className={'is-2'}>
                                Chart 名称：
                            </Column>
                            <Column>
                                <Input name='chartName' value={newAddon.sources.chart.name} onChange={onChangeHandler} placeholder='必填，设置扩展组件命名空间'></Input>
                            </Column>
                        </Columns>
                        <Columns>
                            <Column className={'is-2'}>
                                Chart 仓库：
                            </Column>
                            <Column>
                                <Input name='chartRepo' value={newAddon.sources.chart.repo} onChange={onChangeHandler} placeholder='与 Chart 路径二选一，设置扩展组件 Chart 所在仓库'></Input>
                            </Column>
                        </Columns>
                        <Columns>
                            <Column className={'is-2'}>
                                Chart 路径：
                            </Column>
                            <Column>
                                <Input name='chartPath' value={newAddon.sources.chart.path} onChange={onChangeHandler} placeholder='与 Chart 仓库二选一，设置扩展组件 Chart 所在路径'></Input>
                            </Column>
                        </Columns>
                        <Columns>
                            <Column className={'is-2'}>
                                Values 路径：
                            </Column>
                            <Column>
                                <Input name='chartValuesFile' value={newAddon.sources.chart.valuesFile} onChange={onChangeHandler} placeholder='可选，设置扩展组件 Values 文件路径'></Input>
                            </Column>
                        </Columns>
                    </>
                )
            }
        </div>
    )
    return (
        <>
            <Button onClick={openModal}>添加扩展组件</Button>
            <Modal
                ref={ref}
                visible={visible}
                title="添加扩展组件"
                onCancel={closeModal}
                onOk={onOKHandler}
            >
                <div style={{ paddingTop: '20px', paddingLeft: '30px', paddingBottom: '20px', paddingRight: '30px' }}>
                    {modalContent}
                </div>
            </Modal>
        </>
    );
}

export default AddAddons

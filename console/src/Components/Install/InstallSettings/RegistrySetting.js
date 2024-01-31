import React ,{useState} from 'react';
import useInstallFormContext from "../../../hooks/useInstallFormContext";
import {Column, Input, Columns, TextArea, Toggle, RadioGroup, RadioButton} from "@kube-design/components";
const RegistrySetting = () => {
    const { data, handleChange } = useInstallFormContext()
    const changeInsecureRegistriesHandler = e => {
        handleChange('spec.registry.insecureRegistries',e.split('\n'))
    }
    const changeRegistryMirrorsHandler = e => {
        handleChange('spec.registry.registryMirrors',e.split('\n'))
    }

    const changePrivateRegistryUrlHandler= e => {
        handleChange('spec.registry.privateRegistry',e.target.value)
    }

    const [initRegistry, setInitRegistry] = useState(false)

    const changeInitRegitsryHandler = e => {
        setInitRegistry(e)
        if(!e) {
            handleChange('spec.registry.privateRegistry','')
        }
    }

    const registryTypeOptions = [
        {
            label: 'Registry',
            value: 'registry'
        },
        {
            label: 'Harbor',
            value: 'harbor'
        }
    ]

    const changeRegistryTypeHandler = e => {
        handleChange('spec.registry.type',e)
    }

    return (
        <div>
            <Columns >
                <Column className={'is-2'}>初始化私有镜像仓库:</Column>
                <Column>
                    <Toggle checked={initRegistry} onChange={changeInitRegitsryHandler} onText="是" offText="否" />
                </Column>
            </Columns>
            {
                initRegistry && (
                    <Columns>
                        <Column className={'is-2'}>
                            初始化仓库类型：
                        </Column>
                        <Column>
                            <RadioGroup buttonWidth={100} wrapClassName="radio-group-button" onChange={changeRegistryTypeHandler} defaultValue={data.spec.registry.type === undefined ? "registry": data.spec.registry.type}>
                                {registryTypeOptions.map(option => <RadioButton key={option.value} value={option.value}>
                                    {option.label}
                                </RadioButton>)}
                            </RadioGroup>
                        </Column>
                    </Columns>
                )
            }

            <Columns >
                <Column className={'is-2'}>私有镜像仓库地址:</Column>
                <Column>
                    <Input placeholder={"请输入私有镜像仓库地址，留空表示使用公网仓库"} style={{width:'100%'}} value={data.spec.registry.privateRegistry} onChange={changePrivateRegistryUrlHandler} />
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>Insecure Registry 地址:</Column>
                <Column >
                    <TextArea style={{width:'100%'}} onChange={changeInsecureRegistriesHandler} value={data.spec.registry.insecureRegistries.join('\n')} autoResize maxHeight={200} placeholder="请输入Insecure Registry 地址，每行一个，该参数用于为容器运行时设置 Insecure Registry" />
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>Registry Mirrors 地址:</Column>
                <Column >
                    <TextArea style={{width:'100%'}}  placeholder={"请输入Registry Mirrors 地址，每行一个，该参数用于为容器运行时设置 Registry Mirrors"} onChange={changeRegistryMirrorsHandler} value={data.spec.registry.registryMirrors.join('\n')} autoResize maxHeight={200} />
                </Column>
            </Columns>

        </div>
    )
};

export default RegistrySetting;

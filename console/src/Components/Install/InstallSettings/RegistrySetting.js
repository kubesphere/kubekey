import React from 'react';
import useInstallFormContext from "../../../hooks/useInstallFormContext";
import {Column, Input, Columns, TextArea} from "@kube-design/components";
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
    return (
        <div>
            <Columns >
                <Column className={'is-2'}>私有镜像仓库Url:</Column>
                <Column>
                    <Input placeholder={"请输入私有镜像仓库Url，留空代表不使用"} style={{width:'100%'}} value={data.spec.registry.privateRegistry} onChange={changePrivateRegistryUrlHandler} />
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>非安全仓库:</Column>
                <Column >
                    <TextArea style={{width:'100%'}} onChange={changeInsecureRegistriesHandler} value={data.spec.registry.insecureRegistries.join('\n')} autoResize maxHeight={200} placeholder="请输入非安全仓库，每行一个，留空代表不使用" />
                </Column>
            </Columns>
            <Columns>
                <Column className={'is-2'}>仓库镜像Url:</Column>
                <Column >
                    <TextArea style={{width:'100%'}}  placeholder={"请输入镜像仓库Url，每行一个,留空代表不使用"} onChange={changeRegistryMirrorsHandler} value={data.spec.registry.registryMirrors.join('\n')} autoResize maxHeight={200} />
                </Column>
            </Columns>

        </div>
    )
};

export default RegistrySetting;

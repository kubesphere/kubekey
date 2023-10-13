import React from 'react';
import useInstallFormContext from "../../../hooks/useInstallFormContext";
import {Column, Columns, Input, Toggle} from "@kube-design/components";

const StorageSetting = () => {

    const { data, handleChange,setKsEnable, setKsVersion } = useInstallFormContext()
    const changeEnableLocalStorageHandler = e =>  {
        if(e) {
            handleChange('spec.storage',{
                openebs: {
                    basePath: '/var/openebs/local',
                },
            })
        } else {
            setKsEnable(false)
            setKsVersion('')
            handleChange('spec.storage',{})

        }
    }

    const changeLocalStoragePathHandler = (e)=> {
        handleChange('spec.storage.openebs.basePath', e.target.value)
    }
    return (
        // TODO 待处理
        <div>
            <Columns >
                <Column className={'is-2'}>开启openebs本地存储:</Column>
                <Column>
                    <Toggle checked={data.spec.storage
                        && data.spec.storage.openebs
                        && data.spec.storage.openebs.basePath
                        && data.spec.storage.openebs.basePath!==''}
                            onChange={changeEnableLocalStorageHandler} onText="开启" offText="关闭" />
                </Column>

            </Columns>
            <Columns >
                <Column className={'is-2'}>openebs路径:</Column>
                <Column>
                    <Input value={(data.spec.storage
                                && data.spec.storage.openebs
                                && data.spec.storage.openebs.basePath
                                && data.spec.storage.openebs.basePath!=='')?data.spec.storage.openebs.basePath:''}
                           onChange={changeLocalStoragePathHandler}
                           disabled={!(data.spec.storage
                                && data.spec.storage.openebs
                                && data.spec.storage.openebs.basePath
                                && data.spec.storage.openebs.basePath!=='')}></Input>
                </Column>

            </Columns>
        </div>
    )
};

export default StorageSetting;

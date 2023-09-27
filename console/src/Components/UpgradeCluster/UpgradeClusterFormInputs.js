import React from 'react';
import useUpgradeClusterFormContext from "../../hooks/useUpgradeClusterFormContext";
import UpgradeRegistrySetting from "./UpgradeSettings/UpgradeRegistrySetting";
import UpgradeKubesphereSetting from "./UpgradeSettings/UpgradeKubesphereSetting";
import UpgradeClusterSetting from "./UpgradeSettings/UpgradeClusterSetting";
import ConfirmUpgradeSetting from "./UpgradeSettings/ConfirmUpgradeSetting";
import UpgradeStorageSetting from "./UpgradeSettings/UpgradeStorageSetting";

const UpgradeClusterFormInputs = () => {
    const { page } = useUpgradeClusterFormContext()
    const display = {
        0: <UpgradeClusterSetting/>,
        1: <UpgradeRegistrySetting/>,
        2: <UpgradeKubesphereSetting/>,
        3: <UpgradeStorageSetting/>,
        4: <ConfirmUpgradeSetting/>
    }

    return (
        <div>
            {display[page]}
        </div>
    )
};

export default UpgradeClusterFormInputs;

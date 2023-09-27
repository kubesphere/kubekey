import useInstallFormContext from "../../hooks/useInstallFormContext"
import EtcdSetting from "./InstallSettings/ETCDSetting";
import ClusterSetting from "./InstallSettings/ClusterSetting";
import HostSetting from "./InstallSettings/HostSetting/HostSetting";
import NetworkSetting from "./InstallSettings/NetworkSetting";
import StorageSetting from "./InstallSettings/StorageSetting";
import RegistrySetting from "./InstallSettings/RegistrySetting";
import KubesphereSetting from "./InstallSettings/KubesphereSetting";
import ConfirmInstallSetting from "./InstallSettings/ConfirmInstallSetting";

const InstallFormInputs = () => {

    const { page } = useInstallFormContext()

    const display = {
        0: <HostSetting/>,
        1: <EtcdSetting/>,
        2: <ClusterSetting/>,
        3: <NetworkSetting/>,
        4: <StorageSetting/>,
        5: <RegistrySetting/>,
        6: <KubesphereSetting/>,
        7: <ConfirmInstallSetting/>
    }

    return (
        <div>
            {display[page]}
        </div>
    )
}
export default InstallFormInputs

import React from 'react';
import HostTable from "./HostTable";
import HostAddModal from "../../../Modal/HostAddModal";

const HostSetting = () => {

    return (
        <div>
        <HostAddModal></HostAddModal>
        <HostTable></HostTable>
    </div>
    )
};

export default HostSetting;

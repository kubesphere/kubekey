import React from 'react';
import {Columns, Form, Input, Button, TextArea, Column} from "@kube-design/components";
import AddAddons from "./AddAddons";
import AddonsTable from "./AddonsTable";


const AddonsSetting = () => {
  return (
    <div>
        <AddAddons></AddAddons>
        <AddonsTable></AddonsTable>
    </div>
  );
}

export default AddonsSetting;

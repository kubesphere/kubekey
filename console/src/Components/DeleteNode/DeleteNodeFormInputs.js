import React from 'react';
import useDeleteNodeFormContext from "../../hooks/useDeleteNodeFormContext";
import SelectNodeSetting from "./DeleteNodeSettings/SelectNodeSetting/SelectNodeSetting";
import ConfirmDeleteNodeSetting from "./DeleteNodeSettings/ConfirmDeleteNodeSetting";

const DeleteNodeFormInputs = () => {
    const { page } = useDeleteNodeFormContext()

    const display = {
        0: <SelectNodeSetting />,
        1: <ConfirmDeleteNodeSetting />
    }

    return (
        <div>
            {display[page]}
        </div>
    )
};

export default DeleteNodeFormInputs;

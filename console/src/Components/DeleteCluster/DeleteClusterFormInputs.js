import React from 'react';
import useDeleteClusterContext from "../../hooks/useDeleteClusterContext";
import ConfirmDeleteClusterSetting from "./DeleteClusterSettings/ConfirmDeleteClusterSetting";
import DeleteCriSetting from "./DeleteClusterSettings/DeleteCRISetting";

const DeleteClusterFormInputs = () => {
    const { page } = useDeleteClusterContext()

    const display = {
        0: <DeleteCriSetting/>,
        1: <ConfirmDeleteClusterSetting/>,
    }

    return (
        <div>
            {display[page]}
        </div>
    )
};

export default DeleteClusterFormInputs;

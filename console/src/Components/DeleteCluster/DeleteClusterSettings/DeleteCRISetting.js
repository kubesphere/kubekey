import React from 'react';
import useDeleteClusterContext from "../../../hooks/useDeleteClusterContext";
import {Column, Columns, Toggle} from "@kube-design/components";

const DeleteCriSetting = () => {
    const { deleteCRI,setDeleteCRI } = useDeleteClusterContext()
    const deleteCRIHandler = () => {
        setDeleteCRI(prev=>!prev)
    }
    return (
        <>
            <Columns>
                <Column>
                    是否删除CRI:
                </Column>
                <Column>
                    <Toggle checked={deleteCRI} onChange={deleteCRIHandler} onText="是" offText="否"></Toggle>
                </Column>
            </Columns>
        </>
    );
};

export default DeleteCriSetting;

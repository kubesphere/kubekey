import React from 'react';
import {Button, Modal} from "@kubed/components";

import {Column, Columns} from "@kube-design/components";
import useInstallFormContext from "../../../../hooks/useInstallFormContext";


const AddonsDeleteConfirmModal = ({record}) => {

    const { data, handleChange } = useInstallFormContext()

    const [visible, setVisible] = React.useState(false);


    const ref = React.createRef();
    const openModal = () => {
        setVisible(true);
    };

    const closeModal = () => {
        setVisible(false);
    };

    const onOKHandler = () => {
        const newAddons = data.spec.addons.filter(addon => addon.name !== record.name);
        handleChange('spec.addons',newAddons)
        setVisible(false);
    }

    const textStyle={
        fontSize:"20px",
        height: '30px',
        margin: 0, /* 清除默认的外边距 */
        display: 'flex',
        alignItems: 'center'
    }

    return (
        <div>
            <Button variant="link" onClick={openModal}>删除</Button>
            <Modal
                ref={ref}
                visible={visible}
                title="删除扩展组件"
                onCancel={closeModal}
                onOk={onOKHandler}
            >
                <Columns>
                    <Column className='is-1'></Column>
                    <Column style={{display:`flex`, alignItems: 'center' }}>
                        <p style={textStyle}>是否确定删除？</p>
                    </Column>
                </Columns>
            </Modal>
        </div>
    );
};

export default AddonsDeleteConfirmModal;

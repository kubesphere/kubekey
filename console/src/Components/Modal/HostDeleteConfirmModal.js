import React from 'react';
import {Button, Modal} from "@kubed/components";
import useInstallFormContext from "../../hooks/useInstallFormContext";
import {Column, Columns} from "@kube-design/components";


const HostDeleteConfirmModal = ({record}) => {

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
        const newHosts = data.spec.hosts.filter(host => host.name !== record.name);
        handleChange('spec.hosts',newHosts)
        if(data.spec.roleGroups.master.includes(record.name)){
            const newMasters = data.spec.roleGroups.master.filter(name=>name!==record.name)
            handleChange('spec.roleGroups.master',newMasters)
        }
        if(data.spec.roleGroups.worker.includes(record.name)){
            const newWorkers = data.spec.roleGroups.worker.filter(name => name!==record.name)
            handleChange('spec.roleGroups.worker',newWorkers)
        }
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
                title="删除节点"
                onCancel={closeModal}
                onOk={onOKHandler}
            >
                <Columns>
                    <Column className='is-1'></Column>
                    <Column style={{display:`flex`, alignItems: 'center' }}>
                        <p style={textStyle}>确定删除吗？</p>
                    </Column>
                </Columns>
            </Modal>
        </div>
    );
};

export default HostDeleteConfirmModal;

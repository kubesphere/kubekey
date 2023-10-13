import React from 'react';
import {Button, Modal} from "@kubed/components";
import {Column, Columns} from "@kube-design/components";

const AlertModal = () => {
    const [visible, setVisible] = React.useState(false);


    const ref = React.createRef();
    const openModal = () => {
        setVisible(true);
    };

    const closeModal = () => {
        setVisible(false);
    };
    const onOKHandler = () => {
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
            <Button variant="link" onClick={openModal}></Button>
            <Modal
                ref={ref}
                visible={visible}
                title="开始安装集群"
                onCancel={closeModal}
                onOk={onOKHandler}
            >
                <Columns>
                    <Column className='is-1'></Column>
                    <Column style={{display:`flex`, alignItems: 'center' }}>
                        <p style={textStyle}>集群安装已开始，关闭该提示后可查看实时日志，期间请勿进行其他操作！</p>
                    </Column>
                </Columns>
            </Modal>
        </div>
    );
};

export default AlertModal;

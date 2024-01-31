import InstallFormInputs from './InstallFormInputs'
import useInstallFormContext from "../../hooks/useInstallFormContext"
import {Button, Columns, Column} from "@kube-design/components";
import React from "react";
import {Modal} from "@kubed/components";

const InstallForm = () => {
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
        fontSize:"15px",
        minHeight: '50px',
        margin: 0, /* 清除默认的外边距 */
        display: 'flex',
        alignItems: 'center'
    }

    const {
        page,
        setPage,
        title,
        canSubmit,
        disablePrev,
        disableNext,
        installHandler
    } = useInstallFormContext()

    const handlePrev = () => {
        setPage(prev => {
            return prev <= 0? 0 : +prev-1
        })
    }

    const handleNext = () => {
        setPage(prev => {
            return +prev+1
        })
    }

    return (
        <>
            <Columns>
                <Column className='is-10'>
                    <h3>{title[page]}</h3>
                </Column>
            </Columns>
            <Columns>
                <Column>
                    <InstallFormInputs />
                </Column>
            </Columns>
            <Columns>
                <Column className='is-8'/>
                <Column className='is-2'>
                    <Columns>
                        <Column className='is-5'/>
                        <Column>
                            {page > 0  && <Button onClick={handlePrev} disabled={disablePrev}>上一步</Button>}
                        </Column>
                    </Columns>
                </Column>
                <Column className='is-1'>
                    {page !== Object.keys(title).length - 1 && <Button onClick={handleNext} disabled={disableNext}>下一步</Button>}
                    {page === Object.keys(title).length - 1 && <Button disabled={!canSubmit} onClick={()=>{installHandler();openModal();}}>安装</Button>}
                    <Modal
                        ref={ref}
                        visible={visible}
                        title="开始安装集群"
                        onCancel={closeModal}
                        onOk={onOKHandler}
                    >
                        <Columns>
                            <Column style={{display:`flex`, alignItems: 'center' }}>
                                <p style={textStyle}>集群安装已开始，关闭该提示后可查看实时日志，期间请勿进行其他操作！</p>
                            </Column>
                        </Columns>
                    </Modal>
                </Column>
            </Columns>

        </>
    )
}
export default InstallForm

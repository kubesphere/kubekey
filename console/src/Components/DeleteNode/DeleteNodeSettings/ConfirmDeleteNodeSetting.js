import React, {useRef} from 'react';
import useDeleteNodeFormContext from "../../../hooks/useDeleteNodeFormContext";

const ConfirmDeleteNodeSetting = () => {
    const {logs} = useDeleteNodeFormContext();
    const logContainerRef = useRef(null);

    return (
        <div>
            <div ref={logContainerRef} style={{
                backgroundColor: '#1e1e1e',
                color: '#ffffff',
                padding: '10px',
                borderRadius: '5px',
                maxHeight: '500px',
                maxWidth: '850px',
                overflowY: 'scroll',
                fontFamily: 'Consolas, "Courier New", monospace',
                fontSize: '14px',
                lineHeight: '1.5'
            }}>
                {logs.map((log, index) => (
                    <div key={index} style={{ whiteSpace: 'pre-wrap' }}>
                        {log}
                    </div>
                ))}
            </div>
        </div>
    );
};

export default ConfirmDeleteNodeSetting;

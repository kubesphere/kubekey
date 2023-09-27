import React, {createContext, useEffect, useRef, useState} from 'react';
import useGlobalContext from "../hooks/useGlobalContext";

const UpgradeClusterFormContext = createContext({})
export const UpgradeClusterFormProvider = ({children})=> {
    const [originalClusterVersion,setOriginalClusterVersion] = useState('')
    const [canToHome,setCanToHome] = useState(true)
    const {backendIP} = useGlobalContext();
    const [curCluster,setCurCluster] = useState({});
    const [buttonDisabled,setButtonDisabled] = useState(false)
    const title = {
        0:'集群设置',
        1:'镜像仓库设置',
        2:'Kubesphere设置',
        3:'存储设置',
        4:'确认升级'
    }

    const jsyaml = require('js-yaml');
    const [logs, setLogs] = useState([]);
    const socketRef = useRef(null);
    const logsRef = useRef([]);
    useEffect(() => {
        function hashChangeListener() {
            if (socketRef.current) {
                socketRef.current.close();
            }
        }

        window.addEventListener('hashchange', hashChangeListener);

        return () => {
            window.removeEventListener('hashchange', hashChangeListener);
        };
    }, []);

    const upgradeHandler = () => {
        socketRef.current = new WebSocket(`ws://${backendIP}:8082/upgradeCluster?clusterName=${curCluster.metadata.name}&ksVersion=${ksVersion}`);
        socketRef.current.addEventListener('open', () => {
            setLogs([])
            logsRef.current.push('升级集群开始，请勿进行其他操作！');
            setLogs([...logsRef.current]);
            console.log('WebSocket is open now.');
            setButtonDisabled(true)
            setCanToHome(false)
            socketRef.current.send(jsyaml.dump(curCluster));
        });

        socketRef.current.addEventListener('message', (event) => {
            if(event.data==='升级集群成功') {
                setCanToHome(true)
                if (socketRef.current) {
                    socketRef.current.close();
                }
            }
            if(event.data==='升级集群失败') {
                setButtonDisabled(false)
                setCanToHome(true)
                if (socketRef.current) {
                    socketRef.current.close();
                }
            }
            logsRef.current.push(event.data);
            setLogs([...logsRef.current]);
        });
        socketRef.current.addEventListener('close', () => {
            console.log('WebSocket is closed now.');
            // 在这里处理WebSocket关闭事件
        });

        socketRef.current.addEventListener('error', (event) => {
            logsRef.current.push('与后端建立连接失败，请检查kk console状态');
            setLogs([...logsRef.current])
            console.error('WebSocket error: ', event);
            // 在这里处理WebSocket错误事件
        });
    }


    const [page,setPage] = useState(0)
    const [ksVersion,setKsVersion] = useState('v3.4.0')
    const [ksEnable,setKsEnable] = useState(true)
    const canSubmit = !buttonDisabled

    const canNextPage0To1 = (curCluster.metadata &&curCluster.metadata.name !== '')
        && curCluster.spec.kubernetes.version !== ''
    && curCluster.spec.kubernetes.containerManager !== ''
    //

    const disablePrev = page === 0
    || buttonDisabled

    const handleChange = (fieldName, newValue) => {
        setCurCluster(prevState => {
            if(fieldName==='') {
                return {...prevState, newValue}
            } else {
                const updatedData = { ...prevState };
                // 使用字段名拆分成多级属性
                const fieldNames = fieldName.split('.');
                let currentField = updatedData;
                // 遍历字段名的每一级
                for (let i = 0; i < fieldNames.length; i++) {
                    const name = fieldNames[i];

                    // 如果是最后一级属性，直接更新其值
                    if (i === fieldNames.length - 1) {
                        currentField[name] = newValue;
                    } else {
                        // 如果不是最后一级属性，确保属性存在并进入下一级
                        if (!currentField[name]) {
                            currentField[name] = {};
                        }
                        currentField = currentField[name];
                    }
                }
                return updatedData
            }
        });
    };

    const disableNext =
        (page === Object.keys(title).length - 1)
        || (page === 0 && !canNextPage0To1)

    return (
        <UpgradeClusterFormContext.Provider value={{ ksVersion,setKsVersion,ksEnable,setKsEnable,buttonDisabled,setButtonDisabled, title, page, setPage,disableNext,
            disablePrev, handleChange, canSubmit,curCluster,setCurCluster,
            upgradeHandler,logs,canToHome,originalClusterVersion,setOriginalClusterVersion}}>
            {children}
        </UpgradeClusterFormContext.Provider>
    );
};
export default UpgradeClusterFormContext;

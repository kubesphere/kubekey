import React, {createContext, useEffect, useRef, useState} from 'react';
import useGlobalContext from "../hooks/useGlobalContext";

const AddNodeFormContext = createContext({})
export const AddNodeFormProvider = ({children}) => {
    const [canToHome,setCanToHome] = useState(true)
    const {backendIP} = useGlobalContext();
    const [curCluster,setCurCluster] = useState({});
    const [buttonDisabled,setButtonDisabled] = useState(false)
    const title = {
        0:'新增节点',
        1:'ETCD设置',
        2:'确认新增',
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
    const [curSelectedNodeName,setCurSelectedNodeName] = useState([])

    const [page,setPage] = useState(0)
    const handleChange = (fieldName, newValue) => {
    };

    const addHandler = () => {
        socketRef.current = new WebSocket(`ws://${backendIP}:8082/addNode?clusterName=${curCluster.metadata.name}`);
        socketRef.current.addEventListener('open', () => {
            setLogs([])
            logsRef.current.push('添加节点开始，请勿进行其他操作！');
            setLogs([...logsRef.current]);
            console.log('WebSocket is open now.');
            setButtonDisabled(true)
            setCanToHome(false)
            socketRef.current.send(jsyaml.dump(curCluster));
        });

        socketRef.current.addEventListener('message', (event) => {
            if(event.data==='添加节点成功') {
                // setButtonDisabled(false)
                setCanToHome(true)
                if (socketRef.current) {
                    socketRef.current.close();
                }
            }
            if(event.data==='添加节点失败') {
                setCanToHome(true)
                setButtonDisabled(false)
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

    const canSubmit = !buttonDisabled

    const canNextPage0To1 = (curCluster.spec && curCluster.spec.hosts.length>0)
    //
    const canNextPage1To2 = (curCluster.spec && curCluster.spec.roleGroups.etcd.length>0)

    const disablePrev = page === 0 || buttonDisabled
    // || disableButton

    // const allHostHaveRole =

    const disableNext =
        (page === Object.keys(title).length - 1)
        || (page === 0 && !canNextPage0To1)
        || (page === 1 && !canNextPage1To2)

    return (
        <AddNodeFormContext.Provider value={{ title, page, setPage,disableNext,
            disablePrev, handleChange, canSubmit,
            curSelectedNodeName,setCurSelectedNodeName,curCluster,setCurCluster,
            logs,addHandler,buttonDisabled,setButtonDisabled,canToHome}}>
            {children}
        </AddNodeFormContext.Provider>
    );
};

export default AddNodeFormContext;

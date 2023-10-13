import React, {createContext, useEffect, useRef, useState} from 'react';
import useGlobalContext from "../hooks/useGlobalContext";

const DeleteNodeFormContext = createContext({})

export const DeleteNodeFormProvider = ({children}) => {
    const [canToHome,setCanToHome] = useState(true)
    const {backendIP} = useGlobalContext();
    const [curCluster,setCurCluster] = useState({});
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


    const deleteHandler = () => {
        socketRef.current = new WebSocket(`ws://${backendIP}:8082/deleteNode?clusterName=${curCluster.metadata.name}&nodeName=${curSelectedNodeName}`);
        socketRef.current.addEventListener('open', () => {
            setLogs([])
            logsRef.current.push('删除节点开始，请勿进行其他操作！');
            setLogs([...logsRef.current]);
            console.log('WebSocket is open now.');
            setButtonDisabled(true)
            setCanToHome(false)
            socketRef.current.send(jsyaml.dump(curCluster));
        });

        socketRef.current.addEventListener('message', (event) => {
            if(event.data==='删除节点成功') {
                setCanToHome(true)
                if (socketRef.current) {
                    socketRef.current.close();
                }
            }
            if(event.data==='删除节点失败') {
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

    const [buttonDisabled,setButtonDisabled] = useState(false)
    const title = {
        0:'选择节点',
        1:'确认删除'
    }
    const [curSelectedNodeName,setCurSelectedNodeName] = useState('')

    const [page,setPage] = useState(0)
    const handleChange = (fieldName, newValue) => {
    };

    const canSubmit = !buttonDisabled

    const canNextPage0To1 = curSelectedNodeName!==''


    const disablePrev = page === 0
        || buttonDisabled

    const disableNext =
        (page === Object.keys(title).length - 1)
        || (page === 0 && !canNextPage0To1)

    return (
        <DeleteNodeFormContext.Provider value={{ buttonDisabled,setButtonDisabled, title, page, setPage,disableNext,
            disablePrev, handleChange, canSubmit,curSelectedNodeName,setCurSelectedNodeName,
            curCluster,setCurCluster,deleteHandler,logs,canToHome}}>
            {children}
        </DeleteNodeFormContext.Provider>
    );
};

export default DeleteNodeFormContext;

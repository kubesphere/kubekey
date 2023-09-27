import React, {createContext, useEffect, useRef, useState} from 'react';
import useGlobalContext from "../hooks/useGlobalContext";

const DeleteClusterContext = createContext({})

export const DeleteClusterProvider = ({children}) => {
    const [canToHome,setCanToHome] = useState(true)
    const [deleteCRI,setDeleteCRI] = useState(false)
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
        socketRef.current = new WebSocket(`ws://${backendIP}:8082/deleteCluster?clusterName=${curCluster.metadata.name}&deleteCRI=${deleteCRI?'yes':'no'}`);
        socketRef.current.addEventListener('open', () => {
            setLogs([])
            logsRef.current.push('删除集群开始，请勿进行其他操作！');
            setLogs([...logsRef.current]);
            console.log('WebSocket is open now.');
            setButtonDisabled(true)
            setCanToHome(false)
            socketRef.current.send(jsyaml.dump(curCluster));
        });

        socketRef.current.addEventListener('message', (event) => {
            if(event.data==='删除集群成功') {
                // setButtonDisabled(false)
                setCanToHome(true)
                if (socketRef.current) {
                    socketRef.current.close();
                }
            }
            if(event.data==='删除集群失败') {
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
        0: 'CRI设置',
        1: '确认删除',
    }
    const [page,setPage] = useState(0)

    const canSubmit = !buttonDisabled

    const disablePrev = page === 0
        || buttonDisabled

    const disableNext = page === Object.keys(title).length - 1
    return (
        <DeleteClusterContext.Provider value={{ buttonDisabled,setButtonDisabled, title, page, setPage,
            canSubmit,deleteCRI,setDeleteCRI,
            curCluster,setCurCluster,deleteHandler,logs,canToHome,
            disablePrev,disableNext}}>
            {children}
        </DeleteClusterContext.Provider>
    );
}
export default DeleteClusterContext;

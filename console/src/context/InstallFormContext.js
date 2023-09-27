import {createContext, useEffect, useRef, useState} from "react";
import useGlobalContext from "../hooks/useGlobalContext";

const InstallFormContext = createContext({})

export const InstallFormProvider = ({children}) => {

    const [buttonDisabled,setButtonDisabled] = useState(false)
    const [canToHome,setCanToHome] = useState(true)
    const [ksEnable,setKsEnable] = useState(false)
    const [ksVersion,setKsVersion] = useState('')
    const [KubekeyNamespace,setKubekeyNamespace] = useState('kubekey-system')
    const title = {
        0:'主机设置',
        1:'ETCD设置',
        2:'集群设置',
        3:'网络设置',
        4:'存储设置',
        5:'镜像仓库设置',
        6:'KubeSphere 设置',
        7:'确认安装'
    }
    const [page,setPage] = useState(0)
    const {backendIP} = useGlobalContext();
    const [data, setData] = useState({
        apiVersion: 'kubekey.kubesphere.io/v1alpha2',
        kind: 'Cluster',
        metadata: {
            name: 'cluster.local',
            labels: {
                "type.kubekey.kubesphere.io/backend": "true" // 添加你的标签值
            }
        },
        spec:{
            hosts:[{
                name : 'node1',
                address : '192.168.6.1',
                internalAddress : '192.168.6.1',
                user: 'root',
                password : '123456',
                privateKeyPath: '/var/root/.ssh/id_rsa'
            },
                {
                name : 'node2',
                address : '192.168.6.2',
                internalAddress : '192.168.6.2',
                user : 'root',
                password : '123456',
                privateKeyPath: '/var/root/.ssh/id_rsa'
                },
                {
                name : 'node3',
                address : '192.168.6.3',
                internalAddress : '192.168.6.3',
                user : 'root',
                password : '123456',
                privateKeyPath: '/var/root/.ssh/id_rsa'
                }],
            roleGroups: {
                etcd: [],
                master: ['node1','node2'],
                worker: ['node2','node3'],
            },
            controlPlaneEndpoint: {
                // internalLoadbalancer: 'haproxy',
                // externalDNS: false,
                domain: 'lb.kubesphere.local',
                address: '',
                port: 6443,
            },
            // system: {
            //     ntpServers: [
            //         'time1.cloud.tencent.com',
            //         'ntp.aliyun.com',
            //     ],
            //     timezone: 'Asia/Shanghai',
            //     rpms: ['nfs-utils'],
            //     debs: ['nfs-common'],
            // },
            kubernetes: {
                version: 'v1.21.5',
                // apiserverCertExtraSans: ['192.168.8.8', 'lb.kubespheredev.local'],
                containerManager: 'docker',
                clusterName: 'cluster.local',
                autoRenewCerts: true,
                masqueradeAll: false,
                maxPods: 110,
                // podPidsLimit: 10000,
                nodeCidrMaskSize: 24,
                proxyMode: 'ipvs',
                // featureGates: {
                //     CSIStorageCapacity: true,
                //     ExpandCSIVolumes: true,
                //     RotateKubeletServerCertificate: true,
                //     TTLAfterFinished: true,
                // },
                // kubeProxyConfiguration: {
                //     ipvs: {
                //         excludeCIDRs: ['172.16.0.2/24'],
                //     },
                // },
            },
            etcd: {
                type: 'kubekey',
                // dataDir: '/var/lib/etcd',
                // heartbeatInterval: 250,
                // electionTimeout: 5000,
                // snapshotCount: 10000,
                // autoCompactionRetention: 8,
                // metrics: 'basic',
                // quotaBackendBytes: 2147483648,
                // maxRequestBytes: 1572864,
                // maxSnapshots: 5,
                // maxWals: 5,
                // logLevel: 'info',
            },
            network: {
                plugin: 'calico',
                // calico: {
                //     ipipMode: 'Always',
                //     vxlanMode: 'Never',
                //     vethMTU: 0,
                // },
                kubePodsCIDR: '10.233.64.0/18',
                kubeServiceCIDR: '10.233.0.0/18',
            },
            // 不一样
            storage: {
                openebs: {
                    basePath: '/var/openebs/local',
                },
            },
            registry: {
                registryMirrors: [],
                insecureRegistries: [],
                privateRegistry: '',
                namespaceOverride: '',
                // TODO 添加对应输入框
                // auths: {
                //     'dockerhub.kubekey.local': {
                //         username: 'xxx',
                //         password: '***',
                //         skipTLSVerify: false,
                //         plainHTTP: false,
                //         certsPath: '/etc/docker/certs.d/dockerhub.kubekey.local',
                //     },
                // },
            },
            addons: [],
        },
    })
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
    const installHandler = () => {
        socketRef.current = new WebSocket(`ws://${backendIP}:8082/createCluster?clusterName=${data.metadata.name}&ksVersion=${ksVersion}&KubekeyNamespace=${KubekeyNamespace}`);
        socketRef.current.addEventListener('open', () => {
            setLogs([])
            logsRef.current.push('安装开始，请勿进行其他操作！');
            setLogs([...logsRef.current]);
            console.log('WebSocket is open now.');
            setButtonDisabled(true)
            setCanToHome(false)
            socketRef.current.send(jsyaml.dump(data));
        });

        socketRef.current.addEventListener('message', (event) => {
            if(event.data==='安装集群成功') {
                // setButtonDisabled(false)
                setCanToHome(true)
                if (socketRef.current) {
                    socketRef.current.close();
                }
            }
            if(event.data==='安装集群失败') {
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
        });
    }

    const handleChange = (fieldName, newValue) => {
        setData(prevState => {
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

    const canSubmit = !buttonDisabled

    const canNextPage0To1 = data.spec.hosts.length>0 && data.spec.roleGroups.master.length > 0
    //
    const canNextPage1To2 = data.spec.roleGroups.etcd.length>0
        // && data.spec.etcd.type !== ''
    //
    const canNextPage2To3 = data.metadata.name !== '' && data.spec.kubernetes.version !== '' && data.spec.kubernetes.containerManager !== '' && KubekeyNamespace !== ''
    //
    const canNextPage3To4 = data.spec.network.plugin !=='' && data.spec.network.kubePodsCIDR !=='' && data.spec.network.kubeServiceCIDR !== ''

    const canNextPage6To7 = !ksEnable || ksVersion !== ''

    const disablePrev = page === 0 || buttonDisabled

    const disableNext =
        (page === Object.keys(title).length - 1)
        || (page === 0 && !canNextPage0To1)
        || (page === 1 && !canNextPage1To2)
        || (page === 2 && !canNextPage2To3)
        || (page === 3 && !canNextPage3To4)
        // || (page === 5 && !canNextPage5To6)
        || (page === 6 && !canNextPage6To7)

    return (
        <InstallFormContext.Provider value={{ logs, installHandler, buttonDisabled,setButtonDisabled,ksVersion, setKsVersion, ksEnable,setKsEnable,KubekeyNamespace,setKubekeyNamespace, title, page, setPage, data, setData, canSubmit, canToHome,handleChange, disablePrev, disableNext}}>
            {children}
        </InstallFormContext.Provider>
    )
}
export default InstallFormContext

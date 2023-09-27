import React, { useState, useEffect } from 'react';
import { sortBy } from 'lodash';
import {Button, Dropdown,Menu} from "@kube-design/components";
import {Link} from "react-router-dom";

const ClusterTableDataWrapper= ({ children,clusterData }) => {

    const autoRenewCertColumn = (_,record) => {
        return(
            <div style={{display: `flex`}}>
                {record.spec.kubernetes.autoRenewCerts && <div>是</div>}
                {!record.spec.kubernetes.autoRenewCerts && <div>否</div>}
            </div>
        )
    }

    const MenuColumn = (_, record) => {
        return <Dropdown content={
            <Menu>
                <Menu.MenuItem key="upgradeCluster" >
                    <Link to={`/UpgradeCluster/${record.metadata.name}`}>
                    升级集群
                    </Link>
                </Menu.MenuItem>
                <Menu.MenuItem key="deleteCluster" >
                    <Link to={`/DeleteCluster/${record.metadata.name}`}>
                        删除集群
                    </Link>
                </Menu.MenuItem>
                <Menu.MenuItem key="addNode">
                    <Link to={`/addNode/${record.metadata.name}`}>
                    增加节点
                    </Link>
                </Menu.MenuItem>
                    <Menu.MenuItem key="deleteNode">
                        <Link to={`/deleteNode/${record.metadata.name}`}>
                            删除节点
                        </Link>
                    </Menu.MenuItem>

            </Menu>}>
            <Button type="control" size='small'>操作</Button>
        </Dropdown>
    }
    const storageColumn = (_,record) =>{
        if(record.spec && record.spec.storage && record.spec.storage.openebs !== undefined) {
            return <p>开启</p>
        } else {
            return <p>关闭</p>
        }
    }
    const initialColumns = [
        {
            children: [
                { title: '集群名',  width: '14%',dataIndex: 'metadata.name', sorter: true, search: true },
                { title: '节点数', width: '8%',render:(_, record) => record.spec.hosts.length},
                { title: 'Kubernetes 版本', dataIndex: 'spec.kubernetes.version', width: '13%' },
                {
                    title: '自动续费证书',
                    width: '14%',
                    dataIndex: 'spec.kubernetes.autoRenewCert',
                    // filters: [
                    //     { text: '是', value: true },
                    //     { text: '否', value: false },
                    // ],
                    search: true,
                    render:autoRenewCertColumn
                },
                // {title: '证书有效期', dataIndex: '', width: '14%',render:certExpiraionColumn},
                {title: '网络插件', dataIndex: 'spec.network.plugin', width: '13%'},
                {title: '容器运行时', dataIndex: 'spec.kubernetes.containerManager', width: '13%'},
                {title: '本地存储', width: '18%',render:storageColumn},
                {title: '操作', dataIndex: '', width: '18%', render: MenuColumn},
            ],
        },
    ];

    const [list, setList] = useState({
        data: [],
        isLoading: false,
        selectedRowKeys: [],
        filters: {},
        sorter: {},
        pagination: { page: 1, total: 0, limit: 10 },
    });

    useEffect(() => {
        fetchList();
    }, [clusterData]);

    const setSelectedRowKeys = (value) => {
        setList((prevState) => ({ ...prevState, selectedRowKeys: value }));
    };

    const fetchList = ({ name, pagination = {}, filters = {}, sorter = {} } = {}) => {
        setList((prevState) => ({ ...prevState, isLoading: true }));
        setTimeout(() => {
            let data = [...clusterData];

            if (name) {
                data = data.filter((item) => item.clusterName.indexOf(name) !== -1);
            }

            const filterKeys = Object.keys(filters);
            if (filterKeys.length > 0) {
                data = data.filter((item) =>
                    filterKeys.every((key) => filters[key] === item[key])
                );
            }

            if (sorter.field && sorter.order) {
                data = sortBy(data, [sorter.field]);
                if (sorter.order === 'descend') {
                    data = data.reverse();
                }
            }

            const total = data.length;
            const { page = 1, limit = 10 } = pagination;
            data = data.slice((page - 1) * limit, page * limit);

            setList({
                data,
                filters,
                sorter,
                pagination: { total, page, limit },
                isLoading: false,
            });
        }, 300);
    };

    return (
        <div>
            {children({
                list,
                columns: initialColumns,
                fetchList,
                setSelectedRowKeys,
            })}
        </div>
    );
}

export default ClusterTableDataWrapper;

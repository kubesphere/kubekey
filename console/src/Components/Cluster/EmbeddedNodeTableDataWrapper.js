import React, { useState, useEffect } from 'react';
import { sortBy } from 'lodash';
import {Tag} from "@kube-design/components";
const EmbeddedNodeTableDataWrapper= ({ children,curClusterData }) => {

    const  initialData = curClusterData.spec.hosts
    const roleColumn = (_,record) => {
        return(
            <div style={{display: `flex`}}>
                {curClusterData.spec.roleGroups.master.includes(record.name) && <Tag type="warning">MASTER</Tag>}
                {curClusterData.spec.roleGroups.master.includes(record.name) && <div style={{width:'10px'}}/>}
                {curClusterData.spec.roleGroups.worker.includes(record.name) && <Tag type="primary">WORKER</Tag>}
            </div>
        )
    }
    const initialColumns = [
        {
            children: [
                { title: 'Name', dataIndex: 'name',width:'13%', sorter: true, search: true },
                { title: 'Address', dataIndex: 'address', width: '12%' },
                { title: 'InternalAddress', dataIndex: 'internalAddress', width: '12%' },
                {
                    title: '角色',
                    dataIndex: 'role',
                    width: '15%',
                    // filters: [
                    //     { text: 'MASTER', value: 'Master' },
                    //     { text: 'WORKER', value: ['Worker'] },
                    // ],
                    search: true,
                    render:roleColumn
                },
                { title: '用户名', dataIndex: 'user', width: '12%' },
                { title: '密码', dataIndex: 'password', width: '15%' },
                { title: 'id_rsa路径', dataIndex: 'privateKeyPath', width: '20%' },
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
    }, [curClusterData.spec.hosts]);

    const setSelectedRowKeys = (value) => {
        setList((prevState) => ({ ...prevState, selectedRowKeys: value }));
    };

    const fetchList = ({ name, pagination = {}, filters = {}, sorter = {} } = {}) => {
        setList((prevState) => ({ ...prevState, isLoading: true }));
        setTimeout(() => {
            let data = [...initialData];

            if (name) {
                data = data.filter((item) => item.nodeName.indexOf(name) !== -1);
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
            const { page = 1, limit = 5 } = pagination;
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

export default EmbeddedNodeTableDataWrapper;

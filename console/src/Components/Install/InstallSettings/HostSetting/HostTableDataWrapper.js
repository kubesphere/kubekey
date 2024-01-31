import React, { useState, useEffect } from 'react';
import { sortBy } from 'lodash';
import useInstallFormContext from "../../../../hooks/useInstallFormContext";

import {Tag} from "@kube-design/components";
import HostEditModal from "../../../Modal/HostEditModal";
import HostDeleteConfirmModal from "../../../Modal/HostDeleteConfirmModal";
const HostTableDataWrapper= ({ children }) => {

    const {data} = useInstallFormContext()

    const initialData = data.spec.hosts;

    const menuColumn = (_,record) => {
        return (
            <div style={{display: `flex`}}>
                <HostEditModal record={record}/>
                <HostDeleteConfirmModal record={record}/>
            </div>
        )
    }
    const roleColumn = (_,record) => {
        return(
            <div style={{display: `flex`}}>
                {data.spec.roleGroups.master.includes(record.name) && <Tag type="warning">MASTER</Tag>}
                {data.spec.roleGroups.master.includes(record.name) && <div style={{width:'10px'}}/>}
                {data.spec.roleGroups.worker.includes(record.name) && <Tag type="primary">WORKER</Tag>}
            </div>
        )
    }

    const labelColumn = (_,record) => {
        return(
               <>
                   {
                       record.hasOwnProperty('labels') && Object.entries(record.labels).map((label, index) => {
                           if (label[0] === undefined) {
                                 return <> </>;
                           } else {
                               return (
                                   <div style={{ marginBottom: "5px" }}>
                                       <Tag key={index} type="info">{`${label[0]}: ${label[1]}`}</Tag>
                                   </div>
                               )
                           }
                       })
                   }
               </>
        )
    }

    const initialColumns = [
        {
            children: [
                { title: '主机名', dataIndex: 'name', sorter: true, search: true ,width: '8%'},
                { title: 'SSH 地址', dataIndex: 'address', width: '8%' },
                { title: '节点 IP 地址', dataIndex: 'internalAddress', width: '8%' },
                { title: '用户名', dataIndex: 'user', width: '6%' },
                {
                    title: 'CPU 架构',
                    dataIndex: 'arch',
                    width: '8%',
                    search: true,
                    // render:archColumn
                },
                {
                    title: '角色',
                    dataIndex: 'role',
                    width: '15%',
                    search: true,
                    render:roleColumn
                },
                {
                    title: '标签',
                    dataIndex: 'role',
                    width: '21%',
                    search: true,
                    render:labelColumn
                },

                { title:'操作', dataIndex:'', width: '10%', render:menuColumn }
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
    }, [data.spec.hosts]);

    const setSelectedRowKeys = (value) => {
        setList((prevState) => ({ ...prevState, selectedRowKeys: value }));
    };

    const fetchList = ({ name, pagination = {}, filters = {}, sorter = {} } = {}) => {
        setList((prevState) => ({ ...prevState, isLoading: true }));
        setTimeout(() => {
            let data = [...initialData];

            if (name) {
                data = data.filter((item) => item.name.indexOf(name) !== -1);
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

export default HostTableDataWrapper;

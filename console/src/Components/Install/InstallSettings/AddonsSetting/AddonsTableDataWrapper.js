import React, { useState, useEffect } from 'react';
import { sortBy } from 'lodash';
import useInstallFormContext from "../../../../hooks/useInstallFormContext";

import {Tag} from "@kube-design/components";

import AddonsEdit from "./AddonsEdit";
import AddonsDeleteConfirmModal from "./AddonsDeleteConfirmModal";

const AddonsTableDataWrapper= ({ children }) => {

    const {data} = useInstallFormContext()

    const initialData = data.spec.addons;

    const menuColumn = (_,record) => {
        return (
            <div style={{display: `flex`}}>
                <AddonsEdit record={record}/>
                <AddonsDeleteConfirmModal record={record}/>
            </div>
        )
    }

    const typeColumn = (_,record) => {
        return(
            <div style={{display: `flex`}}>
                {record.sources.hasOwnProperty('yaml')  && record.sources.yaml.path.length > 0 && <Tag type="warning">YAML</Tag>}
                {record.sources.hasOwnProperty('yaml')  && record.sources.yaml.path.length > 0 && <div style={{width:'10px'}}/>}
                {record.sources.hasOwnProperty('chart')  && record.sources.chart.name !== '' && <Tag type="primary">HELM</Tag>}
            </div>
        )
    }

    const initialColumns = [
        {
            children: [
                { title: '名称', dataIndex: 'name', sorter: true, search: true ,width: '10%'},
                { title: '部署命名空间', dataIndex: 'namespace', width: '10%' },
                {
                    title: '部署类型',
                    dataIndex: 'type',
                    width: '20%',
                    search: true,
                    render:typeColumn
                },
                { title:'操作', dataIndex:'', width: '13%', render:menuColumn }
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
    }, [data.spec.addons]);

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

export default AddonsTableDataWrapper;

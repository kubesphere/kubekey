import { Tag } from "@kube-design/components";
import { sortBy } from "lodash";
import React, { useEffect, useState } from "react";
import useAddNodeFormContext from "../../../hooks/useAddNodeFormContext";

const AddNodeTableDataWrapper = ({ children }) => {
  const { curCluster } = useAddNodeFormContext();
  const [initialData, setInitialData] = useState([]);
  const [initialColumns, setInitialColumns] = useState([]);

  useEffect(() => {
    if (Object.keys(curCluster).length > 0) {
      setInitialData(curCluster.spec.hosts);
      setInitialColumns([
        {
          children: [
            {
              title: "Name",
              dataIndex: "name",
              sorter: true,
              search: true,
              width: "10%",
            },
            { title: "Address", dataIndex: "address", width: "10%" },
            {
              title: "InternalAddress",
              dataIndex: "internalAddress",
              width: "13%",
            },
            {
              title: "角色",
              dataIndex: "role",
              width: "20%",
              search: true,
              render: roleColumn,
            },
            { title: "用户名", dataIndex: "user", width: "12%" },
            { title: "密码", dataIndex: "password", width: "15%" },
            { title: "id_rsa路径", dataIndex: "privateKeyPath", width: "20%" },
          ],
        },
      ]);
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [curCluster]);

  useEffect(() => {
    if (initialData.length > 0) {
      fetchList();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialData]);

  const roleColumn = (_, record) => {
    return (
      <div style={{ display: `flex` }}>
        {curCluster.spec.roleGroups.master.includes(record.name) && (
          <Tag type="warning">MASTER</Tag>
        )}
        {curCluster.spec.roleGroups.master.includes(record.name) && (
          <div style={{ width: "10px" }} />
        )}
        {curCluster.spec.roleGroups.worker.includes(record.name) && (
          <Tag type="primary">WORKER</Tag>
        )}
      </div>
    );
  };

  const [list, setList] = useState({
    data: [],
    isLoading: false,
    selectedRowKeys: [],
    filters: {},
    sorter: {},
    pagination: { page: 1, total: 0, limit: 10 },
  });
  // const setSelectedRowKeys = (value) => {
  //     setList((prevState) => ({ ...prevState, selectedRowKeys: value }));
  //     console.log('setSelectedRowKeys',value)
  // };

  const fetchList = ({
    name,
    pagination = {},
    filters = {},
    sorter = {},
  } = {}) => {
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
        if (sorter.order === "descend") {
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
        selectedRowKeys: [],
      });
    }, 300);
  };

  return (
    <div>
      {children({
        list,
        columns: initialColumns,
        fetchList,
        // setSelectedRowKeys,
      })}
    </div>
  );
};

export default AddNodeTableDataWrapper;

import React from 'react';
import {Button, InputSearch, Pagination, Table} from "@kube-design/components";
import AddNodeTableDataWrapper from "./AddNodeTableDataWrapper";

const AddNodeTable = () => {
    return (
        <AddNodeTableDataWrapper>
            {({
                  fetchList,
                  list: {
                      pagination,
                      filters,
                      sorter,
                      data,
                      isLoading,
                      // selectedRowKeys
                  },
                  setSelectedRowKeys,
                  columns
              }) => {
                // const rowSelection = {
                //     // selectedRowKeys,
                //     onSelect: (record, checked, rowKeys) => {
                //         setSelectedRowKeys(rowKeys);
                //     },
                //     onSelectAll: (checked, rowKeys) => {
                //         setSelectedRowKeys(rowKeys);
                //     },
                //     getCheckboxProps: record => ({
                //         // disabled: record.name === 'node3'
                //     })
                // };
                const title = <div style={{
                    display: "flex"
                }}>
                    <InputSearch style={{
                        flex: 1
                    }} placeholder="主机名" onSearch={name => fetchList({
                        name
                    })} />
                    <Button style={{
                        marginLeft: 12
                    }} icon="refresh" type="flat" onClick={() => fetchList({
                        pagination,
                        filters,
                        sorter
                    })} />
                </div>;
                const footer = <Pagination {...pagination} onChange={page => fetchList({
                    pagination: { ...pagination,
                        page
                    },
                    filters,
                    sorter
                })} />;
                // console.log("data is :",data)
                return <Table rowKey="name" columns={columns} filters={filters} sorter={sorter} dataSource={data} loading={isLoading} title={title} footer={footer} onChange={(filters, sorter) => fetchList({
                    filters,
                    sorter
                })} />;
            }}
        </AddNodeTableDataWrapper>
    )
}

export default AddNodeTable;

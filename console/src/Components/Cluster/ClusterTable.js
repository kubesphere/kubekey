import React from 'react';
import {Button, InputSearch, Pagination, Table} from "@kube-design/components";
import ClusterTableDataWrapper from "./ClusterTableDataWrapper";
import EmbeddedNodeTable from "./EmbeddedNodeTable";
import useClusterTableContext from "../../hooks/useClusterTableContext";


const ClusterTable = () => {
    const {clusterData} = useClusterTableContext()
    const embeddedNodeTable= record => {
        const curClusterData = record
        return <EmbeddedNodeTable curClusterData={curClusterData}/>
    }
    return (
        <div>
            <ClusterTableDataWrapper clusterData={clusterData}>
                {({
                      fetchList,
                      list: {
                          pagination,
                          filters,
                          sorter,
                          data,
                          isLoading,
                          selectedRowKeys
                      },
                      setSelectedRowKeys,
                      columns
                  }) => {
                    const title = <div style={{
                        display: "flex"
                    }}>
                        <InputSearch style={{
                            flex: 1
                        }} placeholder="please input a word" onSearch={name => fetchList({
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
                    return <Table rowKey="name" columns={columns} filters={filters} sorter={sorter} dataSource={data} loading={isLoading} title={title} footer={footer} onChange={(filters, sorter) => fetchList({
                        filters,
                        sorter
                    })} expandedRowRender={embeddedNodeTable} />;
                }}
            </ClusterTableDataWrapper>
        </div>
    )
};

export default ClusterTable;

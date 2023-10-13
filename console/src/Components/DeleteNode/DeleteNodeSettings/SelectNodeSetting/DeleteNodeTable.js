import {Table, Button, InputSearch, Pagination} from "@kube-design/components";
import DeleteNodeTableDataWrapper from "./DeleteNodeTableDataWrapper";
import useDeleteNodeFormContext from "../../../../hooks/useDeleteNodeFormContext";

const DeleteNodeTable = () => {
    const {curSelectedNodeName} = useDeleteNodeFormContext();
    return (
        <DeleteNodeTableDataWrapper >
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
            const rowSelection = {
                selectedRowKeys,
                onSelect: (record, checked, rowKeys) => {
                    setSelectedRowKeys(rowKeys);
                },
                onSelectAll: (checked, rowKeys) => {
                    // setSelectedRowKeys(rowKeys);
                },
                getCheckboxProps: record => ({
                    disabled: record.name!==curSelectedNodeName && curSelectedNodeName!==''
                })
            };
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
            return <Table rowKey="name" columns={columns} filters={filters} sorter={sorter} dataSource={data} loading={isLoading} title={title} footer={footer} rowSelection={rowSelection} onChange={(filters, sorter) => fetchList({
                filters,
                sorter
            })} />;
        }}
    </DeleteNodeTableDataWrapper>
    )
}
export default DeleteNodeTable

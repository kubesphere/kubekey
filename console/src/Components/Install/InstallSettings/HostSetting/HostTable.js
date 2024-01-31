import {Table, Button, InputSearch, Pagination} from "@kube-design/components";
import HostTableDataWrapper from "./HostTableDataWrapper";

const HostTable = () => {
    return (
        <HostTableDataWrapper>
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
            return <Table rowKey="name" columns={columns} filters={filters} sorter={sorter} dataSource={data} loading={isLoading} title={title} footer={footer} onChange={(filters, sorter) => fetchList({
                filters,
                sorter
            })} />;
        }}
    </HostTableDataWrapper>
    )
}
export default HostTable

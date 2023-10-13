import { useContext } from "react"
import ClusterTableContext from "../context/ClusterTableContext"

const useClusterTableContext = () => {
    return useContext(ClusterTableContext)
}

export default useClusterTableContext

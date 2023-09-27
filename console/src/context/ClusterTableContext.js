import {createContext, useEffect, useState} from "react";
import useGlobalContext from "../hooks/useGlobalContext";

const ClusterTableContext = createContext({})

export const ClusterTableProvider = ({children}) => {
    const {backendIP} = useGlobalContext();
    const [clusterData,setClusterData] = useState([])
    useEffect(() => {
        if(backendIP!=='') {
            fetch(`http://${backendIP}:8082/scanCluster`)
                .then(res => {
                    return res.json()
                })
                .then(data => {
                    setClusterData(data.clusterData);
                })
                .catch(error => {
                    console.error('Error fetching cluster list:', error);
                });
        }
    }, [backendIP]);
    const handleChange = newV => {
        setClusterData(prevState => [...prevState,newV])
    }

    const getClusterByName = async (clusterName) => {
        return clusterData.find(item => item.metadata.name === clusterName);
    }

    return (
        <ClusterTableContext.Provider value={{ getClusterByName,clusterData, handleChange}}>
            {children}
        </ClusterTableContext.Provider>
    )
}
export default ClusterTableContext

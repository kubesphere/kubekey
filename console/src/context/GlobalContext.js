import React, {createContext, useEffect, useState} from 'react';

const GlobalContext = createContext({})
export const GlobalProvider = ({children}) => {
    const [backendIP,setBackendIP] = useState('')
    useEffect(() => {
        setBackendIP(window.location.hostname)
    }, []);
    useEffect(()=>{
        // console.log(backendIP)
    },[backendIP])

    return (
        <GlobalContext.Provider value={{ backendIP}}>
            {children}
        </GlobalContext.Provider>
    )
}

export default GlobalContext;

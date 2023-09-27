import { useContext } from "react"
import InstallFormContext from "../context/InstallFormContext"

const useInstallFormContext = () => {
    return useContext(InstallFormContext)
}

export default useInstallFormContext

import axios from "axios"
import { isDev } from "./Utils"

export const http = axios.create({
    baseURL: isDev() ? "http://localhost:8181/api/" : "api/",
})
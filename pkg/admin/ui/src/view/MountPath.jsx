import React, { Component, useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import YAML from 'yaml'
import { http } from '../lib/Axios'

export const MountPath = () => {
    const [searchParams, setSearchParams] = useSearchParams()
    const path = searchParams.get("path")

    const [config, setConfig] = useState({})

    useEffect(() => {
        const updateState = () => {
            http.get("config").then(data => {
                setConfig(data.data)
            })
        }
        updateState()
        let intervalHandler = setInterval(updateState, 5000)
        return () => {
            clearInterval(intervalHandler)
        }
    },[])

    let middlewares = (<></>)
    if (config.Middlewares !== undefined)  {
        console.log(middlewares)
        let d = new YAML.Document()
        d.contents = config.Middlewares
        middlewares = d.toString()
    }

    return (
        <div>asd {middlewares}</div>
    )
}
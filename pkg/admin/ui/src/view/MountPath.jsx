import React, { useEffect, useState } from 'react';
import { Breadcrumb, BreadcrumbItem, Col, Container, Row, Table } from 'react-bootstrap';
import { useSearchParams } from 'react-router-dom';
import YAML from 'yaml'
import { http } from '../lib/Axios'
import { Link } from 'react-router-dom';

export const MountPath = () => {
    const [searchParams, setSearchParams] = useSearchParams()
    const path = searchParams.get("path")
    const matchHost = searchParams.get("host")

    const [config, setConfig] = useState({})

    useEffect(() => {
        const updateState = () => {
            http.get("config/yaml").then(data => {
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
    let mountPoint = {}
    if (config != "") {
        let doc = YAML.parse(config)
        if (doc != null) {
            for (let mp of doc.mountPoints) {
                    // host = mp.middlewares.matchHost
                if (mp.path === path) {
                    middlewares = YAML.stringify(mp)
                    if (mp.middlewares.matchHost !== undefined) {
                        if (mp.middlewares.matchHost === matchHost) {
                            mountPoint = mp        
                        }
                    } else {
                        mountPoint = mp
                    }
                }
            }
        }
    }
    
    return (
        <Container>
            <Breadcrumb>
                <BreadcrumbItem linkAs={Link} linkProps={{ to: "/" }}>Home</BreadcrumbItem>
                <BreadcrumbItem active>- {mountPoint.path}</BreadcrumbItem>
            </Breadcrumb>
            <Row>
                <Col><h3>MountPoint</h3></Col>
            </Row>
            <Row>
                <Col>
                <Table striped bordered hover>
                    <thead>
                    </thead>
                    <tbody>
                        <tr>
                            <th>Match Host</th>
                            <td>{matchHost}</td>
                        </tr>
                        <tr>
                            <th>Path</th>
                            <td>{mountPoint.path}</td>
                        </tr>
                        <tr>
                            <th>Upstream</th>
                            <td>{mountPoint.upstream}</td>
                        </tr>
                    </tbody>
                </Table>
                </Col>
            </Row>
            <Row style={{marginTop: "30px"}}>
                <Col>
                    <h5>Middlewares</h5>
                    <pre>{middlewares}</pre>
                </Col>
            </Row>
        </Container>
    )
}
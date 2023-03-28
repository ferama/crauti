import React, { useEffect, useState } from 'react';
import { Breadcrumb, BreadcrumbItem, Col, Container, Row, Table } from 'react-bootstrap';
import { useSearchParams } from 'react-router-dom';
import YAML from 'yaml'
import { http } from '../lib/Axios'
import { Link } from 'react-router-dom';

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
    let mountPoint = {}
    if (config.MountPoints !== undefined)  {
        for (let mp of config.MountPoints) {
            if (mp.Path === path) {
                mountPoint = mp
                let d = new YAML.Document()
                d.contents = mp.Middlewares
                middlewares = d.toString()
            }
        }
    }

    return (
        <Container>
            <Breadcrumb>
                <BreadcrumbItem linkAs={Link} linkProps={{ to: "/" }}>Home</BreadcrumbItem>
                <BreadcrumbItem active>- {mountPoint.Path}</BreadcrumbItem>
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
                            <th>Path</th>
                            <td>{mountPoint.Path}</td>
                        </tr>
                        <tr>
                            <th>Upstream</th>
                            <td>{mountPoint.Upstream}</td>
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
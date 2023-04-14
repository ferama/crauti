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

    const [mountPoint, setMountPoint] = useState({})

    useEffect(() => {
        const updateState = () => {
            http.get(`mount-point?path=${path}&host=${matchHost}`).then(data => {
                let d = data.data
                if (d.length > 0) {
                    setMountPoint(d[0])
                }
            })
        }
        updateState()
        let intervalHandler = setInterval(updateState, 5000)
        return () => {
            clearInterval(intervalHandler)
        }
    },[])

    let middlewares = (<></>)
    if (mountPoint != "") {
        middlewares = YAML.stringify(mountPoint.Middlewares)
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
                            <th>Match Host</th>
                            <td>{mountPoint.MatchHost}</td>
                        </tr>
                        <tr>
                            <th>Path</th>
                            <td>{mountPoint.Path}</td>
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
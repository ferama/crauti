import React, { useEffect, useState } from 'react';
import { Breadcrumb, BreadcrumbItem, Col, Container, Row } from 'react-bootstrap';
import Table from 'react-bootstrap/Table';
import { http } from '../lib/Axios'
import { Link } from 'react-router-dom';
import YAML from 'yaml'

export const Home = () => {
  const [mountPoints, setMountPoints] = useState([])

  useEffect(() => {
    // like componentDidMount
    const updateState = () => {
        http.get("mount-point").then(data => {
            // setMountPoints(data.data)
            setMountPoints(YAML.parse(data.data))
        })
    }
    updateState()
    let intervalHandler = setInterval(updateState, 5000)
    return () => {
        // like componentWillUnmount
        clearInterval(intervalHandler)
    }
  },[])

  let rows = (<></>)
  rows = mountPoints.map(mp => {
    let key = `${mp.path}-${mp.matchHost}`
    return (
      <tr key={key}>
        <td>{mp.matchHost}</td>
        <td>{mp.path}</td>
        <td>{mp.upstream}</td>
        <td><Link to={"/mount?path=" + mp.path + "&host=" + mp.matchHost}>details</Link></td>
      </tr>
  )})
  return (
    <Container>
      <Breadcrumb>
        <BreadcrumbItem active>Home</BreadcrumbItem>
      </Breadcrumb>
      <Row>
        <Col><h3>MountPoints</h3></Col>
      </Row>
      <Row>
        <Col>
          <Table striped bordered hover>
            <thead>
              <tr>
                <th>Match Host</th>
                <th>Path</th>
                <th>Upstream</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {rows}
            </tbody>
          </Table>
        </Col>
      </Row>
    </Container>
  )
}
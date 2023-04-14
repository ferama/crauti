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
            setMountPoints(data.data)
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
    let key = `${mp.Path}-${mp.MatchHost}`
    return (
      <tr key={key}>
        <td>{mp.MatchHost}</td>
        <td>{mp.Path}</td>
        <td>{mp.Upstream}</td>
        <td><Link to={"/mount?path=" + mp.Path + "&host=" + mp.MatchHost}>details</Link></td>
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
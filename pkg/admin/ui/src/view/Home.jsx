import React, { useEffect, useState } from 'react';
import { Breadcrumb, BreadcrumbItem, Col, Container, Row } from 'react-bootstrap';
import Table from 'react-bootstrap/Table';
import { http } from '../lib/Axios'
import { Link } from 'react-router-dom';

export const Home = () => {
  const [config, setConfig] = useState({})

  useEffect(() => {
    // like componentDidMount
    const updateState = () => {
        http.get("config").then(data => {
            setConfig(data.data)
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
  if ((config.MountPoints !== undefined) && (config.MountPoints != null)) {
    rows = config.MountPoints.map(mp => {
      let key = `${mp.Path}-${mp.Middlewares.MatchHost}`
      return (
        <tr key={key}>
          <td>{mp.Middlewares.MatchHost}</td>
          <td>{mp.Path}</td>
          <td>{mp.Upstream}</td>
          <td><Link to={"/mount?path=" + mp.Path + "&host=" + mp.Middlewares.MatchHost}>details</Link></td>
        </tr>
    )})
  }
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
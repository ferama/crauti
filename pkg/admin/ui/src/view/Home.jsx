import React, { Component } from 'react';
import { Col, Container, Row } from 'react-bootstrap';
import Table from 'react-bootstrap/Table';
import { http } from '../lib/Axios'
import YAML from 'yaml'

export class Home extends Component {
  state = {
    config: {}
  }
  intervalHandler = null

  async componentDidMount() {
    await this.updateState()
    this.intervalHandler = setInterval(this.updateState, 5000)
  }

  componentWillUnmount() {
    clearInterval(this.intervalHandler)
  }

  updateState = async () => {
    let data
    try {
        data = await http.get("config")
    } catch {
        return
    }
    if (data.data === null) return
    this.setState({
        "config": data.data
    })
  }

  render() {
    let rows = (<></>)
    if (this.state.config.MountPoints !== undefined) {
      rows = this.state.config.MountPoints.map(mp => (
        <tr key={mp.Path}>
          <td>{mp.Path}</td>
          <td>{mp.Upstream}</td>
        </tr>
      ))
    }

    let middlewares = (<></>)
    if (this.state.config.Middlewares !== undefined)  {
      let d = new YAML.Document()
      d.contents = this.state.config.Middlewares
      middlewares = d.toString()
    }

    return (
      <Container>
        <Row>
          <Col><h3>Middlewares</h3></Col>
        </Row>
        <Row>
          <Col><pre>{middlewares}</pre></Col>
        </Row>
        <Row>
          <Col><h3>MountPoints</h3></Col>
        </Row>
        <Row>
          <Col>
            <Table>
              <thead>
                <tr>
                  <th>MountPoint</th>
                  <th>Upstream</th>
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
}
import React, { Component } from 'react';
import { Col, Container, Row } from 'react-bootstrap';
import Table from 'react-bootstrap/Table';
import { http } from '../lib/Axios'

export class Home extends Component {
  state = {
    config: {}
  }

  async componentDidMount() {
    await this.updateState()
  }

  updateState = async () => {
    let data
    try {
        data = await http.get("config")
    } catch {
        return
    }
    console.log(data)
  }

  render() {
    return (
      <Container>
        <Row>
          <Col><h1>Home</h1></Col>
        </Row>
        <Row>
          <Col>
            <Table>
            <thead>
              <tr>
                <th>#</th>
                <th>Col</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>1</td>
                <td>value1</td>
              </tr>
              <tr>
                <td>2</td>
                <td>value2</td>
              </tr>
            </tbody>
            </Table>
          </Col>
        </Row>
      </Container>
    )
  }
}
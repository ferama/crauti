import React, { useEffect, useState } from 'react';
import { Breadcrumb, BreadcrumbItem, Col, Container, Row } from 'react-bootstrap';
import { http } from '../lib/Axios'
import YAML from 'yaml'

export const Config = () => {
    const [config, setConfig] = useState({})

    useEffect(() => {
      // like componentDidMount
      const updateState = () => {
          http.get("config/yaml").then(data => {
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

    let out = (<></>)
    if (config != "") {
        let doc = YAML.parse(config)
        if (doc != null) delete doc.mountPoints
        out = YAML.stringify(doc)
    }
    return(
        <Container>
            <Breadcrumb>
                <BreadcrumbItem active>Config</BreadcrumbItem>
            </Breadcrumb>
            <Row>
                <Col><h3>Global Config</h3></Col>
            </Row>
            <Row>
                <Col>
                    <pre>{out}</pre>
                </Col>
            </Row>
        </Container>
    )
}
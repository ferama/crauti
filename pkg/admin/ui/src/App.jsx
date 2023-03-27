import {
  // BrowserRouter as Router,
  HashRouter as Router
} from "react-router-dom";

import ThemeProvider from 'react-bootstrap/ThemeProvider'

import { AppBar } from './components/AppBar'
import { Routes } from "./Routes";

export const App = () => {
  const style = {
    marginTop: "40px",
  }
  return (
    <Router>
      <ThemeProvider
          breakpoints={['xxxl', 'xxl', 'xl', 'lg', 'md', 'sm', 'xs', 'xxs']}
          minBreakpoint="xs"
        >
        <AppBar/>
        <div style={style}>
          <Routes />
        </div>
      </ThemeProvider>
    </Router>
  )
}
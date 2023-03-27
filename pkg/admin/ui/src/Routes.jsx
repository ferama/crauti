import React from 'react';
import {
    Routes as RouterRoutes,
    Route,
  } from "react-router-dom";
import { Home } from './view/Home';

export const Routes = () => (
    <RouterRoutes>
        <Route path="/" element={<Home />} />
        <Route path="/home" element={<Home />} />
    </RouterRoutes>
)
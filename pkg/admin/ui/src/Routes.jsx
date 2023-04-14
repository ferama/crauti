import React from 'react';
import {
    Routes as RouterRoutes,
    Route,
  } from "react-router-dom";
import { Config } from './view/Config';
import { Home } from './view/Home';
import { MountPoint } from './view/MountPoint';

export const Routes = () => (
    <RouterRoutes>
        <Route path="/" element={<Home />} />
        <Route path="/home" element={<Home />} />
        <Route path="/mount" element={<MountPoint />}></Route>
        <Route path="/config" element={<Config />}></Route>
    </RouterRoutes>
)


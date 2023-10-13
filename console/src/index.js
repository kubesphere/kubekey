import React from 'react';
import { createRoot } from 'react-dom/client';
import { CssBaseline, KubedConfigProvider} from '@kubed/components';
import App from "./App";
import "@kube-design/components/esm/styles/index.css";
import {HashRouter} from "react-router-dom";
const Application = () => (
    <HashRouter>
        <KubedConfigProvider>
            <CssBaseline />
            <App />
        </KubedConfigProvider>
    </HashRouter>
);
const root = document.getElementById('root');
createRoot(root).render(<Application />);

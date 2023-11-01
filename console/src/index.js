import "@kube-design/components/esm/styles/index.css";
import { CssBaseline, KubedConfigProvider } from "@kubed/components";
import React from "react";
import { createRoot } from "react-dom/client";
import { HashRouter } from "react-router-dom";
import App from "./App";
const Application = () => (
  <HashRouter>
    <KubedConfigProvider>
      <CssBaseline />
      <App />
    </KubedConfigProvider>
  </HashRouter>
);
const root = document.getElementById("root");
createRoot(root).render(<Application />);

import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import App from "@/App";
import Docs from "@/Docs";
import { siteTitle } from "@/config";
import "@/index.css";

document.title = siteTitle;

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/docs" element={<Docs />} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
);

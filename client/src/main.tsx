import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import "bootstrap/dist/css/bootstrap.min.css";
import "bootstrap-icons/font/bootstrap-icons.css";
import App from "./App";
import { DemoProvider } from "./context/DemoContext";
import { ImportProvider } from "./context/ImportProvider";
import { HelmetProvider } from "react-helmet-async";
import { BrowserRouter } from "react-router-dom";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <HelmetProvider>
      <DemoProvider>
        <ImportProvider>
          <BrowserRouter>
            <App />
          </BrowserRouter>
        </ImportProvider>
      </DemoProvider>
    </HelmetProvider>
  </StrictMode>,
);

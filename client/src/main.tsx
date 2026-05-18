import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import "bootstrap/dist/css/bootstrap.min.css";
import "bootstrap-icons/font/bootstrap-icons.css";
import App from "./App";
import { DemoProvider } from "./context/DemoContext";
import { HelmetProvider } from "react-helmet-async";
import { BrowserRouter } from "react-router-dom";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <HelmetProvider>
      <DemoProvider>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </DemoProvider>
    </HelmetProvider>
  </StrictMode>,
);

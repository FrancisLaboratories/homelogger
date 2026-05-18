import React from "react";
import { Routes, Route, useLocation } from "react-router-dom";
import { Helmet } from "react-helmet-async";
import MyNavbar from "@/components/Navbar";
import HomePage from "./pages/HomePage";
import AppliancesPage from "./pages/AppliancesPage";
import AppliancePage from "./pages/AppliancePage";
import SpacePage from "./pages/SpacePage";
import SettingsPage from "./pages/SettingsPage";
import { APP_VERSION } from "./version";

const getPageName = (p: string) => {
  if (!p || p === "/") return "Home";
  const first = p.split("/").filter(Boolean)[0] || "Home";
  return first
    .split(/[-_]/)
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
};

const App: React.FC = () => {
  const location = useLocation();
  const pageName = getPageName(location.pathname);

  return (
    <>
      <Helmet>
        <title>{`HomeLogger | ${pageName}`}</title>
      </Helmet>
      <div className="container">
        <MyNavbar />
        <main style={{ marginTop: 18 }}>
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/appliances" element={<AppliancesPage />} />
            <Route path="/appliance" element={<AppliancePage />} />

            <Route
              path="/building-exterior"
              element={
                <SpacePage
                  spaceType="BuildingExterior"
                  title="Building Exterior"
                />
              }
            />
            <Route
              path="/building-interior"
              element={
                <SpacePage
                  spaceType="BuildingInterior"
                  title="Building Interior"
                />
              }
            />
            <Route
              path="/electrical"
              element={<SpacePage spaceType="Electrical" title="Electrical" />}
            />
            <Route
              path="/hvac"
              element={<SpacePage spaceType="HVAC" title="HVAC" />}
            />
            <Route
              path="/plumbing"
              element={<SpacePage spaceType="Plumbing" title="Plumbing" />}
            />
            <Route
              path="/yard"
              element={<SpacePage spaceType="Yard" title="Yard" />}
            />

            <Route path="/settings" element={<SettingsPage />} />
            <Route path="*" element={<div>Not found</div>} />
          </Routes>
        </main>
        <footer style={{ padding: "12px 0", marginTop: "24px" }}>
          <div
            style={{
              textAlign: "center",
              color: "#6c757d",
              fontSize: "0.9rem",
            }}
          >
            <span
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: 8,
              }}
            >
              <img
                src="/logoname.png"
                alt="HomeLogger"
                style={{ height: 22 }}
              />
              <span>v{APP_VERSION}</span>
              <a
                href="https://github.com/FrancisLaboratories/homelogger"
                target="_blank"
                rel="noopener noreferrer"
                style={{
                  display: "inline-flex",
                  alignItems: "center",
                  color: "#6c757d",
                  textDecoration: "none",
                  marginLeft: 8,
                }}
              >
                <img
                  src="/github.png"
                  alt="GitHub"
                  style={{ width: 16, height: 16, marginLeft: 6 }}
                />
              </a>
            </span>
          </div>
          <div
            style={{
              textAlign: "center",
              color: "#6c757d",
              fontSize: "0.8rem",
              marginTop: "4px",
            }}
          >
            Made with &#x2665; in Detroit
          </div>
        </footer>
      </div>
    </>
  );
};

export default App;

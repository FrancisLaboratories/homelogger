import React, { useContext } from "react";
import { Nav, Navbar } from "react-bootstrap";
import { Link, useLocation } from "react-router-dom";
import { DemoContext } from "@/context/DemoContext";

const MyNavbar: React.FC = () => {
  const location = useLocation();
  const { isDemo } = useContext(DemoContext);

  return (
    <>
      {isDemo && (
        <div
          style={{
            width: "100%",
            background: "#fff3cd",
            color: "#856404",
            padding: "6px 12px",
            fontWeight: 600,
            textAlign: "center",
            fontSize: "0.95rem",
            boxSizing: "border-box",
          }}
          aria-hidden={false}
        >
          <div>This is a public demo — data will reset periodically</div>
          <div
            style={{
              fontSize: "0.85rem",
              color: "#6c757d",
              fontWeight: 600,
              marginTop: 4,
            }}
          >
            Please do not enter personal or sensitive information. You should have no expectation of privacy when using this demo site. All data is publicly accessible and may be deleted at any time.
          </div>
          <div
            style={{
              fontSize: "0.75rem",
              color: "#6c757d",
              fontWeight: 500,
              marginTop: 6,
            }}
          >
            All users must follow the{" "}
            <a
              href="https://github.com/FrancisLaboratories/homelogger/blob/main/CODE_OF_CONDUCT.md"
              target="_blank"
              rel="noopener noreferrer"
              style={{ color: "#856404", textDecoration: "underline" }}
            >
              Code of Conduct
            </a>
            . Inappropriate behavior may result in removal of demo or IP
            blacklisting.
          </div>
        </div>
      )}
      <Navbar expand="lg">
        <Navbar.Brand as={Link} to="/">
          <img src="/logoname.png" alt="HomeLogger" style={{ height: 28 }} />
        </Navbar.Brand>
        <Navbar.Toggle aria-controls="basic-navbar-nav" />
        <Navbar.Collapse id="basic-navbar-nav">
          <Nav className="mr-auto">
            <Nav.Link as={Link} to="/" active={location.pathname === "/"}>
              Home
            </Nav.Link>
            <Nav.Link
              as={Link}
              to="/appliances"
              active={location.pathname === "/appliances"}
            >
              Appliances
            </Nav.Link>
            <Nav.Link
              as={Link}
              to="/building-exterior"
              active={location.pathname === "/building-exterior"}
            >
              Building Exterior
            </Nav.Link>
            <Nav.Link
              as={Link}
              to="/building-interior"
              active={location.pathname === "/building-interior"}
            >
              Building Interior
            </Nav.Link>
            <Nav.Link
              as={Link}
              to="/electrical"
              active={location.pathname === "/electrical"}
            >
              Electrical
            </Nav.Link>
            <Nav.Link
              as={Link}
              to="/hvac"
              active={location.pathname === "/hvac"}
            >
              HVAC
            </Nav.Link>
            <Nav.Link
              as={Link}
              to="/plumbing"
              active={location.pathname === "/plumbing"}
            >
              Plumbing
            </Nav.Link>
            <Nav.Link
              as={Link}
              to="/yard"
              active={location.pathname === "/yard"}
            >
              Yard
            </Nav.Link>
          </Nav>

          <Nav className="ms-auto">
            <Nav.Link
              as={Link}
              to="/settings"
              active={location.pathname === "/settings"}
              aria-label="Settings"
              title="Settings"
            >
              <i className="bi bi-gear-fill" style={{ fontSize: "1.15rem" }} />
            </Nav.Link>
          </Nav>
        </Navbar.Collapse>
      </Navbar>
    </>
  );
};

export default MyNavbar;

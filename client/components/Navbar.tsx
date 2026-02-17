// client/components/Navbar.tsx
import { useRouter } from 'next/router';
import { Nav, Navbar } from 'react-bootstrap';
import Link from 'next/link';

const MyNavbar: React.FC = () => {
  const router = useRouter();

  return (
    <Navbar expand="lg">
      <Navbar.Brand as={Link} href="/">
        <img src="/logoname.png" alt="HomeLogger" style={{ height: 28 }} />
      </Navbar.Brand>
      <Navbar.Toggle aria-controls="basic-navbar-nav" />
      <Navbar.Collapse id="basic-navbar-nav">
        <Nav className="mr-auto">
          <Nav.Link as={Link} href="/" active={router.pathname === '/'}>Home</Nav.Link>
          <Nav.Link as={Link} href="/upgrades.html" active={router.pathname === '/upgrades'}>Upgrades</Nav.Link>
          <Nav.Link as={Link} href="/budgeting.html" active={router.pathname === '/budgeting'}>Budgeting</Nav.Link>
          <Nav.Link as={Link} href="/recurring.html" active={router.pathname === '/recurring'}>Recurring</Nav.Link>
          <Nav.Link as={Link} href="/appliances.html" active={router.pathname === '/appliances.html'}>Appliances</Nav.Link>
          <Nav.Link as={Link} href="/building-exterior.html" active={router.pathname === '/building-exterior.html'}>Building Exterior</Nav.Link>
          <Nav.Link as={Link} href="/building-interior.html" active={router.pathname === '/building-interior.html'}>Building Interior</Nav.Link>
          <Nav.Link as={Link} href="/electrical.html" active={router.pathname === '/electrical.html'}>Electrical</Nav.Link>
          <Nav.Link as={Link} href="/hvac.html" active={router.pathname === '/hvac.html'}>HVAC</Nav.Link>
          <Nav.Link as={Link} href="/plumbing.html" active={router.pathname === '/plumbing.html'}>Plumbing</Nav.Link>
          <Nav.Link as={Link} href="/yard.html" active={router.pathname === '/yard.html'}>Yard</Nav.Link>
        </Nav>

        <Nav className="ms-auto">
          <Nav.Link as={Link} href="/settings" active={router.pathname === '/settings'} aria-label="Settings" title="Settings">
            <i className="bi bi-gear-fill" style={{ fontSize: '1.15rem' }} />
          </Nav.Link>
        </Nav>
      </Navbar.Collapse>
    </Navbar>
  );
};

export default MyNavbar;

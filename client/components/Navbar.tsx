// client/components/Navbar.tsx
import React, { useContext } from 'react'
import { useRouter } from 'next/router'
import { Nav, Navbar } from 'react-bootstrap'
import Link from 'next/link'
import { DemoContext } from '../pages/_app'

const MyNavbar: React.FC = () => {
    const router = useRouter()
    const { isDemo } = useContext(DemoContext)

    return (
        <>
            {isDemo && (
                <div
                    style={{
                        width: '100%',
                        background: '#fff3cd',
                        color: '#856404',
                        padding: '6px 12px',
                        fontWeight: 600,
                        textAlign: 'center',
                        fontSize: '0.95rem',
                        boxSizing: 'border-box',
                    }}
                    aria-hidden={false}
                >
                    <div>This is a public demo — data will reset periodically.</div>
                    <div
                        style={{
                            fontSize: '0.85rem',
                            color: '#6c757d',
                            fontWeight: 600,
                            marginTop: 4,
                        }}
                    >
                        Please do not enter personal information.
                    </div>
                    <div
                        style={{
                            fontSize: '0.75rem',
                            color: '#6c757d',
                            fontWeight: 500,
                            marginTop: 6,
                        }}
                    >
                        All users must follow the{' '}
                        <a
                            href="https://github.com/FrancisLaboratories/homelogger/blob/main/CODE_OF_CONDUCT.md"
                            target="_blank"
                            rel="noopener noreferrer"
                            style={{ color: '#856404', textDecoration: 'underline' }}
                        >
                            Code of Conduct
                        </a>
                        . Inappropriate behavior may result in removal of demo or IP blacklisting.
                    </div>
                </div>
            )}
            <Navbar expand="lg">
                <Navbar.Brand as={Link} href="/">
                    <img src="/logoname.png" alt="HomeLogger" style={{ height: 28 }} />
                </Navbar.Brand>
                <Navbar.Toggle aria-controls="basic-navbar-nav" />
                <Navbar.Collapse id="basic-navbar-nav">
                    <Nav className="mr-auto">
                        <Nav.Link as={Link} href="/" active={router.pathname === '/'}>
                            Home
                        </Nav.Link>
                        <Nav.Link
                            as={Link}
                            href="/appliances.html"
                            active={router.pathname === '/appliances.html'}
                        >
                            Appliances
                        </Nav.Link>
                        <Nav.Link
                            as={Link}
                            href="/building-exterior.html"
                            active={router.pathname === '/building-exterior.html'}
                        >
                            Building Exterior
                        </Nav.Link>
                        <Nav.Link
                            as={Link}
                            href="/building-interior.html"
                            active={router.pathname === '/building-interior.html'}
                        >
                            Building Interior
                        </Nav.Link>
                        <Nav.Link
                            as={Link}
                            href="/electrical.html"
                            active={router.pathname === '/electrical.html'}
                        >
                            Electrical
                        </Nav.Link>
                        <Nav.Link
                            as={Link}
                            href="/hvac.html"
                            active={router.pathname === '/hvac.html'}
                        >
                            HVAC
                        </Nav.Link>
                        <Nav.Link
                            as={Link}
                            href="/plumbing.html"
                            active={router.pathname === '/plumbing.html'}
                        >
                            Plumbing
                        </Nav.Link>
                        <Nav.Link
                            as={Link}
                            href="/yard.html"
                            active={router.pathname === '/yard.html'}
                        >
                            Yard
                        </Nav.Link>
                    </Nav>

                    <Nav className="ms-auto">
                        <Nav.Link
                            as={Link}
                            href="/settings"
                            active={router.pathname === '/settings'}
                            aria-label="Settings"
                            title="Settings"
                        >
                            <i className="bi bi-gear-fill" style={{ fontSize: '1.15rem' }} />
                        </Nav.Link>
                    </Nav>
                </Navbar.Collapse>
            </Navbar>
        </>
    )
}

export default MyNavbar

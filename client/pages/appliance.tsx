'use client';
import React, {useEffect, useState} from 'react';
import {Button, Card, Col, Container, Row, Tab, Tabs} from 'react-bootstrap';
import "bootstrap-icons/font/bootstrap-icons.css";
import MaintenanceSection, {MaintenanceReferenceType, MaintenanceSpaceType} from '@/components/MaintenanceSection';
import RepairSection, {RepairReferenceType, RepairSpaceType} from '@/components/RepairSection';
import DocumentationSection from '@/components/DocumentationSection';
import TodosSection from '@/components/TodosSection';
import NotesSection from '@/components/NotesSection';
import MyNavbar from '@/components/Navbar';
import {SERVER_URL} from "@/pages/_app";
import EditApplianceModal from '@/components/EditApplianceModal';

interface Appliance {
    id: number;
    applianceName: string;
    manufacturer: string;
    modelNumber: string;
    serialNumber: string;
    yearPurchased: string;
    purchasePrice: string;
    location: string;
    type: string;
}

const AppliancePage: React.FC = () => {
    // Get id from query string (works with static export)
    const getIdFromQuery = () => {
        if (typeof window !== 'undefined') {
            const params = new URLSearchParams(window.location.search);
            return params.get('id');
        }
        return null;
    };
    const [id, setId] = useState<string | null>(null);
    const [appliance, setAppliance] = useState<Appliance | null>(null);
    const [showEditModal, setShowEditModal] = useState(false);

    useEffect(() => {
        setId(getIdFromQuery());
    }, []);

    useEffect(() => {
        if (id) {
            fetch(`${SERVER_URL}/appliances/${id}`)
                .then(response => response.json())
                .then(data => setAppliance(data))
                .catch(error => console.error('Error fetching appliance:', error));
        }
    }, [id]);

    const handleDelete = async () => {
        if (id && window.confirm('Are you sure you want to delete this appliance?')) {
            try {
                const response = await fetch(`${SERVER_URL}/appliances/delete/${id}`, {
                    method: 'DELETE',
                });

                if (!response.ok) {
                    throw new Error('Failed to delete appliance');
                }

                window.location.href = '/appliances.html';
            } catch (error) {
                console.error('Error deleting appliance:', error);
            }
        }
    };

    const handleShowEditModal = () => setShowEditModal(true);
    const handleCloseEditModal = () => setShowEditModal(false);

    const handleSaveAppliance = (id: number, applianceName: string, manufacturer: string, modelNumber: string, serialNumber: string, yearPurchased: string, purchasePrice: string, location: string, type: string) => {
        setAppliance({
            id,
            applianceName,
            manufacturer,
            modelNumber,
            serialNumber,
            yearPurchased,
            purchasePrice,
            location,
            type
        });
    };

    if (!id) {
        return <div>No appliance id provided.</div>;
    }
    if (!appliance) {
        return <div>Loading...</div>;
    }

    return (
        <Container>
            <MyNavbar/>
            <Row className="justify-content-center">
                <Col md={8}>
                    <Tabs defaultActiveKey="main" id="appliance-tabs">
                        <Tab eventKey="main" title="Main">
                            <Card>
                                <Card.Body>
                                    <Card.Title>{appliance.applianceName}</Card.Title>
                                    <Card.Subtitle className="mb-2 text-muted">{appliance.type}</Card.Subtitle>
                                    <Card.Text>
                                        <strong>Location:</strong> {appliance.location}<br/>
                                        <strong>Manufacturer:</strong> {appliance.manufacturer}<br/>
                                        <strong>Model Number:</strong> {appliance.modelNumber}<br/>
                                        <strong>Serial Number:</strong> {appliance.serialNumber}<br/>
                                        <strong>Year Purchased:</strong> {appliance.yearPurchased}<br/>
                                        <strong>Purchase Price:</strong> {appliance.purchasePrice}<br/>
                                    </Card.Text>
                                    <Row>
                                        <Col>
                                            <Button variant="secondary" style={{marginTop: '10px'}} onClick={handleShowEditModal}>
                                                <i className="bi bi-pencil-fill"></i>
                                            </Button>
                                            <Button variant="danger" onClick={handleDelete} style={{marginTop: '10px', float: 'right'}}>
                                                <i className="bi bi-trash"></i>
                                            </Button>
                                        </Col>
                                    </Row>
                                </Card.Body>
                            </Card>
                        </Tab>
                        <Tab eventKey="maintenance" title="Maintenance">
                            <MaintenanceSection applianceId={appliance.id} referenceType={MaintenanceReferenceType.Appliance}
                                                spaceType={MaintenanceSpaceType.NotApplicable}/>
                        </Tab>
                        <Tab eventKey="repairs" title="Repairs">
                            <RepairSection applianceId={appliance.id} referenceType={RepairReferenceType.Appliance}
                                                spaceType={RepairSpaceType.NotApplicable}/>
                        </Tab>
                        <Tab eventKey="documentation" title="Documentation">
                            <DocumentationSection applianceId={appliance.id} />
                        </Tab>
                        <Tab eventKey="todos" title="To-dos">
                            <TodosSection applianceId={appliance.id} />
                        </Tab>
                        <Tab eventKey="notes" title="Notes">
                            <NotesSection applianceId={appliance.id} />
                        </Tab>
                    </Tabs>
                    <Button variant="secondary" onClick={() => window.location.href = '/appliances.html'} style={{marginTop: '10px'}}>
                        Back
                    </Button>
                </Col>
            </Row>
            <EditApplianceModal
                show={showEditModal}
                handleClose={handleCloseEditModal}
                handleSave={handleSaveAppliance}
                appliance={appliance}
            />
        </Container>
    );
};

export default AppliancePage;

import React from 'react';
import {Container, Tab, Tabs, Card} from 'react-bootstrap';
import MaintenanceSection, {MaintenanceReferenceType, MaintenanceSpaceType} from '@/components/MaintenanceSection';
import RepairSection, {RepairReferenceType, RepairSpaceType} from '@/components/RepairSection';
import DocumentationSection from '@/components/DocumentationSection';
import TodosSection from '@/components/TodosSection';
import NotesSection from '@/components/NotesSection';
import MyNavbar from '../components/Navbar';

const HVACPage: React.FC = () => {
    return (
        <Container style={{marginTop: '16px'}}>
            <MyNavbar />
            <h3>HVAC</h3>
            <Tabs defaultActiveKey="maintenance" id="hvac-tabs" className="mb-3">
                <Tab eventKey="maintenance" title="Maintenance">
                    <MaintenanceSection referenceType={MaintenanceReferenceType.Space} spaceType={MaintenanceSpaceType.HVAC} />
                </Tab>
                <Tab eventKey="repair" title="Repairs">
                    <RepairSection referenceType={RepairReferenceType.Space} spaceType={RepairSpaceType.HVAC} />
                </Tab>
                <Tab eventKey="documents" title="Documents">
                    <DocumentationSection spaceType={MaintenanceSpaceType.HVAC} />
                </Tab>
                <Tab eventKey="notes" title="Notes">
                    <NotesSection spaceType={MaintenanceSpaceType.HVAC} />
                </Tab>
                <Tab eventKey="todos" title="To-dos">
                    <TodosSection spaceType={MaintenanceSpaceType.HVAC} />
                </Tab>
            </Tabs>
        </Container>
    );
};

export default HVACPage;

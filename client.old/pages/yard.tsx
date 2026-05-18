import React from 'react'
import { Container, Tab, Tabs, Card } from 'react-bootstrap'
import MaintenanceSection, {
    MaintenanceReferenceType,
    MaintenanceSpaceType,
} from '@/components/MaintenanceSection'
import RepairSection, { RepairReferenceType, RepairSpaceType } from '@/components/RepairSection'
import DocumentationSection from '@/components/DocumentationSection'
import TasksSection from '@/components/TasksSection'
import NotesSection from '@/components/NotesSection'
import MyNavbar from '../components/Navbar'

const YardPage: React.FC = () => {
    return (
        <Container style={{ marginTop: '16px' }}>
            <MyNavbar />
            <h3>Yard</h3>
            <Tabs defaultActiveKey="maintenance" id="yard-tabs" className="mb-3">
                <Tab eventKey="maintenance" title="Maintenance">
                    <MaintenanceSection
                        referenceType={MaintenanceReferenceType.Space}
                        spaceType={MaintenanceSpaceType.Yard}
                    />
                </Tab>
                <Tab eventKey="repair" title="Repairs">
                    <RepairSection
                        referenceType={RepairReferenceType.Space}
                        spaceType={RepairSpaceType.Yard}
                    />
                </Tab>
                <Tab eventKey="documents" title="Documents">
                    <DocumentationSection spaceType={MaintenanceSpaceType.Yard} />
                </Tab>
                <Tab eventKey="notes" title="Notes">
                    <NotesSection spaceType={MaintenanceSpaceType.Yard} />
                </Tab>
                <Tab eventKey="todos" title="Tasks">
                    <TasksSection spaceType="Yard" />
                </Tab>
            </Tabs>
        </Container>
    )
}

export default YardPage

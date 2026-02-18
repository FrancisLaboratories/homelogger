import React, { useEffect, useState } from 'react'
import { Card, Table } from 'react-bootstrap'
import 'bootstrap-icons/font/bootstrap-icons.css'
import { SERVER_URL } from '@/pages/_app'
import AddMaintenanceModal from '@/components/AddMaintenanceModal'
import ShowMaintenanceModal from '@/components/ShowMaintenanceModal'

export enum MaintenanceReferenceType {
    Appliance = 'Appliance',
    Space = 'Space',
}

export enum MaintenanceSpaceType {
    BuildingExterior = 'BuildingExterior',
    BuildingInterior = 'BuildingInterior',
    Electrical = 'Electrical',
    HVAC = 'HVAC',
    Plumbing = 'Plumbing',
    Yard = 'Yard',
    NotApplicable = 'NotApplicable',
}

export interface MaintenanceRecord {
    id: number
    description: string
    date: string
    cost: number
    notes: string
    spaceType: MaintenanceSpaceType
    referenceType: MaintenanceReferenceType
    applianceId: number
    attachmentIds?: number[]
}

interface MaintenanceProps {
    applianceId?: number
    referenceType: MaintenanceReferenceType
    spaceType?: MaintenanceSpaceType
}

const MaintenanceSection: React.FC<MaintenanceProps> = ({
    applianceId,
    referenceType,
    spaceType,
}) => {
    const [maintenanceRecords, setMaintenanceRecords] = useState<MaintenanceRecord[]>([])
    const [showAddModal, setShowAddModal] = useState(false)
    const [showViewModal, setShowViewModal] = useState(false)
    const [selectedRecord, setSelectedRecord] = useState<MaintenanceRecord | null>(null)

    useEffect(() => {
        const fetchMaintenanceRecords = async () => {
            try {
                const queryParams = new URLSearchParams({
                    applianceId: applianceId?.toString() || '',
                    referenceType,
                    spaceType: spaceType || '',
                }).toString()

                const response = await fetch(`${SERVER_URL}/maintenance?${queryParams}`)
                const data = await response.json()
                setMaintenanceRecords(data)
            } catch (error) {
                console.error('Error fetching maintenance records:', error)
            }
        }

        fetchMaintenanceRecords()
    }, [applianceId, referenceType, spaceType])

    const handleShowAddModal = () => setShowAddModal(true)
    const handleCloseAddModal = () => setShowAddModal(false)
    const handleSaveMaintenance = (newMaintenance: MaintenanceRecord) => {
        setMaintenanceRecords([...maintenanceRecords, newMaintenance])
    }

    const handleRowClick = (record: MaintenanceRecord) => {
        setSelectedRecord(record)
        setShowViewModal(true)
    }

    const handleCloseViewModal = () => setShowViewModal(false)

    const handleDeleteMaintenance = (id: number) => {
        setMaintenanceRecords(maintenanceRecords.filter((record) => record.id !== id))
    }

    const totalCost = maintenanceRecords.reduce((sum, record) => sum + record.cost, 0)

    return (
        <Card>
            <Card.Body>
                <Table striped bordered hover>
                    <thead>
                        <tr>
                            <th>Description</th>
                            <th>Cost</th>
                            <th>Date</th>
                        </tr>
                    </thead>
                    <tbody>
                        {maintenanceRecords.length === 0 ? (
                            <tr>
                                <td colSpan={3} style={{ textAlign: 'center' }}>
                                    No maintenance has been recorded
                                </td>
                            </tr>
                        ) : (
                            maintenanceRecords.map((record) => (
                                <tr
                                    key={record.id}
                                    onClick={() => handleRowClick(record)}
                                    style={{ cursor: 'pointer' }}
                                >
                                    <td>{record.description}</td>
                                    <td>{record.cost}</td>
                                    <td>{record.date}</td>
                                </tr>
                            ))
                        )}
                    </tbody>
                </Table>
                <div
                    style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        marginTop: '10px',
                        fontWeight: 'bold',
                    }}
                >
                    <i
                        className="bi bi-plus-square-fill"
                        style={{ fontSize: '2rem', cursor: 'pointer' }}
                        onClick={handleShowAddModal}
                    ></i>
                    <div>Total Maintenance Cost: ${totalCost}</div>
                </div>
            </Card.Body>
            <AddMaintenanceModal
                show={showAddModal}
                handleClose={handleCloseAddModal}
                handleSave={handleSaveMaintenance}
                applianceId={applianceId}
                referenceType={referenceType}
                spaceType={spaceType || MaintenanceSpaceType.NotApplicable}
            />
            {selectedRecord && (
                <ShowMaintenanceModal
                    show={showViewModal}
                    handleClose={handleCloseViewModal}
                    maintenanceRecord={selectedRecord}
                    handleDeleteMaintenance={handleDeleteMaintenance}
                />
            )}
        </Card>
    )
}

export default MaintenanceSection

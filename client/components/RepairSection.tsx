import React, { useEffect, useState } from 'react'
import { Card, Table } from 'react-bootstrap'
import 'bootstrap-icons/font/bootstrap-icons.css'
import { SERVER_URL } from '@/pages/_app'
import AddRepairModal from '@/components/AddRepairModal'
import ShowRepairModal from '@/components/ShowRepairModal'

export enum RepairReferenceType {
    Appliance = 'Appliance',
    Space = 'Space',
}

export enum RepairSpaceType {
    BuildingExterior = 'BuildingExterior',
    BuildingInterior = 'BuildingInterior',
    Electrical = 'Electrical',
    HVAC = 'HVAC',
    Plumbing = 'Plumbing',
    Yard = 'Yard',
    NotApplicable = 'NotApplicable',
}

export interface RepairRecord {
    id: number
    description: string
    date: string
    cost: number
    notes: string
    spaceType: RepairSpaceType
    referenceType: RepairReferenceType
    applianceId: number
    attachmentIds?: number[]
}

interface RepairProps {
    applianceId?: number
    referenceType: RepairReferenceType
    spaceType?: RepairSpaceType
}

const RepairSection: React.FC<RepairProps> = ({ applianceId, referenceType, spaceType }) => {
    const [repairRecords, setRepairRecords] = useState<RepairRecord[]>([])
    const [showAddModal, setShowAddModal] = useState(false)
    const [showViewModal, setShowViewModal] = useState(false)
    const [selectedRecord, setSelectedRecord] = useState<RepairRecord | null>(null)

    useEffect(() => {
        const fetchRepairRecords = async () => {
            try {
                const queryParams = new URLSearchParams({
                    applianceId: applianceId?.toString() || '',
                    referenceType,
                    spaceType: spaceType || '',
                }).toString()

                const response = await fetch(`${SERVER_URL}/repair?${queryParams}`)
                const data = await response.json()
                setRepairRecords(data)
            } catch (error) {
                console.error('Error fetching repair records:', error)
            }
        }

        fetchRepairRecords()
    }, [applianceId, referenceType, spaceType])

    const handleShowAddModal = () => setShowAddModal(true)
    const handleCloseAddModal = () => setShowAddModal(false)
    const handleSaveRepair = (newRepair: RepairRecord) => {
        setRepairRecords([...repairRecords, newRepair])
    }

    const handleRowClick = (record: RepairRecord) => {
        setSelectedRecord(record)
        setShowViewModal(true)
    }

    const handleCloseViewModal = () => setShowViewModal(false)

    const handleDeleteRepair = (id: number) => {
        setRepairRecords(repairRecords.filter((record) => record.id !== id))
    }

    const totalCost = repairRecords.reduce((sum, record) => sum + record.cost, 0)

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
                        {repairRecords.length === 0 ? (
                            <tr>
                                <td colSpan={3} style={{ textAlign: 'center' }}>
                                    No repairs have been recorded
                                </td>
                            </tr>
                        ) : (
                            repairRecords.map((record) => (
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
                    <div>Total Repair Cost: ${totalCost}</div>
                </div>
            </Card.Body>
            <AddRepairModal
                show={showAddModal}
                handleClose={handleCloseAddModal}
                handleSave={handleSaveRepair}
                applianceId={applianceId}
                referenceType={referenceType}
                spaceType={spaceType || RepairSpaceType.NotApplicable}
            />
            {selectedRecord && (
                <ShowRepairModal
                    show={showViewModal}
                    handleClose={handleCloseViewModal}
                    repairRecord={selectedRecord}
                    handleDeleteRepair={handleDeleteRepair}
                />
            )}
        </Card>
    )
}

export default RepairSection

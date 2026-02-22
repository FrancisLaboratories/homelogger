import React from 'react'
import { Card } from 'react-bootstrap'

interface ApplianceCardProps {
    id: number
    applianceName: string
    manufacturer: string
    modelNumber: string
    serialNumber: string
    yearPurchased: string
    purchasePrice: string
    location: string
    type: string
    onClick?: () => void // Add optional onClick prop
}

const ApplianceCard: React.FC<ApplianceCardProps> = ({
    id,
    applianceName,
    manufacturer,
    modelNumber,
    serialNumber,
    yearPurchased,
    purchasePrice,
    location,
    type,
    onClick,
}) => {
    // If onClick is provided, use Card with click handler; otherwise, fallback to Link
    if (onClick) {
        return (
            <Card
                style={{
                    width: '18rem',
                    margin: '0',
                    padding: '0',
                    cursor: 'pointer',
                }}
                onClick={onClick}
            >
                <Card.Body>
                    <Card.Title>{applianceName}</Card.Title>
                    <Card.Subtitle className="mb-2 text-muted">{type}</Card.Subtitle>
                    <Card.Text>
                        <strong>Location:</strong> {location}
                    </Card.Text>
                </Card.Body>
            </Card>
        )
    }

    return (
        <a style={{ textDecoration: 'none' }} href={`/appliance.html?id=${id}`}>
            <Card
                style={{
                    width: '18rem',
                    margin: '0',
                    padding: '0',
                    cursor: 'pointer',
                }}
            >
                <Card.Body>
                    <Card.Title>{applianceName}</Card.Title>
                    <Card.Subtitle className="mb-2 text-muted">{type}</Card.Subtitle>
                    <Card.Text>
                        <strong>Location:</strong> {location}
                    </Card.Text>
                </Card.Body>
            </Card>
        </a>
    )
}

export default ApplianceCard

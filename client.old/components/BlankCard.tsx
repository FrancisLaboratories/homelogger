import React from 'react'
import { Card } from 'react-bootstrap'

interface BlankCardProps {
    onClick: () => void
}

const BlankCard: React.FC<BlankCardProps> = ({ onClick }) => {
    return (
        <Card
            style={{
                width: '18rem',
                margin: '0',
                padding: '0',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
            }}
            onClick={onClick}
        >
            <Card.Body>
                <Card.Title style={{ fontSize: '2rem', color: '#007bff' }}>+</Card.Title>
            </Card.Body>
        </Card>
    )
}

export default BlankCard

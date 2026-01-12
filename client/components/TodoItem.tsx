import Form from 'react-bootstrap/Form';
import ListGroup from 'react-bootstrap/ListGroup';
import { useState, useEffect } from 'react';
import { SERVER_URL } from "@/pages/_app";

interface TodoItemProps {
    id: string;
    label: string;
    checked: boolean;
    onDelete: (id: string) => void;
    onToggle?: (id: string, checked: boolean) => void;
    applianceId?: number;
    spaceType?: string;
    sourceLabel?: string;
    createdAt?: string;
}

const TodoItem: React.FC<TodoItemProps> = ({ id, label, checked, onDelete, onToggle, applianceId, spaceType, sourceLabel, createdAt }) => {
    const [isChecked, setIsChecked] = useState<boolean>(checked);

    useEffect(() => {
        setIsChecked(checked);
    }, [checked]);

    const handleCheckboxChange = async () => {
        const newChecked = !isChecked;
        setIsChecked(newChecked);
        if (onToggle) onToggle(id, newChecked);

        try {
            const response = await fetch(SERVER_URL + `/todo/update/${id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ checked: newChecked }),
            });

            if (!response.ok) {
                throw new Error('Failed to update todo');
            }
        } catch (error) {
            console.error('Error updating todo:', error);
        }
    };

    const handleDelete = async () => {
        const confirmDelete = confirm('Are you sure you want to delete this?');
        if (confirmDelete) {
            try {
                const response = await fetch(SERVER_URL + `/todo/delete/${id}`, {
                    method: 'DELETE',
                });

                if (!response.ok) {
                    throw new Error('Failed to delete todo');
                }

                onDelete(id);
            } catch (error) {
                console.error('Error deleting todo:', error);
            }
        }
    };

    const prettySpace = (s?: string | null) => {
        if (!s) return null;
        switch (s) {
            case 'BuildingExterior': return 'Building Exterior';
            case 'BuildingInterior': return 'Building Interior';
            case 'Electrical': return 'Electrical';
            case 'HVAC': return 'HVAC';
            case 'Plumbing': return 'Plumbing';
            case 'Yard': return 'Yard';
            default:
                // Fallback: split camel case or return raw
                return s.replace(/([a-z])([A-Z])/g, '$1 $2');
        }
    };

    let sourceText: string;
    if (applianceId) {
        sourceText = `Appliance: ${sourceLabel || `Appliance ${applianceId}`}`;
    } else if (spaceType) {
        const pretty = prettySpace(spaceType);
        sourceText = pretty ? `${pretty}` : `${spaceType}`;
    } else if (sourceLabel) {
        sourceText = sourceLabel;
    } else {
        sourceText = 'General';
    }

    // Format createdAt if present
    let createdAtText: string | null = null;
    if (createdAt) {
        try {
            const d = new Date(createdAt);
            if (!isNaN(d.getTime())) createdAtText = d.toLocaleString();
        } catch (e) {
            createdAtText = null;
        }
    }

    return (
        <ListGroup.Item className="d-flex justify-content-between align-items-center">
            <div style={{ flex: 1 }}>
                <Form.Check
                    type='checkbox'
                    label={
                        <>
                            <div>{label}</div>
                            <div className="text-muted" style={{ fontSize: '0.75rem' }}>{sourceText}</div>
                            {createdAtText && (
                                <div className="text-muted" style={{ fontSize: '0.7rem' }}>{createdAtText}</div>
                            )}
                        </>
                    }
                    checked={isChecked}
                    onChange={handleCheckboxChange}
                />
            </div>
            <span style={{ cursor: 'pointer', marginLeft: '8px' }}>
                <i className="bi bi-trash" onClick={handleDelete}></i>
            </span>
        </ListGroup.Item>
    );
};

export default TodoItem;
import React, { useState, useEffect } from 'react'
import { Button, Form, Modal } from 'react-bootstrap'
import { SERVER_URL } from '@/pages/_app'
import { Task } from './TasksSection'

interface CompleteTaskModalProps {
    show: boolean
    onHide: () => void
    task: Task | null
    onComplete: (updatedTask: Task) => void
}

const CompleteTaskModal: React.FC<CompleteTaskModalProps> = ({
    show,
    onHide,
    task,
    onComplete,
}) => {
    const today = new Date().toISOString().split('T')[0]
    const [completionDate, setCompletionDate] = useState(today)
    const [recordType, setRecordType] = useState<'none' | 'maintenance' | 'repair'>('none')
    const [description, setDescription] = useState('')
    const [cost, setCost] = useState('')
    const [submitting, setSubmitting] = useState(false)
    const [error, setError] = useState('')

    useEffect(() => {
        if (show && task) {
            setCompletionDate(today)
            setRecordType('none')
            setDescription(task.label)
            setCost(task.estimatedCost == null ? '' : String(task.estimatedCost))
            setError('')
        }
    }, [show, task])

    const handleConfirm = async () => {
        if (task == null) return
        setError('')
        setSubmitting(true)

        try {
            const body: Record<string, unknown> = {
                completionDate,
                createRecord: recordType !== 'none',
            }
            if (recordType !== 'none') {
                body.recordType = recordType
                body.description = description
                body.cost = cost === '' ? 0 : Number(cost)
            }

            const res = await fetch(`${SERVER_URL}/task/complete/${task.id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body),
            })
            if (!res.ok) throw new Error(await res.text())
            const updated: Task = await res.json()
            onComplete(updated)
            onHide()
        } catch (e) {
            setError(String(e))
        } finally {
            setSubmitting(false)
        }
    }

    if (!task) return null

    return (
        <Modal show={show} onHide={onHide}>
            <Modal.Header closeButton>
                <Modal.Title>Complete Task</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                {error && <div className="alert alert-danger py-2">{error}</div>}

                <p className="mb-1">
                    <strong>{task.label}</strong>
                </p>

                <Form.Group className="mb-3">
                    <Form.Label>Completion date</Form.Label>
                    <Form.Control
                        type="date"
                        value={completionDate}
                        onChange={(e) => setCompletionDate(e.target.value)}
                    />
                </Form.Group>

                {task.isRecurring && (
                    <div className="alert alert-info py-2 mb-3" style={{ fontSize: '0.875rem' }}>
                        This is a recurring task. The next due date will be calculated
                        automatically.
                    </div>
                )}

                <Form.Group className="mb-3">
                    <Form.Label>Log this as a record?</Form.Label>
                    <div>
                        <Form.Check
                            inline
                            type="radio"
                            label="No log"
                            id="record-none"
                            checked={recordType === 'none'}
                            onChange={() => setRecordType('none')}
                        />
                        <Form.Check
                            inline
                            type="radio"
                            label="Maintenance"
                            id="record-maintenance"
                            checked={recordType === 'maintenance'}
                            onChange={() => setRecordType('maintenance')}
                        />
                        <Form.Check
                            inline
                            type="radio"
                            label="Repair"
                            id="record-repair"
                            checked={recordType === 'repair'}
                            onChange={() => setRecordType('repair')}
                        />
                    </div>
                </Form.Group>

                {recordType !== 'none' && (
                    <div className="border rounded p-3">
                        <Form.Group className="mb-2">
                            <Form.Label>Description</Form.Label>
                            <Form.Control
                                type="text"
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                            />
                        </Form.Group>
                        <Form.Group>
                            <Form.Label>Cost ($)</Form.Label>
                            <Form.Control
                                type="number"
                                min="0"
                                step="0.01"
                                value={cost}
                                onChange={(e) => setCost(e.target.value)}
                                placeholder="0"
                            />
                        </Form.Group>
                    </div>
                )}
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={onHide}>
                    Cancel
                </Button>
                <Button variant="success" onClick={handleConfirm} disabled={submitting}>
                    {submitting ? 'Saving…' : 'Mark Complete'}
                </Button>
            </Modal.Footer>
        </Modal>
    )
}

export default CompleteTaskModal

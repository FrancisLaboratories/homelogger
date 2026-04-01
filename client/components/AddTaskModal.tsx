import React, { useState, useEffect } from 'react'
import { Button, Form, Modal } from 'react-bootstrap'
import { SERVER_URL } from '@/pages/_app'
import { Task } from './TasksSection'

interface AddTaskModalProps {
    show: boolean
    onHide: () => void
    onAdd: (task: Task) => void
    applianceId?: number
    spaceType?: string
}

const PRIORITY_OPTIONS = ['', 'low', 'medium', 'high', 'critical']
const UNIT_OPTIONS = ['days', 'weeks', 'months', 'years']

const AddTaskModal: React.FC<AddTaskModalProps> = ({
    show,
    onHide,
    onAdd,
    applianceId,
    spaceType,
}) => {
    const [quickMode, setQuickMode] = useState(true)
    const [label, setLabel] = useState('')
    const [notes, setNotes] = useState('')
    const [priority, setPriority] = useState('')
    const [dueDate, setDueDate] = useState('')
    const [estimatedCost, setEstimatedCost] = useState('')
    const [isRecurring, setIsRecurring] = useState(false)
    const [recurrenceInterval, setRecurrenceInterval] = useState(1)
    const [recurrenceUnit, setRecurrenceUnit] = useState('months')
    const [recurrenceMode, setRecurrenceMode] = useState('completion_date')
    const [errors, setErrors] = useState<string[]>([])
    const [submitting, setSubmitting] = useState(false)

    useEffect(() => {
        if (!show) {
            setLabel('')
            setNotes('')
            setPriority('')
            setDueDate('')
            setEstimatedCost('')
            setIsRecurring(false)
            setRecurrenceInterval(1)
            setRecurrenceUnit('months')
            setRecurrenceMode('completion_date')
            setErrors([])
            setQuickMode(true)
        }
    }, [show])

    const handleSubmit = async () => {
        const errs: string[] = []
        if (!label.trim()) errs.push('Label is required')
        if (estimatedCost !== '' && isNaN(Number(estimatedCost)))
            errs.push('Estimated cost must be a number')
        if (isRecurring && recurrenceInterval < 1)
            errs.push('Recurrence interval must be at least 1')
        if (errs.length > 0) {
            setErrors(errs)
            return
        }

        setSubmitting(true)
        try {
            const body: Record<string, unknown> = {
                label: label.trim(),
                notes,
                priority,
                dueDate: dueDate || null,
                estimatedCost: estimatedCost !== '' ? Number(estimatedCost) : null,
                isRecurring,
                recurrenceInterval: isRecurring ? recurrenceInterval : 0,
                recurrenceUnit: isRecurring ? recurrenceUnit : '',
                recurrenceMode: isRecurring ? recurrenceMode : '',
            }
            if (applianceId) body.applianceId = applianceId
            if (spaceType) body.spaceType = spaceType

            const res = await fetch(`${SERVER_URL}/task/add`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body),
            })
            if (!res.ok) throw new Error(await res.text())
            const created: Task = await res.json()
            onAdd(created)
            onHide()
        } catch (e) {
            setErrors([String(e)])
        } finally {
            setSubmitting(false)
        }
    }

    return (
        <Modal show={show} onHide={onHide}>
            <Modal.Header closeButton>
                <Modal.Title>Add Task</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <div className="d-flex align-items-center gap-2 mb-3">
                    <span className="text-muted" style={{ fontSize: '0.9rem' }}>
                        Quick
                    </span>
                    <Form.Check
                        type="switch"
                        id="addtask-mode-switch"
                        checked={!quickMode}
                        onChange={(e) => setQuickMode(!e.target.checked)}
                        label=""
                    />
                    <span className="text-muted" style={{ fontSize: '0.9rem' }}>
                        Detailed
                    </span>
                </div>

                {errors.length > 0 && (
                    <div className="alert alert-danger py-2">
                        {errors.map((e, i) => (
                            <div key={i}>{e}</div>
                        ))}
                    </div>
                )}

                <Form.Group className="mb-3">
                    <Form.Label>
                        Label <span className="text-danger">*</span>
                    </Form.Label>
                    <Form.Control
                        type="text"
                        value={label}
                        onChange={(e) => setLabel(e.target.value)}
                        placeholder="What needs to be done?"
                        autoFocus
                    />
                </Form.Group>

                {!quickMode && (
                    <>
                        <Form.Group className="mb-3">
                            <Form.Label>Notes</Form.Label>
                            <Form.Control
                                as="textarea"
                                rows={2}
                                value={notes}
                                onChange={(e) => setNotes(e.target.value)}
                                placeholder="Optional details"
                            />
                        </Form.Group>

                        <div className="row">
                            <Form.Group className="mb-3 col-6">
                                <Form.Label>Priority</Form.Label>
                                <Form.Select
                                    value={priority}
                                    onChange={(e) => setPriority(e.target.value)}
                                >
                                    <option value="">None</option>
                                    <option value="low">Low</option>
                                    <option value="medium">Medium</option>
                                    <option value="high">High</option>
                                    <option value="critical">Critical</option>
                                </Form.Select>
                            </Form.Group>

                            <Form.Group className="mb-3 col-6">
                                <Form.Label>Due Date</Form.Label>
                                <Form.Control
                                    type="date"
                                    value={dueDate}
                                    onChange={(e) => setDueDate(e.target.value)}
                                />
                            </Form.Group>
                        </div>

                        <Form.Group className="mb-3">
                            <Form.Label>Estimated Cost ($)</Form.Label>
                            <Form.Control
                                type="number"
                                min="0"
                                step="0.01"
                                value={estimatedCost}
                                onChange={(e) => setEstimatedCost(e.target.value)}
                                placeholder="Optional"
                            />
                        </Form.Group>

                        <Form.Group className="mb-3">
                            <Form.Check
                                type="switch"
                                id="recurring-switch"
                                label="Recurring task"
                                checked={isRecurring}
                                onChange={(e) => setIsRecurring(e.target.checked)}
                            />
                        </Form.Group>

                        {isRecurring && (
                            <div className="border rounded p-3 mb-3">
                                <div className="row">
                                    <Form.Group className="mb-2 col-4">
                                        <Form.Label>Every</Form.Label>
                                        <Form.Control
                                            type="number"
                                            min="1"
                                            value={recurrenceInterval}
                                            onChange={(e) =>
                                                setRecurrenceInterval(Number(e.target.value))
                                            }
                                        />
                                    </Form.Group>
                                    <Form.Group className="mb-2 col-8">
                                        <Form.Label>Unit</Form.Label>
                                        <Form.Select
                                            value={recurrenceUnit}
                                            onChange={(e) => setRecurrenceUnit(e.target.value)}
                                        >
                                            {UNIT_OPTIONS.map((u) => (
                                                <option key={u} value={u}>
                                                    {u.charAt(0).toUpperCase() + u.slice(1)}
                                                </option>
                                            ))}
                                        </Form.Select>
                                    </Form.Group>
                                </div>
                                <Form.Group>
                                    <Form.Label>Schedule next occurrence based on</Form.Label>
                                    <Form.Select
                                        value={recurrenceMode}
                                        onChange={(e) => setRecurrenceMode(e.target.value)}
                                    >
                                        <option value="completion_date">Completion date</option>
                                        <option value="due_date">Original due date</option>
                                    </Form.Select>
                                </Form.Group>
                            </div>
                        )}
                    </>
                )}
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={onHide}>
                    Cancel
                </Button>
                <Button variant="primary" onClick={handleSubmit} disabled={submitting}>
                    {submitting ? 'Adding…' : 'Add Task'}
                </Button>
            </Modal.Footer>
        </Modal>
    )
}

export default AddTaskModal
// re-export ignored options reference to silence unused import warning
export { PRIORITY_OPTIONS, UNIT_OPTIONS }

import React, { useState, useEffect } from 'react'
import { Button, Modal } from 'react-bootstrap'
import { SERVER_URL } from '@/pages/_app'
import { Task } from './TasksSection'
import TaskForm, { PRIORITY_OPTIONS, UNIT_OPTIONS } from './TaskForm'

interface AddTaskModalProps {
    show: boolean
    onHide: () => void
    onAdd: (task: Task) => void
    applianceId?: number
    spaceType?: string
    startDetailed?: boolean
}

const PRIORITY_OPTIONS = ['', 'low', 'medium', 'high', 'critical']
const UNIT_OPTIONS = ['days', 'weeks', 'months', 'years']

const AddTaskModal: React.FC<AddTaskModalProps> = ({
    show,
    onHide,
    onAdd,
    applianceId,
    spaceType,
    startDetailed = false,
}) => {
    const [quickMode, setQuickMode] = useState(!startDetailed)
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
        if (show) {
            setQuickMode(!startDetailed)
        } else {
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
            setQuickMode(!startDetailed)
        }
    }, [show])

    const handleSubmit = async () => {
        const errs: string[] = []
        if (!label.trim()) errs.push('Label is required')
        if (estimatedCost !== '' && Number.isNaN(Number(estimatedCost)))
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
                estimatedCost: estimatedCost === '' ? null : Number(estimatedCost),
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
                <TaskForm
                    label={label}
                    setLabel={setLabel}
                    notes={notes}
                    setNotes={setNotes}
                    priority={priority}
                    setPriority={setPriority}
                    dueDate={dueDate}
                    setDueDate={setDueDate}
                    estimatedCost={estimatedCost}
                    setEstimatedCost={setEstimatedCost}
                    isRecurring={isRecurring}
                    setIsRecurring={setIsRecurring}
                    recurrenceInterval={recurrenceInterval}
                    setRecurrenceInterval={setRecurrenceInterval}
                    recurrenceUnit={recurrenceUnit}
                    setRecurrenceUnit={setRecurrenceUnit}
                    recurrenceMode={recurrenceMode}
                    setRecurrenceMode={setRecurrenceMode}
                    errors={errors}
                    autoFocus
                    showQuickToggle
                    quickMode={quickMode}
                    setQuickMode={setQuickMode}
                />
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

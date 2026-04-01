import React from 'react'
import { Form } from 'react-bootstrap'

export const PRIORITY_OPTIONS = ['', 'low', 'medium', 'high', 'critical']
export const UNIT_OPTIONS = ['days', 'weeks', 'months', 'years']

interface TaskFormProps {
    label: string
    setLabel: (s: string) => void
    notes: string
    setNotes: (s: string) => void
    priority: string
    setPriority: (s: string) => void
    dueDate: string
    setDueDate: (s: string) => void
    estimatedCost: string
    setEstimatedCost: (s: string) => void
    isRecurring: boolean
    setIsRecurring: (b: boolean) => void
    recurrenceInterval: number
    setRecurrenceInterval: (n: number) => void
    recurrenceUnit: string
    setRecurrenceUnit: (s: string) => void
    recurrenceMode: string
    setRecurrenceMode: (s: string) => void
    errors: string[]
    autoFocus?: boolean
    showQuickToggle?: boolean
    quickMode?: boolean
    setQuickMode?: (b: boolean) => void
}

const TaskForm: React.FC<TaskFormProps> = ({
    label,
    setLabel,
    notes,
    setNotes,
    priority,
    setPriority,
    dueDate,
    setDueDate,
    estimatedCost,
    setEstimatedCost,
    isRecurring,
    setIsRecurring,
    recurrenceInterval,
    setRecurrenceInterval,
    recurrenceUnit,
    setRecurrenceUnit,
    recurrenceMode,
    setRecurrenceMode,
    errors,
    autoFocus = false,
    showQuickToggle = false,
    quickMode = false,
    setQuickMode,
}) => {
    return (
        <>
            {showQuickToggle && (
                <div className="d-flex align-items-center gap-2 mb-3">
                    <span className="text-muted" style={{ fontSize: '0.9rem' }}>
                        Quick
                    </span>
                    <Form.Check
                        type="switch"
                        id="taskform-mode-switch"
                        checked={!quickMode}
                        onChange={(e) => setQuickMode && setQuickMode(!e.target.checked)}
                        label=""
                    />
                    <span className="text-muted" style={{ fontSize: '0.9rem' }}>
                        Detailed
                    </span>
                </div>
            )}

            {errors.length > 0 && (
                <div className="alert alert-danger py-2">
                    {errors.map((e) => (
                        <div key={e}>{e}</div>
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
                    autoFocus={autoFocus}
                />
            </Form.Group>

            {!showQuickToggle || !quickMode ? (
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
            ) : null}
        </>
    )
}

export default TaskForm

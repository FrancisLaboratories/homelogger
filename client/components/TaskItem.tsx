import React, { useState } from 'react'
import { Badge, Button, ListGroup } from 'react-bootstrap'
import { SERVER_URL } from '@/pages/_app'
import { Task } from './TasksSection'
import CompleteTaskModal from './CompleteTaskModal'
import EditTaskModal from './EditTaskModal'

interface TaskItemProps {
    task: Task
    onComplete: (updatedTask: Task) => void
    onDelete: (id: number) => void
    onEdit: (updatedTask: Task) => void
    /** When true, renders source label (appliance name or space type) */
    showSource?: boolean
    /** Human-readable source label (e.g. appliance name or pretty space name) */
    sourceLabel?: string
    /** URL to navigate to the source page */
    sourceHref?: string
}

const priorityBadge: Record<string, { bg: string; label: string }> = {
    low: { bg: 'success', label: 'Low' },
    medium: { bg: 'info', label: 'Medium' },
    high: { bg: 'warning', label: 'High' },
    critical: { bg: 'danger', label: 'Critical' },
}

function dueDateStyle(dueDate: string | null | undefined): { text: string; className: string } | null {
    if (!dueDate) return null
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const due = new Date(dueDate + 'T00:00:00')
    const diffDays = Math.ceil((due.getTime() - today.getTime()) / (1000 * 60 * 60 * 24))
    if (diffDays < 0) return { text: `Overdue (${dueDate})`, className: 'text-danger fw-semibold' }
    if (diffDays === 0) return { text: 'Due today', className: 'text-danger fw-semibold' }
    if (diffDays <= 7) return { text: `Due ${dueDate}`, className: 'text-warning fw-semibold' }
    return { text: `Due ${dueDate}`, className: 'text-muted' }
}

function recurrenceLabel(task: Task): string {
    if (!task.isRecurring || !task.recurrenceInterval || !task.recurrenceUnit) return ''
    const unit = task.recurrenceInterval === 1
        ? task.recurrenceUnit.replace(/s$/, '')
        : task.recurrenceUnit
    return `Every ${task.recurrenceInterval} ${unit}`
}

const TaskItem: React.FC<TaskItemProps> = ({
    task,
    onComplete,
    onDelete,
    onEdit,
    showSource,
    sourceLabel,
    sourceHref,
}) => {
    const [showCompleteModal, setShowCompleteModal] = useState(false)
    const [showEditModal, setShowEditModal] = useState(false)

    const handleDelete = async () => {
        if (!confirm('Delete this task?')) return
        try {
            const res = await fetch(`${SERVER_URL}/task/delete/${task.id}`, { method: 'DELETE' })
            if (!res.ok) throw new Error(await res.text())
            onDelete(task.id)
        } catch (e) {
            console.error('Error deleting task:', e)
        }
    }

    const handleUncomplete = async () => {
        try {
            const res = await fetch(`${SERVER_URL}/task/uncomplete/${task.id}`, { method: 'PUT' })
            if (!res.ok) throw new Error(await res.text())
            const updated: Task = await res.json()
            onEdit(updated)
        } catch (e) {
            console.error('Error uncompleting task:', e)
        }
    }

    const dueStyle = dueDateStyle(task.dueDate)
    const badge = task.priority ? priorityBadge[task.priority] : null
    const recurrence = recurrenceLabel(task)

    return (
        <>
            <ListGroup.Item
                className="d-flex align-items-start gap-2 py-2"
                style={{ opacity: task.checked ? 0.6 : 1 }}
            >
                {/* Checkbox */}
                <div className="pt-1">
                    <input
                        type="checkbox"
                        checked={task.checked}
                        onChange={() => {
                            if (task.checked) handleUncomplete()
                            else setShowCompleteModal(true)
                        }}
                        style={{ cursor: 'pointer', width: '16px', height: '16px' }}
                    />
                </div>

                {/* Main content */}
                <div className="flex-grow-1" style={{ minWidth: 0 }}>
                    <div className="d-flex flex-wrap align-items-center gap-1">
                        <span style={{ textDecoration: task.checked ? 'line-through' : 'none' }}>
                            {task.label}
                        </span>
                        {badge && (
                            <Badge bg={badge.bg} text={badge.bg === 'warning' ? 'dark' : undefined}>
                                {badge.label}
                            </Badge>
                        )}
                        {recurrence && (
                            <Badge bg="secondary" title="Recurring">
                                <i className="bi bi-arrow-repeat me-1" />
                                {recurrence}
                            </Badge>
                        )}
                    </div>

                    <div className="d-flex flex-wrap gap-2 mt-1" style={{ fontSize: '0.8rem' }}>
                        {dueStyle && (
                            <span className={dueStyle.className}>
                                <i className="bi bi-calendar3 me-1" />
                                {dueStyle.text}
                            </span>
                        )}
                        {task.estimatedCost != null && (
                            <span className="text-muted">
                                <i className="bi bi-currency-dollar me-1" />
                                Est. ${task.estimatedCost.toFixed(2)}
                            </span>
                        )}
                        {task.lastCompletedAt && (
                            <span className="text-muted">
                                Last done: {task.lastCompletedAt}
                            </span>
                        )}
                        {showSource && sourceLabel && (
                            <span className="text-muted">
                                {sourceHref ? (
                                    <a href={sourceHref} style={{ textDecoration: 'none', color: 'inherit' }}>
                                        {sourceLabel}
                                    </a>
                                ) : sourceLabel}
                            </span>
                        )}
                    </div>

                    {task.notes && (
                        <div className="text-muted mt-1" style={{ fontSize: '0.8rem', fontStyle: 'italic' }}>
                            {task.notes}
                        </div>
                    )}
                </div>

                {/* Actions */}
                <div className="d-flex gap-1 align-items-start pt-1">
                    <Button
                        variant="link"
                        size="sm"
                        className="text-secondary p-0"
                        onClick={() => setShowEditModal(true)}
                        title="Edit"
                    >
                        <i className="bi bi-pencil" />
                    </Button>
                    <Button
                        variant="link"
                        size="sm"
                        className="text-danger p-0"
                        onClick={handleDelete}
                        title="Delete"
                    >
                        <i className="bi bi-trash" />
                    </Button>
                </div>
            </ListGroup.Item>

            <CompleteTaskModal
                show={showCompleteModal}
                onHide={() => setShowCompleteModal(false)}
                task={task}
                onComplete={onComplete}
            />
            <EditTaskModal
                show={showEditModal}
                onHide={() => setShowEditModal(false)}
                task={task}
                onSave={onEdit}
            />
        </>
    )
}

export default TaskItem

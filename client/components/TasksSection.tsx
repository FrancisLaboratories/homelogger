'use client'

import React, { useEffect, useState, useCallback } from 'react'
import { Button, Form, ListGroup } from 'react-bootstrap'
import 'bootstrap-icons/font/bootstrap-icons.css'
import { SERVER_URL } from '@/pages/_app'
import TaskItem from './TaskItem'
import AddTaskModal from './AddTaskModal'

export interface Task {
    id: number
    label: string
    notes?: string
    checked: boolean
    priority?: string
    dueDate?: string | null
    estimatedCost?: number | null
    isRecurring: boolean
    recurrenceInterval?: number
    recurrenceUnit?: string
    recurrenceMode?: string
    lastCompletedAt?: string | null
    userid?: string
    applianceId?: number | null
    spaceType?: string | null
    CreatedAt?: string
    createdAt?: string
}

interface TasksSectionProps {
    applianceId?: number
    spaceType?: string
}

type SortOption = 'due_asc' | 'due_desc' | 'priority' | 'created_desc' | 'label_asc'
type FilterOption = 'upcoming' | 'overdue' | 'all' | 'completed'

const PRIORITY_ORDER: Record<string, number> = {
    critical: 0,
    high: 1,
    medium: 2,
    low: 3,
    '': 4,
}

function parseDue(dueDate?: string | null): number {
    if (!dueDate) return Infinity
    return new Date(dueDate + 'T00:00:00').getTime()
}

function isOverdue(task: Task): boolean {
    if (!task.dueDate || task.checked) return false
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    return new Date(task.dueDate + 'T00:00:00') < today
}

const TasksSection: React.FC<TasksSectionProps> = ({ applianceId, spaceType }) => {
    const [tasks, setTasks] = useState<Task[]>([])
    const [sortOption, setSortOption] = useState<SortOption>('due_asc')
    const [filterOption, setFilterOption] = useState<FilterOption>('upcoming')
    const [quickLabel, setQuickLabel] = useState('')
    const [showAddModal, setShowAddModal] = useState(false)
    const [showAddDetailed, setShowAddDetailed] = useState(false)

    useEffect(() => {
        try {
            const savedSort = localStorage.getItem('homelogger_tasks_sort')
            const savedFilter = localStorage.getItem('homelogger_tasks_filter')
            if (savedSort) setSortOption(savedSort as SortOption)
            if (savedFilter) setFilterOption(savedFilter as FilterOption)
        } catch (_) {
            /* ignore */
        }
    }, [])

    const fetchTasks = useCallback(async () => {
        const params = new URLSearchParams()
        if (applianceId) params.set('applianceId', String(applianceId))
        else if (spaceType) params.set('spaceType', spaceType)
        params.set('includeCompleted', 'true')

        try {
            const res = await fetch(`${SERVER_URL}/task?${params}`)
            const data: Task[] = await res.json()
            setTasks(data)
        } catch (e) {
            console.error('Error fetching tasks:', e)
        }
    }, [applianceId, spaceType])

    useEffect(() => {
        fetchTasks()
    }, [fetchTasks])

    const handleSortChange = (val: SortOption) => {
        setSortOption(val)
        try {
            localStorage.setItem('homelogger_tasks_sort', val)
        } catch (_) {
            /* ignore */
        }
    }

    const handleFilterChange = (val: FilterOption) => {
        setFilterOption(val)
        try {
            localStorage.setItem('homelogger_tasks_filter', val)
        } catch (_) {
            /* ignore */
        }
    }

    const handleQuickAdd = async () => {
        const label = quickLabel.trim()
        if (!label) return
        try {
            const body: Record<string, unknown> = { label }
            if (applianceId) body.applianceId = applianceId
            if (spaceType) body.spaceType = spaceType

            const res = await fetch(`${SERVER_URL}/task/add`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body),
            })
            if (!res.ok) throw new Error(await res.text())
            const created: Task = await res.json()
            setTasks((prev) => [created, ...prev])
            setQuickLabel('')
        } catch (e) {
            console.error('Error adding task:', e)
        }
    }

    const handleComplete = (updated: Task) => {
        setTasks((prev) => prev.map((t) => (t.id === updated.id ? updated : t)))
    }

    const handleDelete = (id: number) => {
        setTasks((prev) => prev.filter((t) => t.id !== id))
    }

    const handleEdit = (updated: Task) => {
        setTasks((prev) => prev.map((t) => (t.id === updated.id ? updated : t)))
    }

    const filtered = tasks.filter((t) => {
        if (filterOption === 'completed') return t.checked
        if (filterOption === 'upcoming') return !t.checked && !isOverdue(t)
        if (filterOption === 'overdue') return !t.checked && isOverdue(t)
        return true // 'all'
    })

    const sorted = [...filtered].sort((a, b) => {
        switch (sortOption) {
            case 'due_asc':
                return parseDue(a.dueDate) - parseDue(b.dueDate)
            case 'due_desc':
                return parseDue(b.dueDate) - parseDue(a.dueDate)
            case 'priority':
                return (
                    (PRIORITY_ORDER[a.priority || ''] ?? 4) -
                    (PRIORITY_ORDER[b.priority || ''] ?? 4)
                )
            case 'label_asc':
                return a.label.localeCompare(b.label)
            case 'created_desc':
            default: {
                const aDate = a.CreatedAt || a.createdAt || ''
                const bDate = b.CreatedAt || b.createdAt || ''
                return bDate.localeCompare(aDate)
            }
        }
    })

    const overdueCount = tasks.filter((t) => !t.checked && isOverdue(t)).length

    return (
        <div style={{ marginTop: '12px' }}>
            {/* Controls */}
            <div className="d-flex flex-wrap gap-2 mb-2 align-items-center">
                <Form.Select
                    size="sm"
                    style={{ width: 'auto' }}
                    value={filterOption}
                    onChange={(e) => handleFilterChange(e.target.value as FilterOption)}
                    aria-label="Filter tasks"
                >
                    <option value="upcoming">Upcoming{overdueCount > 0 ? '' : ''}</option>
                    <option value="overdue">
                        Overdue {overdueCount > 0 ? `(${overdueCount})` : ''}
                    </option>
                    <option value="all">All</option>
                    <option value="completed">Completed</option>
                </Form.Select>

                <Form.Select
                    size="sm"
                    style={{ width: 'auto' }}
                    value={sortOption}
                    onChange={(e) => handleSortChange(e.target.value as SortOption)}
                    aria-label="Sort tasks"
                >
                    <option value="due_asc">Due date ↑</option>
                    <option value="due_desc">Due date ↓</option>
                    <option value="priority">Priority</option>
                    <option value="label_asc">Label A–Z</option>
                    <option value="created_desc">Recently added</option>
                </Form.Select>

                <Button
                    variant="outline-primary"
                    size="sm"
                    className="ms-auto"
                    onClick={() => {
                        setShowAddDetailed(true)
                        setShowAddModal(true)
                    }}
                >
                    <i className="bi bi-plus-lg me-1" />
                    Add Detailed Task
                </Button>
            </div>

            {overdueCount > 0 && filterOption === 'upcoming' && (
                <div
                    className="alert alert-warning py-1 px-2 mb-2"
                    style={{ fontSize: '0.85rem', cursor: 'pointer' }}
                    onClick={() => handleFilterChange('overdue')}
                >
                    <i className="bi bi-exclamation-triangle me-1" />
                    {overdueCount} overdue task{overdueCount !== 1 ? 's' : ''} — click to view
                </div>
            )}

            {/* Task list */}
            {sorted.length === 0 ? (
                <div className="text-muted" style={{ fontSize: '0.9rem', padding: '8px 0' }}>
                    No tasks to show.
                </div>
            ) : (
                <ListGroup className="mb-2">
                    {sorted.map((task) => (
                        <TaskItem
                            key={task.id}
                            task={task}
                            onComplete={handleComplete}
                            onDelete={handleDelete}
                            onEdit={handleEdit}
                        />
                    ))}
                </ListGroup>
            )}

            {/* Quick-add input */}
            <div className="d-flex gap-2 mt-2">
                <Form.Control
                    type="text"
                    size="sm"
                    placeholder="Add a new task..."
                    value={quickLabel}
                    onChange={(e) => setQuickLabel(e.target.value)}
                    onKeyDown={(e) => {
                        if (e.key === 'Enter') handleQuickAdd()
                    }}
                />
                <Button variant="outline-secondary" size="sm" onClick={handleQuickAdd}>
                    <i className="bi bi-plus-lg" />
                </Button>
            </div>

            <AddTaskModal
                show={showAddModal}
                onHide={() => {
                    setShowAddModal(false)
                    setShowAddDetailed(false)
                }}
                onAdd={(task) => setTasks((prev) => [task, ...prev])}
                applianceId={applianceId}
                spaceType={spaceType}
                startDetailed={showAddDetailed}
            />
        </div>
    )
}

export default TasksSection

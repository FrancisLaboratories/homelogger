'use client'

import React, { useEffect, useState, useCallback } from 'react'
import { Button, Form, ListGroup } from 'react-bootstrap'
import 'bootstrap-icons/font/bootstrap-icons.css'
import { SERVER_URL } from '@/pages/_app'
import { Task } from './TasksSection'
import TaskItem from './TaskItem'
import AddTaskModal from './AddTaskModal'

const SPACE_URLS: Record<string, string> = {
    BuildingExterior: '/building-exterior.html',
    BuildingInterior: '/building-interior.html',
    Electrical: '/electrical.html',
    HVAC: '/hvac.html',
    Plumbing: '/plumbing.html',
    Yard: '/yard.html',
}

const SPACE_LABELS: Record<string, string> = {
    BuildingExterior: 'Building Exterior',
    BuildingInterior: 'Building Interior',
    Electrical: 'Electrical',
    HVAC: 'HVAC',
    Plumbing: 'Plumbing',
    Yard: 'Yard',
}

function isoToday(): string {
    return new Date().toISOString().split('T')[0]
}

function addDays(date: Date, days: number): Date {
    const d = new Date(date)
    d.setDate(d.getDate() + days)
    return d
}

function isBeforeToday(dueDate: string): boolean {
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    return new Date(dueDate + 'T00:00:00') < today
}

interface TaskGroup {
    label: string
    tasks: Task[]
    headerClass: string
}

const TasksDashboard: React.FC = () => {
    const [tasks, setTasks] = useState<Task[]>([])
    const [applianceNames, setApplianceNames] = useState<Record<number, string>>({})
    const [quickLabel, setQuickLabel] = useState('')
    const [showAddModal, setShowAddModal] = useState(false)

    const fetchTasks = useCallback(async () => {
        try {
            const res = await fetch(`${SERVER_URL}/task/dashboard`)
            if (!res.ok) return
            const data: Task[] = await res.json()
            setTasks(data)

            // Enrich with appliance names
            const applianceIds = Array.from(
                new Set(data.filter((t) => t.applianceId).map((t) => Number(t.applianceId)))
            )
            const nameMap: Record<number, string> = {}
            await Promise.all(
                applianceIds.map(async (id) => {
                    try {
                        const r = await fetch(`${SERVER_URL}/appliances/${id}`)
                        if (!r.ok) return
                        const a = await r.json()
                        nameMap[id] = a.applianceName || `Appliance ${id}`
                    } catch (err) {
                        console.warn(`Failed to fetch appliance ${id}:`, err)
                    }
                })
            )
            setApplianceNames(nameMap)
        } catch (e) {
            console.error('Error fetching dashboard tasks:', e)
        }
    }, [])

    useEffect(() => {
        fetchTasks()
    }, [fetchTasks])

    const handleComplete = (updated: Task) => {
        setTasks((prev) =>
            prev
                .map((t) => (t.id === updated.id ? updated : t))
                .filter((t) => !t.checked || t.isRecurring)
        )
        // refresh to pick up any new due date for recurring
        fetchTasks()
    }

    const handleDelete = (id: number) => {
        setTasks((prev) => prev.filter((t) => t.id !== id))
    }

    const handleEdit = (updated: Task) => {
        setTasks((prev) => prev.map((t) => (t.id === updated.id ? updated : t)))
    }

    const handleQuickAdd = async () => {
        const label = quickLabel.trim()
        if (!label) return
        try {
            const res = await fetch(`${SERVER_URL}/task/add`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ label }),
            })
            if (!res.ok) throw new Error(await res.text())
            const created: Task = await res.json()
            setTasks((prev) => [...prev, created])
            setQuickLabel('')
        } catch (e) {
            console.error('Error adding task:', e)
        }
    }

    function getSourceLabel(task: Task): string {
        if (task.applianceId != null) {
            return applianceNames[task.applianceId]
                ? `Appliance: ${applianceNames[task.applianceId]}`
                : `Appliance ${task.applianceId}`
        }
        if (task.spaceType) return SPACE_LABELS[task.spaceType] || task.spaceType
        return 'General'
    }

    function getSourceHref(task: Task): string | undefined {
        if (task.applianceId != null) return `/appliance.html?id=${task.applianceId}`
        if (task.spaceType) return SPACE_URLS[task.spaceType]
        return undefined
    }

    // Group tasks by due date
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const in7 = addDays(today, 7)
    const in30 = addDays(today, 30)

    const groups: TaskGroup[] = [
        {
            label: 'Overdue',
            headerClass: 'text-danger',
            tasks: tasks.filter((t) => t.dueDate && isBeforeToday(t.dueDate)),
        },
        {
            label: 'Due Next 7 Days',
            headerClass: 'text-warning',
            tasks: tasks.filter((t) => {
                if (!t.dueDate) return false
                const d = new Date(t.dueDate + 'T00:00:00')
                return d >= today && d < in7
            }),
        },
        {
            label: 'Due Next 30 Days',
            headerClass: 'text-primary',
            tasks: tasks.filter((t) => {
                if (!t.dueDate) return false
                const d = new Date(t.dueDate + 'T00:00:00')
                return d >= in7 && d < in30
            }),
        },
        {
            label: 'Later',
            headerClass: 'text-muted',
            tasks: tasks.filter((t) => {
                if (!t.dueDate) return false
                return new Date(t.dueDate + 'T00:00:00') >= in30
            }),
        },
        {
            label: 'No Due Date',
            headerClass: 'text-muted',
            tasks: tasks.filter((t) => !t.dueDate),
        },
    ]

    const totalActive = tasks.length

    return (
        <div>
            <div className="d-flex align-items-center justify-content-between mb-3">
                <h4 className="mb-0">
                    Tasks
                    {totalActive > 0 && (
                        <span
                            className="text-muted ms-2"
                            style={{ fontSize: '1rem', fontWeight: 'normal' }}
                        >
                            ({totalActive} active)
                        </span>
                    )}
                </h4>
                <Button variant="outline-primary" size="sm" onClick={() => setShowAddModal(true)}>
                    <i className="bi bi-plus-lg me-1" aria-hidden="true" />{' '}
                    Add Detailed Task
                </Button>
            </div>

            {/* Quick-add */}
            <div className="d-flex gap-2 mb-3">
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
                    <i className="bi bi-plus-lg" aria-hidden="true" />
                </Button>
            </div>

            {totalActive === 0 && <div className="text-muted">No active tasks. Great work!</div>}

            {groups.map((group) => {
                if (group.tasks.length === 0) return null
                return (
                    <div key={group.label} className="mb-3">
                        <h6 className={`${group.headerClass} mb-1`} style={{ fontWeight: 600 }}>
                            {group.label}
                            <span
                                className="ms-2 text-muted fw-normal"
                                style={{ fontSize: '0.85rem' }}
                            >
                                ({group.tasks.length})
                            </span>
                        </h6>
                        <ListGroup>
                            {group.tasks.map((task) => (
                                <TaskItem
                                    key={task.id}
                                    task={task}
                                    onComplete={handleComplete}
                                    onDelete={handleDelete}
                                    onEdit={handleEdit}
                                    showSource
                                    sourceLabel={getSourceLabel(task)}
                                    sourceHref={getSourceHref(task)}
                                />
                            ))}
                        </ListGroup>
                    </div>
                )
            })}

            <AddTaskModal
                show={showAddModal}
                onHide={() => setShowAddModal(false)}
                onAdd={(task) => {
                    setTasks((prev) => [...prev, task])
                    setShowAddModal(false)
                }}
            />
        </div>
    )
}

export default TasksDashboard

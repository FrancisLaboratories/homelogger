'use client'

import React, { useEffect, useState, useCallback } from 'react'
import { Button, Form, ListGroup } from 'react-bootstrap'
import 'bootstrap-icons/font/bootstrap-icons.css'
import { SERVER_URL } from '@/pages/_app'
import { Task } from './TasksSection'
import TaskItem from './TaskItem'
import AddTaskModal from './AddTaskModal'

type SortOption = 'due_asc' | 'due_desc' | 'priority' | 'created_desc' | 'label_asc'
type FilterOption = 'active' | 'completed' | 'all' | 'priority_high' | 'recurring' | 'no_date'

const PRIORITY_ORDER: Record<string, number> = {
    critical: 0,
    high: 1,
    medium: 2,
    low: 3,
    '': 4,
}

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

function parseDue(dueDate?: string | null): number {
    if (dueDate == null) return Infinity
    return new Date(dueDate + 'T00:00:00').getTime()
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
    const [sortOption, setSortOption] = useState<SortOption>('due_asc')
    const [filterOption, setFilterOption] = useState<FilterOption>('active')

    const fetchTasks = useCallback(async () => {
        try {
            const res = await fetch(`${SERVER_URL}/task?includeCompleted=true`)
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
        setTasks((prev) => prev.map((t) => (t.id === updated.id ? updated : t)))
        // refresh to pick up any new due date for recurring tasks
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

    function applyFilter(list: Task[]): Task[] {
        let result = list
        // completion status filter
        if (filterOption === 'active') result = result.filter((t) => !t.checked)
        else if (filterOption === 'completed') result = result.filter((t) => t.checked)
        // secondary filters
        else if (filterOption === 'priority_high') result = result.filter((t) => t.priority === 'high' || t.priority === 'critical')
        else if (filterOption === 'recurring') result = result.filter((t) => t.isRecurring)
        else if (filterOption === 'no_date') result = result.filter((t) => !t.dueDate)
        return result
    }

    function applySort(list: Task[]): Task[] {
        return [...list].sort((a, b) => {
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
    }

    // Group tasks by due date
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const in7 = addDays(today, 7)
    const in30 = addDays(today, 30)

    const filtered = applyFilter(tasks)

    const groups: TaskGroup[] = [
        {
            label: 'Overdue',
            headerClass: 'text-danger',
            tasks: applySort(filtered.filter((t) => t.dueDate && isBeforeToday(t.dueDate))),
        },
        {
            label: 'Due Next 7 Days',
            headerClass: 'text-warning',
            tasks: applySort(filtered.filter((t) => {
                if (!t.dueDate) return false
                const d = new Date(t.dueDate + 'T00:00:00')
                return d >= today && d < in7
            })),
        },
        {
            label: 'Due Next 30 Days',
            headerClass: 'text-primary',
            tasks: applySort(filtered.filter((t) => {
                if (!t.dueDate) return false
                const d = new Date(t.dueDate + 'T00:00:00')
                return d >= in7 && d < in30
            })),
        },
        {
            label: 'Later',
            headerClass: 'text-muted',
            tasks: applySort(filtered.filter((t) => {
                if (!t.dueDate) return false
                return new Date(t.dueDate + 'T00:00:00') >= in30
            })),
        },
        {
            label: 'No Due Date',
            headerClass: 'text-muted',
            tasks: applySort(filtered.filter((t) => !t.dueDate)),
        },
    ]

    const totalActive = tasks.filter((t) => !t.checked).length
    const totalFiltered = groups.reduce((sum, g) => sum + g.tasks.length, 0)

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
                    <i className="bi bi-plus-lg me-1" aria-hidden="true" /> Add Detailed Task
                </Button>
            </div>

            {/* Sort / Filter controls */}
            <div className="d-flex flex-wrap gap-2 mb-2 align-items-center">
                <Form.Select
                    size="sm"
                    style={{ width: 'auto' }}
                    value={filterOption}
                    onChange={(e) => setFilterOption(e.target.value as FilterOption)}
                    aria-label="Filter tasks"
                >
                    <option value="active">Active</option>
                    <option value="completed">Completed</option>
                    <option value="all">All</option>
                    <option value="priority_high">High / Critical priority</option>
                    <option value="recurring">Recurring only</option>
                    <option value="no_date">No due date</option>
                </Form.Select>

                <Form.Select
                    size="sm"
                    style={{ width: 'auto' }}
                    value={sortOption}
                    onChange={(e) => setSortOption(e.target.value as SortOption)}
                    aria-label="Sort tasks"
                >
                    <option value="due_asc">Due date ↑</option>
                    <option value="due_desc">Due date ↓</option>
                    <option value="priority">Priority</option>
                    <option value="label_asc">Label A–Z</option>
                    <option value="created_desc">Recently added</option>
                </Form.Select>
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

            {totalFiltered === 0 && (
                <div className="text-muted">
                    {filterOption === 'active' ? 'No active tasks. Great work!' : 'No tasks match the current filter.'}
                </div>
            )}

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

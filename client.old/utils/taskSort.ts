import type { Task } from '../components/TasksSection'

type SortOption = 'due_asc' | 'due_desc' | 'priority' | 'created_desc' | 'label_asc'

const PRIORITY_ORDER: Record<string, number> = {
    critical: 0,
    high: 1,
    medium: 2,
    low: 3,
    '': 4,
}

function parseDue(dueDate?: string | null): number {
    if (dueDate == null) return Infinity
    return new Date(dueDate + 'T00:00:00').getTime()
}

export function applySort(list: Task[], sortOption: SortOption): Task[] {
    console.debug('applySort called', { sortOption, count: list.length })
    const safeLabel = (s?: string | null) => s || ''

    const tieByLabel = (x: Task, y: Task) => safeLabel(x.label).localeCompare(safeLabel(y.label))

    const cmpDueAsc = (x: Task, y: Task) => {
        const av = parseDue(x.dueDate)
        const bv = parseDue(y.dueDate)
        if (av === bv) return tieByLabel(x, y)
        return av < bv ? -1 : 1
    }

    const cmpDueDesc = (x: Task, y: Task) => {
        const avRaw = parseDue(x.dueDate)
        const bvRaw = parseDue(y.dueDate)
        const av = avRaw === Infinity ? -Infinity : avRaw
        const bv = bvRaw === Infinity ? -Infinity : bvRaw
        if (av === bv) return tieByLabel(x, y)
        return av > bv ? -1 : 1
    }

    const cmpPriority = (x: Task, y: Task) => {
        const ap = PRIORITY_ORDER[x.priority || ''] ?? 4
        const bp = PRIORITY_ORDER[y.priority || ''] ?? 4
        if (ap === bp) return tieByLabel(x, y)
        return ap < bp ? -1 : 1
    }

    const cmpLabelAsc = (x: Task, y: Task) => safeLabel(x.label).localeCompare(safeLabel(y.label))

    const cmpCreatedDesc = (x: Task, y: Task) => {
        const aDate = x.CreatedAt || x.createdAt || ''
        const bDate = y.CreatedAt || y.createdAt || ''
        if (aDate === bDate) return tieByLabel(x, y)
        return bDate.localeCompare(aDate)
    }

    const comparator = (() => {
        switch (sortOption) {
            case 'due_asc':
                return cmpDueAsc
            case 'due_desc':
                return cmpDueDesc
            case 'priority':
                return cmpPriority
            case 'label_asc':
                return cmpLabelAsc
            case 'created_desc':
            default:
                return cmpCreatedDesc
        }
    })()

    return [...list].sort(comparator)
}

export default applySort

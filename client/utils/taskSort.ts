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
    return [...list].sort((a, b) => {
        switch (sortOption) {
            case 'due_asc': {
                const av = parseDue(a.dueDate)
                const bv = parseDue(b.dueDate)
                if (av === bv) return safeLabel(a.label).localeCompare(safeLabel(b.label))
                return av < bv ? -1 : 1
            }
            case 'due_desc': {
                // Ensure items with no due date are treated as "latest" and come last
                const avRaw = parseDue(a.dueDate)
                const bvRaw = parseDue(b.dueDate)
                const av = avRaw === Infinity ? -Infinity : avRaw
                const bv = bvRaw === Infinity ? -Infinity : bvRaw
                if (av === bv) return safeLabel(a.label).localeCompare(safeLabel(b.label))
                return av > bv ? -1 : 1
            }
            case 'priority': {
                const ap = PRIORITY_ORDER[a.priority || ''] ?? 4
                const bp = PRIORITY_ORDER[b.priority || ''] ?? 4
                if (ap === bp) return safeLabel(a.label).localeCompare(safeLabel(b.label))
                return ap < bp ? -1 : 1
            }
            case 'label_asc':
                return safeLabel(a.label).localeCompare(safeLabel(b.label))
            case 'created_desc':
            default: {
                const aDate = a.CreatedAt || a.createdAt || ''
                const bDate = b.CreatedAt || b.createdAt || ''
                if (aDate === bDate) return safeLabel(a.label).localeCompare(safeLabel(b.label))
                return bDate.localeCompare(aDate)
            }
        }
    })
}

export default applySort

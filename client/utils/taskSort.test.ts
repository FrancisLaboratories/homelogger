import { describe, it, expect } from 'vitest'
import { applySort } from './taskSort'

type T = any

const sample: T[] = [
    { id: 1, label: 'B task', dueDate: '2026-06-01', priority: 'low', CreatedAt: '2026-01-01' },
    { id: 2, label: 'A task', dueDate: '2026-05-01', priority: 'high', CreatedAt: '2026-02-01' },
    { id: 3, label: 'C task', dueDate: null, priority: '', CreatedAt: '2026-03-01' },
]

describe('applySort', () => {
    it('sorts by due_asc with tie-breaker on label', () => {
        const out = applySort(sample, 'due_asc')
        expect(out.map((t) => t.id)).toEqual([2, 1, 3])
    })

    it('sorts by due_desc with tie-breaker on label', () => {
        const out = applySort(sample, 'due_desc')
        expect(out.map((t) => t.id)).toEqual([1, 2, 3])
    })

    it('sorts by label_asc', () => {
        const out = applySort(sample, 'label_asc')
        expect(out.map((t) => t.id)).toEqual([2, 1, 3])
    })

    it('sorts by priority', () => {
        const out = applySort(sample, 'priority')
        // high (id 2) should come before low (id 1) and '' (id 3)
        expect(out.map((t) => t.id)).toEqual([2, 1, 3])
    })
})

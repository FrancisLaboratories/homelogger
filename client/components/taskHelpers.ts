export interface TaskFormValues {
    label: string
    notes: string
    priority: string
    dueDate: string
    estimatedCost: string
    isRecurring: boolean
    recurrenceInterval: number
    recurrenceUnit: string
    recurrenceMode: string
}

export function validateTaskValues(vals: TaskFormValues): string[] {
    const errs: string[] = []
    if (!vals.label.trim()) errs.push('Label is required')
    if (vals.estimatedCost !== '' && Number.isNaN(Number(vals.estimatedCost)))
        errs.push('Estimated cost must be a number')
    if (vals.isRecurring && vals.recurrenceInterval < 1)
        errs.push('Recurrence interval must be at least 1')
    return errs
}

export function buildTaskBody(
    vals: TaskFormValues,
    extras?: Record<string, unknown>
): Record<string, unknown> {
    const body: Record<string, unknown> = {
        label: vals.label.trim(),
        notes: vals.notes,
        priority: vals.priority,
        dueDate: vals.dueDate || null,
        estimatedCost: vals.estimatedCost === '' ? null : Number(vals.estimatedCost),
        isRecurring: vals.isRecurring,
        recurrenceInterval: vals.isRecurring ? vals.recurrenceInterval : 0,
        recurrenceUnit: vals.isRecurring ? vals.recurrenceUnit : '',
        recurrenceMode: vals.isRecurring ? vals.recurrenceMode : '',
        ...extras,
    }
    return body
}

export default {
    validateTaskValues,
    buildTaskBody,
}

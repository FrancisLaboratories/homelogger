import React, { useEffect, useRef, useState } from 'react'
import { Button, Modal, Form } from 'react-bootstrap'
import { RepairRecord } from './RepairSection'
import { SERVER_URL } from '@/pages/_app'

interface ShowRepairModalProps {
    show: boolean
    handleClose: () => void
    repairRecord: RepairRecord
    handleDeleteRepair: (id: number) => void
    handleUpdateRepair?: (updated: RepairRecord) => void
}

interface AttachmentInfo {
    id: number
    originalName: string
}

const ShowRepairModal: React.FC<ShowRepairModalProps> = ({
    show,
    handleClose,
    repairRecord,
    handleDeleteRepair,
    handleUpdateRepair,
}) => {
    const [attachments, setAttachments] = useState<AttachmentInfo[]>([])
    const [uploadMessage, setUploadMessage] = useState<string>('')
    const [uploadError, setUploadError] = useState<string>('')
    const fileInputRef = useRef<HTMLInputElement | null>(null)

    // Edit state
    const [editing, setEditing] = useState(false)
    const [editDescription, setEditDescription] = useState('')
    const [editDate, setEditDate] = useState('')
    const [editCost, setEditCost] = useState(0)
    const [editNotes, setEditNotes] = useState('')
    const [isSaving, setIsSaving] = useState(false)
    const [saveError, setSaveError] = useState('')

    useEffect(() => {
        const loadAttachments = async () => {
            try {
                const resp = await fetch(`${SERVER_URL}/files/repair/${repairRecord.id}`)
                if (!resp.ok) return
                const data: Array<{
                    id: number
                    originalName: string
                    userID?: string
                }> = await resp.json()
                setAttachments(
                    data.map((f) => ({
                        id: f.id,
                        originalName: f.originalName,
                    }))
                )
            } catch (err) {
                console.error('Error loading attachments', err)
            }
        }

        if (show) loadAttachments()
    }, [show, repairRecord.id])

    // Reset edit state when record changes or modal closes
    useEffect(() => {
        if (!show) {
            setEditing(false)
            setSaveError('')
        } else {
            setEditDescription(repairRecord.description)
            setEditDate(repairRecord.date)
            setEditCost(repairRecord.cost)
            setEditNotes(repairRecord.notes)
        }
    }, [show, repairRecord])

    const handleDeleteAttachment = async (fileId: number) => {
        if (!window.confirm('Delete attachment?')) return
        try {
            const resp = await fetch(`${SERVER_URL}/files/${fileId}`, {
                method: 'DELETE',
            })
            if (!resp.ok) throw new Error('Failed to delete file')
            setAttachments((prev) => prev.filter((a) => a.id !== fileId))
        } catch (err) {
            console.error('Error deleting attachment', err)
        }
    }

    const handleAddFiles = async (files: FileList | null) => {
        if (!files || files.length === 0) return
        setUploadMessage('')
        setUploadError('')
        for (const f of Array.from(files)) {
            try {
                const formData = new FormData()
                formData.append('file', f)
                formData.append('userID', '1')

                const uploadResp = await fetch(`${SERVER_URL}/files/upload`, {
                    method: 'POST',
                    body: formData,
                })
                if (!uploadResp.ok) throw new Error(await uploadResp.text())
                const uploadData = await uploadResp.json()

                // attach to repair
                const attachResp = await fetch(`${SERVER_URL}/files/attach`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        fileId: uploadData.id,
                        repairId: repairRecord.id,
                    }),
                })
                if (!attachResp.ok) throw new Error('Failed to attach file')

                setAttachments((prev) => [
                    ...prev,
                    {
                        id: uploadData.id,
                        originalName: uploadData.originalName,
                    },
                ])
            } catch (err) {
                console.error('Error adding file', err)
                setUploadError(`Upload failed: ${String(err)}`)
                if (fileInputRef.current) fileInputRef.current.value = ''
                return
            }
        }
        setUploadMessage('Upload successful')
        if (fileInputRef.current) fileInputRef.current.value = ''
        setTimeout(() => setUploadMessage(''), 3000)
    }
    const handleSaveEdit = async () => {
        setSaveError('')
        if (!editDescription.trim()) {
            setSaveError('Description is required')
            return
        }
        if (!editDate) {
            setSaveError('Date is required')
            return
        }
        setIsSaving(true)
        try {
            const res = await fetch(`${SERVER_URL}/repair/update/${repairRecord.id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    description: editDescription.trim(),
                    date: editDate,
                    cost: editCost,
                    notes: editNotes,
                }),
            })
            if (!res.ok) throw new Error(await res.text())
            const updated: RepairRecord = await res.json()
            handleUpdateRepair?.(updated)
            setEditing(false)
        } catch (err) {
            setSaveError(String(err))
        } finally {
            setIsSaving(false)
        }
    }

    const handleDelete = async () => {
        if (window.confirm('Are you sure you want to delete this?')) {
            try {
                const response = await fetch(`${SERVER_URL}/repair/delete/${repairRecord.id}`, {
                    method: 'DELETE',
                })

                if (!response.ok) {
                    throw new Error('Failed to delete repair record')
                }

                console.log('Record deleted')
                handleDeleteRepair(repairRecord.id)
                handleClose()
            } catch (error) {
                console.error('Error deleting repair record:', error)
            }
        }
    }

    return (
        <Modal show={show} onHide={handleClose}>
            <Modal.Header closeButton>
                <Modal.Title>Repair Record Details</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                {saveError && <div className="alert alert-danger py-2">{saveError}</div>}

                {editing ? (
                    <>
                        <Form.Group className="mb-3" controlId="editDescription">
                            <Form.Label>Description</Form.Label>
                            <Form.Control
                                type="text"
                                value={editDescription}
                                onChange={(e) => setEditDescription(e.target.value)}
                            />
                        </Form.Group>
                        <Form.Group className="mb-3" controlId="editDate">
                            <Form.Label>Date</Form.Label>
                            <Form.Control
                                type="date"
                                value={editDate}
                                onChange={(e) => setEditDate(e.target.value)}
                            />
                        </Form.Group>
                        <Form.Group className="mb-3" controlId="editCost">
                            <Form.Label>Cost ($)</Form.Label>
                            <Form.Control
                                type="number"
                                min={0}
                                step="0.01"
                                value={editCost}
                                onChange={(e) => setEditCost(parseFloat(e.target.value) || 0)}
                            />
                        </Form.Group>
                        <Form.Group className="mb-3" controlId="editNotes">
                            <Form.Label>Notes</Form.Label>
                            <Form.Control
                                as="textarea"
                                rows={3}
                                value={editNotes}
                                onChange={(e) => setEditNotes(e.target.value)}
                            />
                        </Form.Group>
                    </>
                ) : (
                    <>
                        <p>
                            <strong>Description:</strong> {repairRecord.description}
                        </p>
                        <p>
                            <strong>Date:</strong> {repairRecord.date}
                        </p>
                        <p>
                            <strong>Cost:</strong> ${repairRecord.cost}
                        </p>
                        <Form.Group>
                            <Form.Label>
                                <strong>Notes:</strong>
                            </Form.Label>
                            <div className="form-control text-muted">
                                {repairRecord.notes || 'None'}
                            </div>
                        </Form.Group>
                    </>
                )}

                <div className="mt-3">
                    <strong>Attachments:</strong>
                    {attachments.length === 0 && !editing ? (
                        <p className="text-muted mb-0">None</p>
                    ) : (
                        <ul className="mb-0">
                            {attachments.map((att) => (
                                <li key={att.id}>
                                    <a
                                        href={`${SERVER_URL}/files/download/${att.id}`}
                                        target="_blank"
                                        rel="noreferrer"
                                    >
                                        {att.originalName || `File ${att.id}`}
                                    </a>{' '}
                                    {editing && (
                                        <Button
                                            variant="danger"
                                            size="sm"
                                            onClick={() => handleDeleteAttachment(att.id)}
                                            style={{ marginLeft: '8px' }}
                                        >
                                            Delete
                                        </Button>
                                    )}
                                </li>
                            ))}
                        </ul>
                    )}
                </div>

                {editing && (
                    <Form.Group controlId="formAddFiles" className="mt-3">
                        <Form.Label>Add attachments</Form.Label>
                        <Form.Control
                            ref={fileInputRef}
                            type="file"
                            multiple
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                                handleAddFiles(e.target.files)
                            }
                        />
                        {uploadMessage && (
                            <div style={{ color: 'green', marginTop: '6px' }}>{uploadMessage}</div>
                        )}
                        {uploadError && (
                            <div style={{ color: 'red', marginTop: '6px' }}>{uploadError}</div>
                        )}
                    </Form.Group>
                )}
            </Modal.Body>
            <Modal.Footer
                className={`d-flex ${editing ? 'justify-content-between' : 'justify-content-end'}`}
            >
                {editing && (
                    <Button variant="danger" onClick={handleDelete}>
                        <i className="bi bi-trash"></i>
                    </Button>
                )}
                <div className="d-flex gap-2">
                    {editing ? (
                        <>
                            <Button
                                variant="secondary"
                                onClick={() => {
                                    setEditing(false)
                                    setSaveError('')
                                }}
                                disabled={isSaving}
                            >
                                Cancel
                            </Button>
                            <Button variant="primary" onClick={handleSaveEdit} disabled={isSaving}>
                                {isSaving ? 'Saving…' : 'Save'}
                            </Button>
                        </>
                    ) : (
                        <>
                            <Button variant="outline-secondary" onClick={() => setEditing(true)}>
                                <i className="bi bi-pencil me-1"></i>Edit
                            </Button>
                            <Button variant="secondary" onClick={handleClose}>
                                Close
                            </Button>
                        </>
                    )}
                </div>
            </Modal.Footer>
        </Modal>
    )
}

export default ShowRepairModal

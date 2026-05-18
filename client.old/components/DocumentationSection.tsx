import React, { useEffect, useRef, useState } from 'react'
import { Card, Button, Form } from 'react-bootstrap'
import { SERVER_URL } from '@/pages/_app'

interface Props {
    applianceId?: number
    spaceType?: string
}

interface FileInfo {
    id: number
    originalName: string
}

const DocumentationSection: React.FC<Props> = ({ applianceId, spaceType }) => {
    const [files, setFiles] = useState<FileInfo[]>([])
    const [uploadMessage, setUploadMessage] = useState<string>('')
    const fileInputRef = useRef<HTMLInputElement | null>(null)

    useEffect(() => {
        const load = async () => {
            try {
                const url = spaceType
                    ? `${SERVER_URL}/files/space/${spaceType}`
                    : `${SERVER_URL}/files/appliance/${applianceId}`
                const resp = await fetch(url)
                if (!resp.ok) return
                const data: Array<{ id: number; originalName: string }> = await resp.json()
                setFiles(
                    data.map((d) => ({
                        id: d.id,
                        originalName: d.originalName,
                    }))
                )
            } catch (err) {
                console.error('Error loading docs', err)
            }
        }
        load()
    }, [applianceId, spaceType])

    const handleAddFiles = async (filesList: FileList | null) => {
        if (!filesList || filesList.length === 0) return
        for (const f of Array.from(filesList)) {
            try {
                const formData = new FormData()
                formData.append('file', f)
                formData.append('userID', '1')

                if (spaceType) {
                    formData.append('spaceType', spaceType)
                }

                const uploadResp = await fetch(`${SERVER_URL}/files/upload`, {
                    method: 'POST',
                    body: formData,
                })
                if (!uploadResp.ok) throw new Error('Upload failed')
                const uploadData: { id: number; originalName?: string } = await uploadResp.json()

                type AttachBody = {
                    fileId: number
                    applianceId?: number
                    spaceType?: string
                }
                const attachBody: AttachBody = { fileId: uploadData.id }
                if (applianceId) attachBody.applianceId = applianceId
                if (spaceType) attachBody.spaceType = spaceType

                await fetch(`${SERVER_URL}/files/attach`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(attachBody),
                })

                setFiles((prev) => [
                    ...prev,
                    {
                        id: uploadData.id,
                        originalName: uploadData.originalName || `File ${uploadData.id}`,
                    },
                ])
            } catch (err) {
                console.error('Error adding doc', err)
            }
        }
        setUploadMessage('Upload successful')
        if (fileInputRef.current) fileInputRef.current.value = ''
        setTimeout(() => setUploadMessage(''), 3000)
    }

    const handleDelete = async (fileId: number) => {
        if (!window.confirm('Delete this document?')) return
        try {
            const resp = await fetch(`${SERVER_URL}/files/${fileId}`, {
                method: 'DELETE',
            })
            if (!resp.ok) throw new Error('Delete failed')
            setFiles((prev) => prev.filter((f) => f.id !== fileId))
        } catch (err) {
            console.error('Error deleting file', err)
        }
    }

    return (
        <Card>
            <Card.Body>
                <h5>Documents</h5>
                {files.length === 0 ? (
                    <div>No documents uploaded.</div>
                ) : (
                    <ul>
                        {files.map((f) => (
                            <li key={f.id}>
                                <a
                                    href={`${SERVER_URL}/files/download/${f.id}`}
                                    target="_blank"
                                    rel="noreferrer"
                                >
                                    {f.originalName}
                                </a>{' '}
                                <Button
                                    variant="danger"
                                    size="sm"
                                    onClick={() => handleDelete(f.id)}
                                    style={{ marginLeft: '8px' }}
                                >
                                    Delete
                                </Button>
                            </li>
                        ))}
                    </ul>
                )}

                <Form.Group controlId="applianceFiles">
                    <Form.Label>Add document</Form.Label>
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
                </Form.Group>
            </Card.Body>
        </Card>
    )
}

export default DocumentationSection

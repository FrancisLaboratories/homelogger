import React, {useEffect, useRef, useState} from 'react';
import {Button, Modal, Form} from 'react-bootstrap';
import {MaintenanceRecord} from './MaintenanceSection';
import {SERVER_URL} from "@/lib/config";
import {useSettings} from "@/components/SettingsContext";
import {formatCurrency, formatDate} from "@/lib/formatters";

interface ShowMaintenanceModalProps {
    show: boolean;
    handleClose: () => void;
    maintenanceRecord: MaintenanceRecord;
    handleDeleteMaintenance: (id: number) => void;
}

interface AttachmentInfo { id: number; originalName: string }

const ShowMaintenanceModal: React.FC<ShowMaintenanceModalProps> = ({show, handleClose, maintenanceRecord, handleDeleteMaintenance}) => {
    const {settings} = useSettings();
    const [attachments, setAttachments] = useState<AttachmentInfo[]>([]);
    const [uploadMessage, setUploadMessage] = useState<string>('');
    const fileInputRef = useRef<HTMLInputElement | null>(null);

    useEffect(() => {
        const loadAttachments = async () => {
            try {
                const resp = await fetch(`${SERVER_URL}/files/maintenance/${maintenanceRecord.id}`);
                if (!resp.ok) return;
                const data: Array<{ id: number; originalName: string; userID?: string }> = await resp.json();
                setAttachments(data.map((f) => ({ id: f.id, originalName: f.originalName })));
            } catch (err) {
                console.error('Error loading attachments', err);
            }
        };

        if (show) loadAttachments();
    }, [show, maintenanceRecord.id]);

    const handleDeleteAttachment = async (fileId: number) => {
        if (!window.confirm('Delete attachment?')) return;
        try {
            const resp = await fetch(`${SERVER_URL}/files/${fileId}`, { method: 'DELETE' });
            if (!resp.ok) throw new Error('Failed to delete file');
            setAttachments(prev => prev.filter(a => a.id !== fileId));
        } catch (err) {
            console.error('Error deleting attachment', err);
        }
    };

    const handleAddFiles = async (files: FileList | null) => {
        if (!files || files.length === 0) return;
        for (const f of Array.from(files)) {
            try {
                const formData = new FormData();
                formData.append('file', f);
                formData.append('userID', '1');

                const uploadResp = await fetch(`${SERVER_URL}/files/upload`, { method: 'POST', body: formData });
                if (!uploadResp.ok) throw new Error('Upload failed');
                const uploadData = await uploadResp.json();

                // attach to maintenance
                await fetch(`${SERVER_URL}/files/attach`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ fileId: uploadData.id, maintenanceId: maintenanceRecord.id })
                });

                setAttachments(prev => [...prev, { id: uploadData.id, originalName: uploadData.originalName }]);
            } catch (err) {
                console.error('Error adding file', err);
            }
        }
        // notify and clear file input
        setUploadMessage('Upload successful');
        if (fileInputRef.current) fileInputRef.current.value = '';
        setTimeout(() => setUploadMessage(''), 3000);
    };
    const handleDelete = async () => {
        if (window.confirm('Are you sure you want to delete this?')) {
            try {
                const response = await fetch(`${SERVER_URL}/maintenance/delete/${maintenanceRecord.id}`, {
                    method: 'DELETE',
                });

                if (!response.ok) {
                    throw new Error('Failed to delete maintenance record');
                }

                console.log('Record deleted');
                handleDeleteMaintenance(maintenanceRecord.id);
                handleClose();
            } catch (error) {
                console.error('Error deleting maintenance record:', error);
            }
        }
    };

    return (
        <Modal show={show} onHide={handleClose}>
            <Modal.Header closeButton>
                <Modal.Title>Maintenance Record Details</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <p><strong>Description:</strong> {maintenanceRecord.description}</p>
                <p><strong>Date:</strong> {formatDate(maintenanceRecord.date, settings)}</p>
                <p><strong>Cost:</strong> {formatCurrency(maintenanceRecord.cost, settings)}</p>
                {attachments.length > 0 && (
                    <div>
                        <strong>Attachments:</strong>
                        <ul>
                            {attachments.map(att => (
                                <li key={att.id}>
                                    <a href={`${SERVER_URL}/files/download/${att.id}`} target="_blank" rel="noreferrer">{att.originalName || `File ${att.id}`}</a>
                                    {' '}
                                    <Button variant="danger" size="sm" onClick={() => handleDeleteAttachment(att.id)} style={{marginLeft: '8px'}}>Delete</Button>
                                </li>
                            ))}
                        </ul>
                    </div>
                )}

                <Form.Group controlId="formAddFiles">
                    <Form.Label>Add attachments</Form.Label>
                    <Form.Control ref={fileInputRef} type="file" multiple onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleAddFiles(e.target.files)} />
                    {uploadMessage && <div style={{color: 'green', marginTop: '6px'}}>{uploadMessage}</div>}
                </Form.Group>
                <Form.Group>
                    <Form.Label><strong>Notes:</strong></Form.Label>
                    <div className='form-control'>
                        {maintenanceRecord.notes}
                    </div>
                </Form.Group>
            </Modal.Body>
            <Modal.Footer className="d-flex justify-content-between">
                <Button variant="danger" onClick={handleDelete}>
                    <i className="bi bi-trash"></i>
                </Button>
                <Button variant="secondary" onClick={handleClose}>
                    Close
                </Button>
            </Modal.Footer>
        </Modal>
    );
};

export default ShowMaintenanceModal;

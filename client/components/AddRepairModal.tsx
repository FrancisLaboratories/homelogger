import React, {useEffect, useState} from 'react';
import {Button, Form, Modal} from 'react-bootstrap';
import {SERVER_URL} from "@/lib/config";
import {RepairRecord, RepairReferenceType, RepairSpaceType} from './RepairSection';
import {useSettings} from "@/components/SettingsContext";

interface AddRepairModalProps {
    show: boolean;
    handleClose: () => void;
    handleSave: (repair: RepairRecord) => void;
    applianceId?: number;
    referenceType: RepairReferenceType;
    spaceType: RepairSpaceType;
}

const AddRepairModal: React.FC<AddRepairModalProps> = ({
    show,
    handleClose,
    handleSave,
    applianceId,
    referenceType,
    spaceType
}) => {
    const {settings} = useSettings();
    const [description, setDescription] = useState('');
    const [date, setDate] = useState('');
    const [cost, setCost] = useState(0);
    const [notes, setNotes] = useState('');
    const [files, setFiles] = useState<File[]>([]);
    const [errors, setErrors] = useState<string[]>([]);
    const [isSubmitting, setIsSubmitting] = useState(false);

    useEffect(() => {
        if (!show) {
            setDescription('');
            setDate('');
            setCost(0);
            setNotes('');
            setFiles([]);
        }
    }, [show]);

    const handleSubmit = async () => {
        setErrors([]);
        const errs: string[] = [];
        if (!description || description.trim() === '') errs.push('Description is required');
        if (!date) errs.push('Date is required');
        if (isNaN(cost) || cost < 0) errs.push('Cost must be a positive number');
        if (errs.length > 0) {
            setErrors(errs);
            return;
        }

        const standardizedDate = new Date(date).toISOString().split('T')[0];
        setIsSubmitting(true);

        const attachmentIds: number[] = [];

        for (const file of files) {
            try {
                const formData = new FormData();
                formData.append('file', file);
                formData.append('userID', '1');

                const uploadResp = await fetch(`${SERVER_URL}/files/upload`, {
                    method: 'POST',
                    body: formData,
                });

                if (!uploadResp.ok) {
                    console.error('Failed to upload attachment', file.name);
                    continue;
                }

                const uploadData: { id: number; originalName?: string } = await uploadResp.json();
                if (uploadData && uploadData.id) {
                    attachmentIds.push(uploadData.id);
                }
            } catch (err) {
                console.error('Error uploading file:', err);
            }
        }

        const newRepair: {
            description: string;
            date: string;
            cost: number;
            notes: string;
            spaceType: RepairSpaceType;
            referenceType: RepairReferenceType;
            applianceId: number;
            attachmentIds: number[];
        } = {
            description,
            date: standardizedDate,
            cost,
            notes,
            spaceType,
            referenceType,
            applianceId: applianceId || 0,
            attachmentIds,
        };

        try {
            const response = await fetch(`${SERVER_URL}/repair/add`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(newRepair),
            });

            if (!response.ok) {
                throw new Error('Failed to add repair record');
            }

            const addedRepair = await response.json();
            handleSave(addedRepair);
            handleClose();
        } catch (error) {
            console.error('Error adding repair record:', error);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <Modal show={show} onHide={handleClose}>
            <Modal.Header closeButton>
                <Modal.Title>Add Repair Record</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <Form>
                    <Form.Group controlId="formDescription">
                        <Form.Label>Description</Form.Label>
                        <Form.Control
                            type="text"
                            placeholder="Enter description"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group controlId="formDate">
                        <Form.Label>Date</Form.Label>
                        <Form.Control
                            type="date"
                            value={date}
                            onChange={(e) => setDate(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group controlId="formCost">
                        <Form.Label>Cost ({settings.currency})</Form.Label>
                        <Form.Control
                            type="number"
                            placeholder={`Enter cost in ${settings.currency}`}
                            value={cost}
                            onChange={(e) => setCost(parseFloat(e.target.value))}
                        />
                    </Form.Group>
                    <Form.Group controlId="formNotes">
                        <Form.Label>Notes</Form.Label>
                        <Form.Control
                            as="textarea"
                            placeholder="Enter notes"
                            value={notes}
                            onChange={(e) => setNotes(e.target.value)}
                        />
                    </Form.Group>
                        <Form.Group controlId="formFile">
                            <Form.Label>Attachment</Form.Label>
                            <Form.Control
                                type="file"
                                multiple
                                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFiles(e.target.files ? Array.from(e.target.files) : [])}
                            />
                        </Form.Group>
                        {errors.length > 0 && (
                            <div style={{color: 'red', marginTop: '8px'}}>
                                {errors.map((e, idx) => <div key={idx}>{e}</div>)}
                            </div>
                        )}
                </Form>
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={handleClose}>
                    Close
                </Button>
                        <Button variant="primary" onClick={handleSubmit} disabled={isSubmitting}>
                            {isSubmitting ? 'Saving...' : 'Save Changes'}
                        </Button>
            </Modal.Footer>
        </Modal>
    );
};

export default AddRepairModal;

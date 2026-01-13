import React, {useEffect, useState} from 'react';
import {Button, Form, Modal} from 'react-bootstrap';

interface AddApplianceModalProps {
    show: boolean;
    handleClose: () => void;
    handleSave: (applianceName: string, manufacturer: string, modelNumber: string, serialNumber: string, yearPurchased: string, purchasePrice: string, location: string, type: string) => void;
}

const AddApplianceModal: React.FC<AddApplianceModalProps> = ({show, handleClose, handleSave}) => {
    const [applianceName, setApplianceName] = useState('');
    const [manufacturer, setManufacturer] = useState('');
    const [modelNumber, setModelNumber] = useState('');
    const [serialNumber, setSerialNumber] = useState('');
    const [yearPurchased, setYearPurchased] = useState('');
    const [purchasePrice, setPurchasePrice] = useState('');
    const [location, setLocation] = useState('');
    const [type, setType] = useState('');
    const [errors, setErrors] = useState<string[]>([]);
    const [isSubmitting, setIsSubmitting] = useState(false);

    useEffect(() => {
        if (!show) {
            setApplianceName('');
            setManufacturer('');
            setModelNumber('');
            setSerialNumber('');
            setYearPurchased('');
            setPurchasePrice('');
            setLocation('');
            setType('');
        }
    }, [show]);

    const handleSubmit = () => {
        setErrors([]);
        const errs: string[] = [];
        if (!applianceName || applianceName.trim() === '') errs.push('Appliance name is required');
        if (!type || type === '') errs.push('Type is required');
        if (errs.length > 0) {
            setErrors(errs);
            return;
        }
        setIsSubmitting(true);
        handleSave(applianceName, manufacturer, modelNumber, serialNumber, yearPurchased, purchasePrice, location, type);
        setIsSubmitting(false);
        handleClose();
    };

    return (
        <Modal show={show} onHide={handleClose}>
            <Modal.Header closeButton>
                <Modal.Title>Add Appliance</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <Form>
                    <Form.Group controlId="formApplianceName">
                        <Form.Label>Appliance name</Form.Label>
                        <Form.Control
                            type="text"
                            placeholder="e.g. Kitchen Refrigerator, Upstairs Vacuum..."
                            value={applianceName}
                            onChange={(e) => setApplianceName(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group controlId="formType">
                        <Form.Label>Type</Form.Label>
                        <Form.Control
                            type="text"
                            placeholder="e.g. Refrigerator, Oven, Washer..."
                            value={type}
                            onChange={(e) => setType(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group controlId="formLocation">
                        <Form.Label>Location</Form.Label>
                        <Form.Control
                            type="text"
                            placeholder="e.g. Living Room, Kitchen..."
                            value={location}
                            onChange={(e) => setLocation(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group controlId="formManufacturer">
                        <Form.Label>Manufacturer</Form.Label>
                        <Form.Control
                            type="text"
                            placeholder="e.g. Whirlpool, Samsung, LG..."
                            value={manufacturer}
                            onChange={(e) => setManufacturer(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group controlId="formModelNumber">
                        <Form.Label>Model Number</Form.Label>
                        <Form.Control
                            type="text"
                            placeholder="e.g. WRX735SDHZ, RF28R7351SG..."
                            value={modelNumber}
                            onChange={(e) => setModelNumber(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group controlId="formSerialNumber">
                        <Form.Label>Serial Number</Form.Label>
                        <Form.Control
                            type="text"
                            placeholder="e.g. 12345ABC, SN987654321..."
                            value={serialNumber}
                            onChange={(e) => setSerialNumber(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group controlId="formYearPurchased">
                        <Form.Label>Year Purchased</Form.Label>
                        <Form.Control
                            type="text"
                            placeholder="e.g. 2020, 2021..."
                            value={yearPurchased}
                            onChange={(e) => setYearPurchased(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group controlId="formPurchasePrice">
                        <Form.Label>Purchase Price</Form.Label>
                        <Form.Control
                            type="text"
                            placeholder="e.g. 499.99, 1299.00..."
                            value={purchasePrice}
                            onChange={(e) => setPurchasePrice(e.target.value)}
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

export default AddApplianceModal;
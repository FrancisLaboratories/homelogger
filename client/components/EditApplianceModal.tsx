import React, {useEffect, useState} from 'react';
import {Button, Form, Modal} from 'react-bootstrap';
import {SERVER_URL} from "@/pages/_app";

interface EditApplianceModalProps {
    show: boolean;
    handleClose: () => void;
    handleSave: (id: number, applianceName: string, manufacturer: string, modelNumber: string, serialNumber: string, yearPurchased: string, purchasePrice: string, location: string, type: string) => void;
    appliance: {
        id: number;
        applianceName: string;
        manufacturer: string;
        modelNumber: string;
        serialNumber: string;
        yearPurchased: string;
        purchasePrice: string;
        location: string;
        type: string;
    };
}

const EditApplianceModal: React.FC<EditApplianceModalProps> = ({show, handleClose, handleSave, appliance}) => {
    const [applianceName, setApplianceName] = useState(appliance.applianceName);
    const [manufacturer, setManufacturer] = useState(appliance.manufacturer);
    const [modelNumber, setModelNumber] = useState(appliance.modelNumber);
    const [serialNumber, setSerialNumber] = useState(appliance.serialNumber);
    const [yearPurchased, setYearPurchased] = useState(appliance.yearPurchased);
    const [purchasePrice, setPurchasePrice] = useState(appliance.purchasePrice);
    const [location, setLocation] = useState(appliance.location);
    const [type, setType] = useState(appliance.type);

    useEffect(() => {
        if (show) {
            setApplianceName(appliance.applianceName);
            setManufacturer(appliance.manufacturer);
            setModelNumber(appliance.modelNumber);
            setSerialNumber(appliance.serialNumber);
            setYearPurchased(appliance.yearPurchased);
            setPurchasePrice(appliance.purchasePrice);
            setLocation(appliance.location);
            setType(appliance.type);
        }
    }, [show, appliance]);

    const handleSubmit = async () => {
        const updatedAppliance = {
            id: appliance.id,
            applianceName,
            manufacturer,
            modelNumber,
            serialNumber,
            yearPurchased,
            purchasePrice,
            location,
            type
        };

        try {
            const response = await fetch(`${SERVER_URL}/appliances/update/${appliance.id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(updatedAppliance),
            });

            if (!response.ok) {
                throw new Error('Failed to update appliance');
            }

            const savedAppliance = await response.json();
            handleSave(savedAppliance.id, savedAppliance.applianceName, savedAppliance.manufacturer, savedAppliance.modelNumber, savedAppliance.serialNumber, savedAppliance.yearPurchased, savedAppliance.purchasePrice, savedAppliance.location, savedAppliance.type);
            handleClose();
        } catch (error) {
            console.error('Error updating appliance:', error);
        }
    };

    return (
        <Modal show={show} onHide={handleClose}>
            <Modal.Header closeButton>
                <Modal.Title>Edit Appliance</Modal.Title>
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
                </Form>
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={handleClose}>
                    Close
                </Button>
                <Button variant="primary" onClick={handleSubmit}>
                    Save Changes
                </Button>
            </Modal.Footer>
        </Modal>
    );
};

export default EditApplianceModal;
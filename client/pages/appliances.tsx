'use client';

import React, {useEffect, useState} from 'react';
import {Container, Row} from 'react-bootstrap';
import MyNavbar from '../components/Navbar';
import BlankCard from '../components/BlankCard';
import AddApplianceModal from '../components/AddApplianceModal';
import ApplianceCard from '../components/ApplianceCard';
import {SERVER_URL} from "@/lib/config";

const appliancesUrl = `${SERVER_URL}/appliances`;
const appliancesAddUrl = `${SERVER_URL}/appliances/add`;

interface Appliance {
    id: number;
    applianceName: string;
    manufacturer: string;
    modelNumber: string;
    serialNumber: string;
    yearPurchased: string;
    purchasePrice: string;
    location: string;
    type: string;
}

const AppliancesPage: React.FC = () => {
    const [appliances, setAppliances] = useState<Appliance[]>([]);
    const [showModal, setShowModal] = useState(false);

    useEffect(() => {
        fetch(appliancesUrl)
            .then(response => response.json())
            .then(data => setAppliances(data))
            .catch(error => console.error('Error fetching appliances:', error));
    }, []);

    const handleAddCardClick = () => {
        setShowModal(true);
    };

    const handleCloseModal = () => {
        setShowModal(false);
    };

    const handleSaveAppliance = async (applianceName: string, manufacturer: string, modelNumber: string, serialNumber: string, yearPurchased: string, purchasePrice: string, location: string, type: string) => {
        const newAppliance = {
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
            const response = await fetch(appliancesAddUrl, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(newAppliance),
            });

            if (!response.ok) {
                throw new Error('Failed to add appliance');
            }

            const addedAppliance = await response.json();
            setAppliances(prevAppliances => [...prevAppliances, addedAppliance]);
        } catch (error) {
            console.error('Error adding appliance:', error);
        }

        setShowModal(false);
    };

    const handleApplianceClick = (id: number) => {
        window.location.href = `/appliance.html?id=${id}`;
    };

    return (
        <Container>
            <MyNavbar/>
            <h4 id='maintext'>Appliances</h4>
            <Row style={{display: 'flex', flexWrap: 'wrap', justifyContent: 'flex-start', padding: '0'}}>
                {appliances.map(appliance => (
                    <div key={appliance.id} style={{flex: '0 0 18rem', margin: '0.25rem'}}>
                        <ApplianceCard
                            id={appliance.id}
                            applianceName={appliance.applianceName}
                            manufacturer={appliance.manufacturer}
                            modelNumber={appliance.modelNumber}
                            serialNumber={appliance.serialNumber}
                            yearPurchased={appliance.yearPurchased}
                            purchasePrice={appliance.purchasePrice}
                            location={appliance.location}
                            type={appliance.type}
                            onClick={() => handleApplianceClick(appliance.id)}
                        />
                    </div>
                ))}
                <div style={{flex: '0 0 18rem', margin: '0.25rem'}}>
                    <BlankCard onClick={handleAddCardClick}/>
                </div>
            </Row>
            <AddApplianceModal
                show={showModal}
                handleClose={handleCloseModal}
                handleSave={handleSaveAppliance}
            />
        </Container>
    );
};

export default AppliancesPage;
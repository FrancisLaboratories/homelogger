import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Row } from "react-bootstrap";
import AddApplianceModal from "@/components/AddApplianceModal";
import ApplianceCard from "@/components/ApplianceCard";
import BlankCard from "@/components/BlankCard";
import { SERVER_URL } from "@/context/DemoContext";

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

const appliancesUrl = `${SERVER_URL}/appliances`;
const appliancesAddUrl = `${SERVER_URL}/appliances/add`;

const AppliancesPage: React.FC = () => {
  const [appliances, setAppliances] = useState<Appliance[]>([]);
  const [showAdd, setShowAdd] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    fetch(appliancesUrl)
      .then((response) => response.json())
      .then((data) => setAppliances(data))
      .catch((error) => console.error("Error fetching appliances:", error));
  }, []);

  const handleSave = async (
    applianceName: string,
    manufacturer: string,
    modelNumber: string,
    serialNumber: string,
    yearPurchased: string,
    purchasePrice: string,
    location: string,
    type: string,
  ) => {
    const newAppliance = {
      applianceName,
      manufacturer,
      modelNumber,
      serialNumber,
      yearPurchased,
      purchasePrice,
      location,
      type,
    };
    try {
      const response = await fetch(appliancesAddUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(newAppliance),
      });

      if (!response.ok) {
        throw new Error("Failed to add appliance");
      }

      const addedAppliance = await response.json();
      setAppliances((prevAppliances) => [...prevAppliances, addedAppliance]);
    } catch (error) {
      console.error("Error adding appliance:", error);
    }

    setShowAdd(false);
  };

  return (
    <div>
      <h4 id="maintext">Appliances</h4>
      <Row
        style={{
          display: "flex",
          flexWrap: "wrap",
          justifyContent: "flex-start",
          padding: "0",
        }}
      >
        {appliances.map((a) => (
          <div key={a.id} style={{ flex: "0 0 18rem", margin: "0.25rem" }}>
            <ApplianceCard
              id={a.id}
              applianceName={a.applianceName}
              location={a.location}
              type={a.type}
              onClick={() => navigate(`/appliance?id=${a.id}`)}
            />
          </div>
        ))}
        <div style={{ flex: "0 0 18rem", margin: "0.25rem" }}>
          <BlankCard onClick={() => setShowAdd(true)} />
        </div>
      </Row>

      <AddApplianceModal
        show={showAdd}
        handleClose={() => setShowAdd(false)}
        handleSave={handleSave}
      />
    </div>
  );
};

export default AppliancesPage;

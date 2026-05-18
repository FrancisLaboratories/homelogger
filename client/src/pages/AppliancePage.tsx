import React, { useEffect, useState } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { Button, Card, Col, Row, Tab, Tabs } from "react-bootstrap";
import "bootstrap-icons/font/bootstrap-icons.css";
import MaintenanceSection from "@/components/MaintenanceSection";
import RepairSection from "@/components/RepairSection";
import DocumentationSection from "@/components/DocumentationSection";
import TasksSection from "@/components/TasksSection";
import NotesSection from "@/components/NotesSection";
import EditApplianceModal from "@/components/EditApplianceModal";
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

const AppliancePage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const idParam = searchParams.get("id");
  const id = idParam ? Number(idParam) : null;

  const [appliance, setAppliance] = useState<Appliance | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);

  useEffect(() => {
    if (id) {
      fetch(`${SERVER_URL}/appliances/${id}`)
        .then((res) => res.json())
        .then((data) => setAppliance(data))
        .catch((err) => console.error("Error fetching appliance:", err));
    }
  }, [id]);

  const handleDelete = async () => {
    if (!id) return;
    if (!window.confirm("Are you sure you want to delete this appliance?"))
      return;
    try {
      const res = await fetch(`${SERVER_URL}/appliances/delete/${id}`, {
        method: "DELETE",
      });
      if (!res.ok) throw new Error("Failed to delete appliance");
      navigate("/appliances");
    } catch (err) {
      console.error("Error deleting appliance:", err);
    }
  };

  const handleSaveAppliance = (
    savedId: number,
    applianceName: string,
    manufacturer: string,
    modelNumber: string,
    serialNumber: string,
    yearPurchased: string,
    purchasePrice: string,
    location: string,
    type: string,
  ) => {
    setAppliance({
      id: savedId,
      applianceName,
      manufacturer,
      modelNumber,
      serialNumber,
      yearPurchased,
      purchasePrice,
      location,
      type,
    });
  };

  if (!id) {
    return <div>No appliance id provided.</div>;
  }
  if (!appliance) {
    return <div>Loading...</div>;
  }

  return (
    <Row className="justify-content-center">
      <Col md={8}>
        <Tabs defaultActiveKey="main" id="appliance-tabs">
          <Tab eventKey="main" title="Main">
            <Card>
              <Card.Body>
                <Card.Title>{appliance.applianceName}</Card.Title>
                <Card.Subtitle className="mb-2 text-muted">
                  {appliance.type}
                </Card.Subtitle>
                <Card.Text>
                  <strong>Location:</strong> {appliance.location}
                  <br />
                  <strong>Manufacturer:</strong> {appliance.manufacturer}
                  <br />
                  <strong>Model Number:</strong> {appliance.modelNumber}
                  <br />
                  <strong>Serial Number:</strong> {appliance.serialNumber}
                  <br />
                  <strong>Year Purchased:</strong> {appliance.yearPurchased}
                  <br />
                  <strong>Purchase Price:</strong> {appliance.purchasePrice}
                  <br />
                </Card.Text>
                <Row>
                  <Col>
                    <Button
                      variant="secondary"
                      style={{ marginTop: "10px" }}
                      onClick={() => setShowEditModal(true)}
                    >
                      <i className="bi bi-pencil-fill"></i>
                    </Button>
                    <Button
                      variant="danger"
                      onClick={handleDelete}
                      style={{ marginTop: "10px", float: "right" }}
                    >
                      <i className="bi bi-trash"></i>
                    </Button>
                  </Col>
                </Row>
              </Card.Body>
            </Card>
          </Tab>
          <Tab eventKey="maintenance" title="Maintenance">
            <MaintenanceSection
              applianceId={appliance.id}
              referenceType="Appliance"
              spaceType="NotApplicable"
            />
          </Tab>
          <Tab eventKey="repairs" title="Repairs">
            <RepairSection applianceId={appliance.id} />
          </Tab>
          <Tab eventKey="documentation" title="Documentation">
            <DocumentationSection applianceId={appliance.id} />
          </Tab>
          <Tab eventKey="todos" title="Tasks">
            <TasksSection applianceId={appliance.id} />
          </Tab>
          <Tab eventKey="notes" title="Notes">
            <NotesSection applianceId={appliance.id} />
          </Tab>
        </Tabs>
        <Button
          variant="secondary"
          onClick={() => navigate("/appliances")}
          style={{ marginTop: "10px" }}
        >
          Back
        </Button>
      </Col>
      <EditApplianceModal
        show={showEditModal}
        handleClose={() => setShowEditModal(false)}
        handleSave={handleSaveAppliance}
        appliance={appliance}
      />
    </Row>
  );
};

export default AppliancePage;

import React, { useEffect, useState } from "react";
import { Card, Table } from "react-bootstrap";
import { SERVER_URL } from "@/context/DemoContext";
import AddRepairModal from "@/components/AddRepairModal";
import ShowRepairModal from "@/components/ShowRepairModal";

export type RepairRecord = {
  id: number;
  description: string;
  date: string;
  cost: number;
  notes: string;
  resolved: boolean;
};

export type RepairReferenceType = "Appliance" | "Space";

export type RepairSpaceType =
  | "BuildingExterior"
  | "BuildingInterior"
  | "Electrical"
  | "HVAC"
  | "Plumbing"
  | "Yard"
  | "NotApplicable";

interface RepairSectionProps {
  applianceId?: number;
  spaceType?: RepairSpaceType;
}

const RepairSection: React.FC<RepairSectionProps> = ({
  applianceId,
  spaceType,
}) => {
  const [records, setRecords] = useState<RepairRecord[]>([]);
  const [showAdd, setShowAdd] = useState(false);
  const [showView, setShowView] = useState(false);
  const [selected, setSelected] = useState<RepairRecord | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const referenceType = spaceType ? "Space" : "Appliance";
        const queryParams = new URLSearchParams({
          applianceId: applianceId?.toString() || "",
          referenceType,
          spaceType: spaceType || "",
        }).toString();
        const resp = await fetch(`${SERVER_URL}/repair?${queryParams}`);
        if (!resp.ok) return;
        const data: RepairRecord[] = await resp.json();
        setRecords(data);
      } catch (err) {
        console.error("Error loading repairs", err);
      }
    };
    load();
  }, [applianceId, spaceType]);

  const handleShowAdd = () => setShowAdd(true);
  const handleCloseAdd = () => setShowAdd(false);
  const handleSave = (saved: RepairRecord) =>
    setRecords((prev) => [...prev, saved]);

  const handleRowClick = (r: RepairRecord) => {
    setSelected(r);
    setShowView(true);
  };

  const handleCloseView = () => setShowView(false);

  const handleDeleteRepair = (id: number) =>
    setRecords((prev) => prev.filter((r) => r.id !== id));

  const handleUpdateRepair = (updated: RepairRecord) => {
    setRecords((prev) => prev.map((r) => (r.id === updated.id ? updated : r)));
    setSelected(updated);
  };

  const totalCost = records.reduce((s, r) => s + r.cost, 0);

  return (
    <Card>
      <Card.Body>
        <Table striped bordered hover>
          <thead>
            <tr>
              <th>Description</th>
              <th>Cost</th>
              <th>Date</th>
            </tr>
          </thead>
          <tbody>
            {records.length === 0 ? (
              <tr>
                <td colSpan={3} style={{ textAlign: "center" }}>
                  No repairs recorded
                </td>
              </tr>
            ) : (
              records.map((r) => (
                <tr
                  key={r.id}
                  onClick={() => handleRowClick(r)}
                  style={{ cursor: "pointer" }}
                >
                  <td>{r.description}</td>
                  <td>{r.cost}</td>
                  <td>{r.date}</td>
                </tr>
              ))
            )}
          </tbody>
        </Table>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <i
            className="bi bi-plus-square-fill"
            style={{ fontSize: "2rem", cursor: "pointer" }}
            onClick={handleShowAdd}
          ></i>
          <div style={{ fontWeight: "bold" }}>
            Total Repair Cost: ${totalCost}
          </div>
        </div>
      </Card.Body>
      <AddRepairModal
        show={showAdd}
        handleClose={handleCloseAdd}
        handleSave={handleSave}
        applianceId={applianceId}
        referenceType={applianceId ? "Appliance" : "Space"}
        spaceType={(spaceType as RepairSpaceType) ?? "NotApplicable"}
      />
      {selected && (
        <ShowRepairModal
          show={showView}
          handleClose={handleCloseView}
          repairRecord={selected}
          handleDeleteRepair={handleDeleteRepair}
          handleUpdateRepair={handleUpdateRepair}
        />
      )}
    </Card>
  );
};

export default RepairSection;

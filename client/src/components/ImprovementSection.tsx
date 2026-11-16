import React, { useEffect, useState } from "react";
import { Card, Table } from "react-bootstrap";
import { SERVER_URL } from "@/context/DemoContext";
import AddRepairModal from "@/components/AddRepairModal";
import ShowRepairModal from "@/components/ShowRepairModal";
import type { RepairRecord, RepairSpaceType } from "@/components/RepairSection";
import { useCurrency } from "@/context/useCurrency";
import { formatCurrency } from "@/utils/currency";

export type ImprovementRecord = Omit<RepairRecord, "resolved">;

interface ImprovementSectionProps {
  applianceId?: number;
  spaceType?: RepairSpaceType;
}

const ImprovementSection: React.FC<ImprovementSectionProps> = ({
  applianceId,
  spaceType,
}) => {
  const { currency } = useCurrency();
  const [records, setRecords] = useState<ImprovementRecord[]>([]);
  const [showAdd, setShowAdd] = useState(false);
  const [showView, setShowView] = useState(false);
  const [selected, setSelected] = useState<ImprovementRecord | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const referenceType = spaceType ? "Space" : "Appliance";
        const queryParams = new URLSearchParams({
          applianceId: applianceId?.toString() || "",
          referenceType,
          spaceType: spaceType || "",
        }).toString();
        const response = await fetch(
          `${SERVER_URL}/improvement?${queryParams}`,
        );
        if (!response.ok) return;
        setRecords(await response.json());
      } catch (error) {
        console.error("Error loading improvements:", error);
      }
    };
    void load();
  }, [applianceId, spaceType]);

  const handleUpdate = (updated: RepairRecord) => {
    const improvement = updated as ImprovementRecord;
    setRecords((previous) =>
      previous.map((record) =>
        record.id === improvement.id ? improvement : record,
      ),
    );
    setSelected(improvement);
  };

  const totalCost = records.reduce((sum, record) => sum + record.cost, 0);

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
                <td colSpan={3} className="text-center">
                  No improvements recorded
                </td>
              </tr>
            ) : (
              records.map((record) => (
                <tr
                  key={record.id}
                  onClick={() => {
                    setSelected(record);
                    setShowView(true);
                  }}
                  style={{ cursor: "pointer" }}
                >
                  <td>{record.description}</td>
                  <td>{formatCurrency(record.cost, currency)}</td>
                  <td>{record.date}</td>
                </tr>
              ))
            )}
          </tbody>
        </Table>
        <div
          className="d-flex justify-content-between align-items-center"
          style={{ fontWeight: "bold" }}
        >
          <i
            className="bi bi-plus-square-fill"
            style={{ fontSize: "2rem", cursor: "pointer" }}
            onClick={() => setShowAdd(true)}
            role="button"
            aria-label="Add improvement"
          />
          <div>
            Total Improvement Cost: {formatCurrency(totalCost, currency)}
          </div>
        </div>
      </Card.Body>
      <AddRepairModal
        show={showAdd}
        handleClose={() => setShowAdd(false)}
        handleSave={(record) => setRecords((previous) => [...previous, record])}
        applianceId={applianceId}
        referenceType={spaceType ? "Space" : "Appliance"}
        spaceType={spaceType ?? "NotApplicable"}
        recordKind="improvement"
      />
      {selected && (
        <ShowRepairModal
          show={showView}
          handleClose={() => setShowView(false)}
          repairRecord={selected}
          handleDeleteRepair={(id) =>
            setRecords((previous) =>
              previous.filter((record) => record.id !== id),
            )
          }
          handleUpdateRepair={handleUpdate}
          recordKind="improvement"
        />
      )}
    </Card>
  );
};

export default ImprovementSection;

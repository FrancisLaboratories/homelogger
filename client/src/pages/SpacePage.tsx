import React from "react";
import { Tab, Tabs } from "react-bootstrap";
import MaintenanceSection from "@/components/MaintenanceSection";
import type { MaintenanceSpaceType } from "@/components/MaintenanceSection";
import RepairSection from "@/components/RepairSection";
import DocumentationSection from "@/components/DocumentationSection";
import TasksSection from "@/components/TasksSection";
import NotesSection from "@/components/NotesSection";

interface Props {
  spaceType: MaintenanceSpaceType;
  title: string;
}

const SpacePage: React.FC<Props> = ({ spaceType, title }) => {
  return (
    <>
      <h3>{title}</h3>
      <Tabs defaultActiveKey="maintenance" id="space-tabs" className="mb-3">
        <Tab eventKey="maintenance" title="Maintenance">
          <MaintenanceSection referenceType="Space" spaceType={spaceType} />
        </Tab>
        <Tab eventKey="repair" title="Repairs">
          <RepairSection spaceType={spaceType} />
        </Tab>
        <Tab eventKey="documents" title="Documents">
          <DocumentationSection spaceType={spaceType} />
        </Tab>
        <Tab eventKey="notes" title="Notes">
          <NotesSection spaceType={spaceType} />
        </Tab>
        <Tab eventKey="todos" title="Tasks">
          <TasksSection spaceType={spaceType} />
        </Tab>
      </Tabs>
    </>
  );
};

export default SpacePage;

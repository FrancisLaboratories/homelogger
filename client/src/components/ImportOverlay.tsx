import React, { useContext } from "react";
import { Modal, Spinner } from "react-bootstrap";
import { ImportContext } from "@/context/ImportContext";

const ImportOverlay: React.FC = () => {
  const { isImporting } = useContext(ImportContext);

  return (
    <Modal show={isImporting} backdrop="static" keyboard={false}>
      <Modal.Body className="text-center py-5">
        <Spinner animation="border" role="status" />
        <p className="mt-3 mb-1 fw-bold">Import in progress...</p>
        <p className="text-muted small mb-0">
          Please do not close or navigate away from this page.
        </p>
      </Modal.Body>
    </Modal>
  );
};

export default ImportOverlay;

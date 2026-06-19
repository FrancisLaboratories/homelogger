import React, { useState, type ChangeEvent } from "react";
import { Button, Modal } from "react-bootstrap";

import { SERVER_URL } from "@/context/DemoContext";

const SettingsPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [importLoading, setImportLoading] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [showFileSelectionModal, setShowFileSelectionModal] = useState(false);
  const [showOverwriteConfirmModal, setShowOverwriteConfirmModal] = useState(false);

  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files.length > 0) {
      setSelectedFile(event.target.files[0]);
    } else {
      setSelectedFile(null);
    }
  };

  const handleOpenImportModal = () => {
    setShowFileSelectionModal(true);
    setSelectedFile(null); // Clear any previously selected file
  };

  const handleCloseFileSelectionModal = () => {
    setShowFileSelectionModal(false);
  };

  const handleImportBackupClick = () => {
    // This will now be the trigger for the confirmation modal, not the import itself
    if (!selectedFile) {
      alert("Please select a backup file to import.");
      return;
    }
    setShowOverwriteConfirmModal(true);
  };

  const handleConfirmImport = async () => {
    setShowOverwriteConfirmModal(false);
    setShowFileSelectionModal(false); // Close file selection modal too
    setImportLoading(true);
    try {
      const formData = new FormData();
      if (selectedFile) {
        formData.append("backup", selectedFile);
      } else {
        throw new Error("No file selected for backup.");
      }

      const res = await fetch(`${SERVER_URL}/backup/import`, {
        method: "POST",
        body: formData,
      });

      if (!res.ok) {
        const errorText = await res.text();
        throw new Error(`Failed to import backup: ${errorText}`);
      }

      alert("Backup imported successfully! The server may restart.");
      setSelectedFile(null);
    } catch (err: any) {
      console.error(err);
      alert(`Error importing backup: ${err.message || "See console for details."}`);
    } finally {
      setImportLoading(false);
    }
  };

  const handleCancelImport = () => {
    setShowOverwriteConfirmModal(false);
  };

  const handleDownloadBackup = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${SERVER_URL}/backup/download`);
      if (!res.ok) throw new Error("Failed to download backup");

      const blob = await res.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      const timestamp = new Date()
        .toISOString()
        .slice(0, 19)
        .replaceAll(":", "-");
      a.download = `homelogger-backup-${timestamp}.zip`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } catch (err) {
      console.error(err);
      alert("Error downloading backup. See console for details.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h4 id="maintext" style={{ marginTop: "1rem" }}>
        Settings
      </h4>
      <div style={{ marginTop: "1rem" }}>
        <p>Download a backup of the database and uploaded files.</p>
        <Button
          onClick={handleDownloadBackup}
          disabled={loading}
          variant="primary"
        >
          {loading ? "Preparing backup..." : "Download Backup"}
        </Button>
      </div>

      <div style={{ marginTop: "2rem" }}>
        <p>Import a backup ZIP file to restore your database and uploaded files. This will wipe and replace all existing data.</p>
        <Button
          onClick={handleOpenImportModal}
          variant="danger"
        >
          Import Backup
        </Button>
      </div>

      <Modal show={showOverwriteConfirmModal} onHide={handleCancelImport} backdrop="static" keyboard={false}>
        <Modal.Header closeButton>
          <Modal.Title>Confirm Backup Import</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p><strong>Warning:</strong> Importing a backup will irreversibly delete ALL existing database records and uploaded files, replacing them with the contents of the backup.</p>
          <p>Are you sure you want to proceed?</p>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={handleCancelImport}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleConfirmImport}>
            I Understand, Import Backup
          </Button>
        </Modal.Footer>
      </Modal>

      <Modal show={showFileSelectionModal} onHide={handleCloseFileSelectionModal} backdrop="static" keyboard={false}>
        <Modal.Header closeButton>
          <Modal.Title>Upload Backup File</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>Select a backup ZIP file to import. This will replace all existing data.</p>
          <input type="file" accept=".zip" onChange={handleFileChange} />
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={handleCloseFileSelectionModal}>
            Cancel
          </Button>
          <Button
            onClick={handleImportBackupClick}
            disabled={importLoading || !selectedFile}
            variant="danger"
          >
            {importLoading ? "Importing..." : "Import Data"}
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
};

export default SettingsPage;

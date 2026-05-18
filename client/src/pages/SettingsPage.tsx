import React, { useState } from "react";
import { Button } from "react-bootstrap";
import { SERVER_URL } from "@/context/DemoContext";

const SettingsPage: React.FC = () => {
  const [loading, setLoading] = useState(false);

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
    </div>
  );
};

export default SettingsPage;

package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportUploads(t *testing.T) {
	t.Run("empty path returns nil", func(t *testing.T) {
		if err := ImportUploads(""); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("successful swap replaces uploads", func(t *testing.T) {
		origDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}
		testDir := t.TempDir()
		if err := os.Chdir(testDir); err != nil {
			t.Fatalf("chdir: %v", err)
		}
		defer os.Chdir(origDir)

		stageDir := filepath.Join(testDir, "stage-uploads")
		if err := os.MkdirAll(stageDir, 0755); err != nil {
			t.Fatalf("mkdir stage: %v", err)
		}

		// Create a file in the stage dir
		stageFile := filepath.Join(stageDir, "photo.jpg")
		if err := os.WriteFile(stageFile, []byte("photo-data"), 0644); err != nil {
			t.Fatalf("write stage file: %v", err)
		}

		// Create existing uploads with different content
		oldUploads := "./data/uploads"
		if err := os.MkdirAll(oldUploads, 0755); err != nil {
			t.Fatalf("mkdir old uploads: %v", err)
		}
		oldFile := filepath.Join(oldUploads, "old-doc.txt")
		if err := os.WriteFile(oldFile, []byte("old-data"), 0644); err != nil {
			t.Fatalf("write old file: %v", err)
		}

		if err := ImportUploads(stageDir); err != nil {
			t.Fatalf("ImportUploads failed: %v", err)
		}

		// Verify old file is gone
		if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
			t.Error("expected old upload file to be removed")
		}

		// Verify new file is in place
		newFile := filepath.Join(oldUploads, "photo.jpg")
		data, err := os.ReadFile(newFile)
		if err != nil {
			t.Fatalf("read new file: %v", err)
		}
		if string(data) != "photo-data" {
			t.Errorf("expected 'photo-data', got %q", string(data))
		}

		// Verify .bak was cleaned up
		if _, err := os.Stat("./data/uploads.bak"); !os.IsNotExist(err) {
			t.Error("expected .bak to be removed after successful swap")
		}
	})

	t.Run("no existing uploads creates new directory", func(t *testing.T) {
		origDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}
		testDir := t.TempDir()
		if err := os.Chdir(testDir); err != nil {
			t.Fatalf("chdir: %v", err)
		}
		defer os.Chdir(origDir)

		// ImportUploads creates temp dir under ./data, so ./data must exist
		if err := os.MkdirAll("./data", 0755); err != nil {
			t.Fatalf("mkdir ./data: %v", err)
		}

		stageDir := filepath.Join(testDir, "stage-uploads")
		if err := os.MkdirAll(stageDir, 0755); err != nil {
			t.Fatalf("mkdir stage: %v", err)
		}

		stageFile := filepath.Join(stageDir, "doc.pdf")
		if err := os.WriteFile(stageFile, []byte("pdf-data"), 0644); err != nil {
			t.Fatalf("write stage file: %v", err)
		}

		if err := ImportUploads(stageDir); err != nil {
			t.Fatalf("ImportUploads failed: %v", err)
		}

		// Verify the uploads directory was created
		newFile := filepath.Join("./data/uploads", "doc.pdf")
		data, err := os.ReadFile(newFile)
		if err != nil {
			t.Fatalf("read file: %v", err)
		}
		if string(data) != "pdf-data" {
			t.Errorf("expected 'pdf-data', got %q", string(data))
		}
	})
}

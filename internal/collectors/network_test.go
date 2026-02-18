package collectors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBaselineRoundTrip(t *testing.T) {
	// Use a temp dir to avoid polluting the real baseline directory
	tmpDir := t.TempDir()
	date := time.Now().Format("2006-01-02")
	path := filepath.Join(tmpDir, "network-"+date+".json")

	b := networkBaseline{
		Interface:     "en0",
		BytesReceived: 12345,
		BytesSent:     67890,
		Timestamp:     time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	readData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	var loaded networkBaseline
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if loaded.Interface != "en0" {
		t.Errorf("Interface = %q, want %q", loaded.Interface, "en0")
	}
	if loaded.BytesReceived != 12345 {
		t.Errorf("BytesReceived = %d, want %d", loaded.BytesReceived, 12345)
	}
	if loaded.BytesSent != 67890 {
		t.Errorf("BytesSent = %d, want %d", loaded.BytesSent, 67890)
	}
}

func TestBaselineCorruptedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "network-corrupt.json")

	if err := os.WriteFile(path, []byte("not valid json{{{"), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	var b networkBaseline
	err = json.Unmarshal(data, &b)
	if err == nil {
		t.Error("expected error for corrupted JSON, got nil")
	}
}

func TestCleanOldBaselines(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with old and recent dates
	oldFile := filepath.Join(tmpDir, "network-2020-01-01.json")
	recentFile := filepath.Join(tmpDir, "network-"+time.Now().Format("2006-01-02")+".json")
	unrelatedFile := filepath.Join(tmpDir, "other-file.txt")

	for _, f := range []string{oldFile, recentFile, unrelatedFile} {
		if err := os.WriteFile(f, []byte("{}"), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", f, err)
		}
	}

	cleanOldBaselines(tmpDir)

	// Old file should be deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("old baseline file should have been deleted")
	}

	// Recent file should still exist
	if _, err := os.Stat(recentFile); err != nil {
		t.Error("recent baseline file should still exist")
	}

	// Unrelated file should still exist
	if _, err := os.Stat(unrelatedFile); err != nil {
		t.Error("unrelated file should still exist")
	}
}

package caseidpathtraversal

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/persistence"
)

func TestLoadRejectsCaseIDPathTraversal(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	// This file is outside the cases directory and must never be addressable as a case.
	data := []byte(`{"case":{"id":"victim","status":"draft","revision":1},"requests":[]}`)
	if err := os.WriteFile(filepath.Join(dir, "victim.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(context.Background(), "../victim"); err == nil {
		t.Fatal("path traversal loaded snapshot outside cases directory")
	}
}

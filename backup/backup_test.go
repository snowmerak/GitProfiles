package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupRestore(t *testing.T) {
	d := t.TempDir()
	// create sample files
	os.MkdirAll(filepath.Join(d, "keys"), 0o700)
	os.WriteFile(filepath.Join(d, "keys", "a"), []byte("secret"), 0o600)
	os.MkdirAll(filepath.Join(d, "meta"), 0o700)
	os.WriteFile(filepath.Join(d, "meta", "keys.json"), []byte("{}"), 0o600)

	out := filepath.Join(d, "backup.gpbk")
	pass := []byte("password123")
	// make params smaller for speed in test
	oldN := ScryptN
	ScryptN = 1 << 10
	defer func() { ScryptN = oldN }()

	if err := Backup(d, out, pass); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// quick check: read file header
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read backup file: %v", err)
	}
	if len(b) < 4 || string(b[:4]) != magic {
		t.Fatalf("invalid magic in backup")
	}

	// restore to new dir
	d2 := filepath.Join(d, "restored")
	if err := Restore(out, d2, pass); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	// check restored file
	b, err = os.ReadFile(filepath.Join(d2, "keys", "a"))
	if err != nil {
		t.Fatalf("missing restored file: %v", err)
	}
	if string(b) != "secret" {
		t.Fatalf("restored content mismatch: %s", string(b))
	}
}

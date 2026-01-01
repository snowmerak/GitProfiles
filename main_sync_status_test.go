package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPreviewSSHConfig(t *testing.T) {
	d := t.TempDir()
	metaDir := filepath.Join(d, "meta")
	os.MkdirAll(metaDir, 0o700)

	meta := map[string]map[string]string{
		"alice": {
			"algo":    "ed25519",
			"private": filepath.Join(d, "keys", "alice_id_ed25519"),
			"public":  filepath.Join(d, "keys", "alice_id_ed25519.pub"),
			"email":   "alice@example.com",
			"host":    "github.com",
		},
		"bob": {
			"algo":    "ed25519",
			"private": filepath.Join(d, "keys", "bob_id_ed25519"),
			"public":  filepath.Join(d, "keys", "bob_id_ed25519.pub"),
			"email":   "bob@example.com",
			// no host
		},
	}
	b, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(metaDir, "keys.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}

	// create key files for alice
	os.MkdirAll(filepath.Join(d, "keys"), 0o700)
	os.WriteFile(filepath.Join(d, "keys", "alice_id_ed25519"), []byte("PRIVATE"), 0o600)

	cfg := filepath.Join(d, "config")
	// empty config
	os.WriteFile(cfg, []byte(""), 0o600)

	adds, removes, err := PreviewSSHConfig(d, cfg, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(adds) != 1 {
		t.Fatalf("expected 1 add, got %#v", adds)
	}
	if len(removes) != 0 {
		t.Fatalf("expected 0 removes, got %#v", removes)
	}
	if adds[0].Alias != "git-alice-github-com" {
		t.Fatalf("unexpected alias: %s", adds[0].Alias)
	}
}

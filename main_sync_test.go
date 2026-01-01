package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSyncSSHConfig(t *testing.T) {
	d := t.TempDir()
	metaDir := filepath.Join(d, "meta")
	os.MkdirAll(metaDir, 0o700)

	// create two profiles: one with host and one without
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
	// add a stale entry to cfg
	os.WriteFile(cfg, []byte("# BEGIN GITPROFILES git-stale\nHost git-stale\n    HostName oldhost\n# END GITPROFILES git-stale\n"), 0o600)

	// run sync with prune = true
	if err := SyncSSHConfig(d, cfg, true); err != nil {
		t.Fatal(err)
	}
	b2, _ := os.ReadFile(cfg)
	s := string(b2)
	if !stringsContains(s, "git-alice-github-com") {
		t.Fatalf("expected alice alias present, got: %s", s)
	}
	if stringsContains(s, "git-stale") {
		t.Fatalf("expected stale alias removed, got: %s", s)
	}

	// test preview/status
	adds, removes, err := PreviewSSHConfig(d, cfg, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(adds) != 0 { // already synced, no adds
		t.Fatalf("expected no adds after sync, got: %#v", adds)
	}
	for _, r := range removes {
		if r == "git-stale" {
			t.Fatalf("expected stale already removed, got remove: %s", r)
		}
	}
}

func stringsContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (s != "" && (len(sub) == 0 || len(s) >= len(sub) && (s != "" && (indexOf(s, sub) >= 0)))))
}

func indexOf(s, sub string) int {
	for i := range s {
		if i+len(sub) <= len(s) && s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

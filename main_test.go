package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestInitAndAdd(t *testing.T) {
	dir := os.TempDir()
	defer os.RemoveAll(dir)

	if err := Init(dir); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// add a key
	priv, pub, err := Add(dir, "ed25519", "alice", "alice@example.com", "github.com")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// check files exist
	if _, err := os.Stat(priv); err != nil {
		t.Fatalf("private key not found: %v", err)
	}
	if _, err := os.Stat(pub); err != nil {
		t.Fatalf("public key not found: %v", err)
	}

	b, err := os.ReadFile(priv)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "PRIVATE KEY") {
		t.Fatalf("private key content invalid: %s", string(b))
	}

	// public key parse
	pubb, err := os.ReadFile(pub)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, _, err := ssh.ParseAuthorizedKey(pubb); err != nil {
		t.Fatalf("invalid public key: %v", err)
	}

	// check meta
	metaPath := filepath.Join(dir, "meta", "keys.json")
	mb, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mb), "alice") {
		t.Fatalf("meta does not contain alice: %s", string(mb))
	}
}

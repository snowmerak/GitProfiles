package key

import (
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestGenerators(t *testing.T) {
	tests := []struct {
		alg string
		gen KeyGenerator
	}{
		{RSA2048, generators[RSA2048]},
		{RSA4096, generators[RSA4096]},
		{P256, generators[P256]},
		{P384, generators[P384]},
		{P521, generators[P521]},
		{ED25519, generators[ED25519]},
	}

	for _, tt := range tests {
		t.Run(tt.alg, func(t *testing.T) {
			priv, pub, err := tt.gen.Generate("alice", "example.com")
			if err != nil {
				t.Fatalf("Generate error for %s: %v", tt.alg, err)
			}
			if !strings.Contains(priv, "PRIVATE KEY") {
				t.Fatalf("private key does not contain 'PRIVATE KEY' for %s: %s", tt.alg, priv)
			}
			if !strings.Contains(pub, "alice@example.com") {
				t.Fatalf("pub key does not contain identity for %s: %s", tt.alg, pub)
			}
			if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pub)); err != nil {
				t.Fatalf("invalid public key for %s: %v", tt.alg, err)
			}
		})
	}
}

package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestTarGzipRoundtrip(t *testing.T) {
	d := t.TempDir()
	os.MkdirAll(filepath.Join(d, "keys"), 0o700)
	os.WriteFile(filepath.Join(d, "keys", "a"), []byte("secret"), 0o600)
	data, err := createTarGzip(d)
	if err != nil {
		t.Fatalf("createTarGzip failed: %v", err)
	}
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("gzip read failed: %v", err)
	}
	tr := tar.NewReader(gr)
	found := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar read error: %v", err)
		}
		if hdr.Name == "keys/a" {
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, tr); err != nil {
				t.Fatalf("read content fail: %v", err)
			}
			if buf.String() != "secret" {
				t.Fatalf("content mismatch: %s", buf.String())
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("didn't find keys/a in tar")
	}
}

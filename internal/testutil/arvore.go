package testutil

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// CopiaArvore copia src para dst (criado se preciso), preservando a estrutura e os bytes.
func CopiaArvore(t testing.TB, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		alvo := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(alvo, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(alvo, b, 0o644)
	})
	if err != nil {
		t.Fatalf("copiar %s: %v", src, err)
	}
}

// SnapshotSHA devolve caminho relativo (com /) para o sha256 de cada arquivo regular de dir.
func SnapshotSHA(t testing.TB, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		h := sha256.Sum256(b)
		out[filepath.ToSlash(rel)] = hex.EncodeToString(h[:])
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", dir, err)
	}
	return out
}

// ReadFile lê root/rel e falha o teste se der erro.
func ReadFile(t testing.TB, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("ler %s: %v", rel, err)
	}
	return string(b)
}

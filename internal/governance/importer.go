package governance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ModeBloco          = "bloco"
	ModeSubstituir     = "substituir"
	ModeCriarSeAusente = "criar_se_ausente"
	ModeMesclar        = "mesclar"
	ModeCriarUmaVez    = "criar_uma_vez"
)

// ApplyBlock aplica de forma idempotente um bloco demarcado (ex: <!-- pkg-governanca:inicio --> ... <!-- pkg-governanca:fim -->)
// ao conteúdo de sdd/memory/constitution.md, preservando todo o restante byte a byte.
func ApplyBlock(existingContent, startMarkerPrefix, endMarker, newBlock string) string {
	newBlock = strings.TrimSpace(newBlock)
	startIdx := strings.Index(existingContent, startMarkerPrefix)
	if startIdx != -1 {
		endIdx := strings.Index(existingContent[startIdx:], endMarker)
		if endIdx != -1 {
			absEnd := startIdx + endIdx + len(endMarker)
			before := existingContent[:startIdx]
			after := existingContent[absEnd:]
			return before + newBlock + after
		}
	}

	trimmed := strings.TrimRight(existingContent, "\n")
	if trimmed == "" {
		return newBlock + "\n"
	}
	return trimmed + "\n\n" + newBlock + "\n"
}

// ApplyFile aplica um arquivo de pacote de governança conforme o modo do manifest.yaml.
// Retorna a ação executada: "created", "replaced", "merged", "skipped".
func ApplyFile(targetDir, dstRel, mode string, srcContent []byte) (string, error) {
	dest := filepath.Join(targetDir, dstRel)
	_, statErr := os.Stat(dest)
	exists := statErr == nil

	switch mode {
	case ModeSubstituir:
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return "", fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err)
		}
		if err := os.WriteFile(dest, srcContent, 0644); err != nil {
			return "", fmt.Errorf("write %s: %w", dest, err)
		}
		if exists {
			return "replaced", nil
		}
		return "created", nil

	case ModeCriarSeAusente, ModeCriarUmaVez:
		if exists {
			return "skipped", nil
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return "", fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err)
		}
		if err := os.WriteFile(dest, srcContent, 0644); err != nil {
			return "", fmt.Errorf("write %s: %w", dest, err)
		}
		return "created", nil

	case ModeMesclar:
		if !exists {
			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				return "", fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err)
			}
			if err := os.WriteFile(dest, srcContent, 0644); err != nil {
				return "", fmt.Errorf("write %s: %w", dest, err)
			}
			return "created", nil
		}
		existing, err := os.ReadFile(dest)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", dest, err)
		}
		merged := mergeMarkdownSections(string(existing), string(srcContent))
		if err := os.WriteFile(dest, []byte(merged), 0644); err != nil {
			return "", fmt.Errorf("write %s: %w", dest, err)
		}
		return "merged", nil

	default:
		return "", fmt.Errorf("modo de importação desconhecido: %q", mode)
	}
}

func mergeMarkdownSections(existing, incoming string) string {
	if strings.Contains(existing, strings.TrimSpace(incoming)) {
		return existing
	}
	return strings.TrimRight(existing, "\n") + "\n\n" + strings.TrimSpace(incoming) + "\n"
}

// WriteVersionRecord grava o registro de importação em docs/governanca/VERSAO.
func WriteVersionRecord(targetDir, pkgName, pkgVersion, commitSHA string, now time.Time) (string, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	dest := filepath.Join(targetDir, "docs", "governanca", "VERSAO")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err)
	}
	content := fmt.Sprintf("pacote: %s\nversao: %s\nsha: %s\ndata: %s\n",
		pkgName, pkgVersion, commitSHA, now.UTC().Format(time.RFC3339))
	if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write %s: %w", dest, err)
	}
	return dest, nil
}

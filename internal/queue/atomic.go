package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileAtomic grava dados em um arquivo temporário no mesmo diretório
// e em seguida realiza a substituição atômica via os.Rename.
func WriteFileAtomic(targetPath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("falha ao criar diretorio %s: %w", dir, err)
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo temporario em %s: %w", dir, err)
	}
	tmpPath := tmpFile.Name()

	// Garante remoção em caso de erro antes do rename
	defer func() {
		if tmpPath != "" {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("falha ao escrever em %s: %w", tmpPath, err)
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("falha ao sincronizar %s: %w", tmpPath, err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("falha ao fechar %s: %w", tmpPath, err)
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("falha ao definir permissoes em %s: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("falha ao renomear %s -> %s: %w", tmpPath, targetPath, err)
	}

	tmpPath = "" // rename bem-sucedido, cancela defer remove
	return nil
}

// WriteJSONAtomic serializa e grava uma estrutura em JSON de forma indentada e atômica.
func WriteJSONAtomic(targetPath string, v any, perm os.FileMode) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("falha ao serializar JSON para %s: %w", targetPath, err)
	}
	data = append(data, '\n')
	return WriteFileAtomic(targetPath, data, perm)
}

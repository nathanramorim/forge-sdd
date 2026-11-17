package config

import "sync/atomic"

var binaryVersion atomic.Value

// SetBinaryVersion registra a versão do binário (injetada por ldflags em main.version).
func SetBinaryVersion(v string) {
	binaryVersion.Store(v)
}

// BinaryVersion devolve a versão do binário; "dev" quando ninguém a registrou. É a única fonte
// de versão em telemetria, ACP e defaults (DUP-04).
func BinaryVersion() string {
	if v, _ := binaryVersion.Load().(string); v != "" {
		return v
	}
	return "dev"
}

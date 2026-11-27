// Package eol normaliza quebras de linha de arquivos lidos do disco. No Windows o checkout do git
// pode gravar CRLF; o forge compara e interpreta o conteúdo sempre em LF.
package eol

import "bytes"

var crlf, lf = []byte("\r\n"), []byte("\n")

// LF troca CRLF por LF. Sem CRLF, devolve o próprio slice (sem cópia).
func LF(b []byte) []byte {
	if !bytes.Contains(b, crlf) {
		return b
	}
	return bytes.ReplaceAll(b, crlf, lf)
}

// Iguais compara dois conteúdos ignorando a diferença entre CRLF e LF.
func Iguais(a, b []byte) bool {
	return bytes.Equal(LF(a), LF(b))
}

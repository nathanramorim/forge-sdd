---
description: Como manter qualidade de código Go, higiene de pendências, cobertura de testes executáveis e sincronia de golden files.
globs: ["**/*.go", "**/*_test.go"]
regras: [COD-01, COD-02, COD-04, COD-06, TST-01, TST-02, TST-05]
---

# Guia Prático — Código Limpo e Testes Executáveis

Rotina técnica para atender aos IDs `COD-01`, `COD-02`, `COD-04`, `COD-06`, `TST-01`, `TST-02` e `TST-05` de `sdd/memory/constitution.md` (`GOV-04`).

## 1. Higiene de Código e Análise Estática (`COD-01`, `COD-02`, `COD-04`, `COD-06`)

- **Coesão e remoção de resíduos (`COD-01`, `COD-04`):**
  - Mantenha funções focadas em uma única etapa de transformação ou validação.
  - Ao substituir um template ou comando antigo, remova as referências órfãs e adicione o caminho descontinuado em `cleanObsoleteFiles` quando necessário para migração limpa.
- **Rastreabilidade de pendências (`COD-06`):**
  - Nunca deixe `// TODO:` genérico. Caso uma pendência futura seja intencional e documentada na Spec, anote no formato `// TODO(feat-XX): descrição objetiva`.
- **Verificação com `go vet` (`COD-02`):**
  - Execute `go vet ./...` ao concluir qualquer edição em arquivos `.go` e corrija imediatamente alertas de formatação de strings, variáveis não utilizadas ou construções suspeitas.

## 2. Testes Automatizados e Golden Files (`TST-01`, `TST-02`, `TST-05`)

- **Desenvolvimento orientado aos critérios da Spec (`TST-01`):**
  - Para cada critério de aceitação da task, escreva ou estenda um teste `Test*` em `*_test.go` usando `t.TempDir()` para isolar operações de sistema de arquivos e `testify/require` + `testify/assert` para asserções claras.
  - Inclua testes tanto para entradas válidas quanto para violações esperadas (ex.: JSON inválido, arquivo ausente, regra violada).
- **Sincronização de templates e Golden Files (`TST-02`):**
  - Sempre que criar ou editar arquivos sob `internal/scaffold/templates/`, atualize os golden files correspondentes com `go test ./internal/scaffold/... -update` e inspecione o `git diff internal/scaffold/testdata/golden/` para confirmar que a saída gerada reflete exatamente o esperado.
- **Validação bloqueante antes de concluir a task (`TST-05`):**
  - A task só pode avançar para `review_pending` na fila quando `go test ./...` e `go vet ./...` terminarem com código de saída `0`.

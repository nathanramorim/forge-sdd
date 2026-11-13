# feat-63-01 — Endpoint de Leitura de Spec (`GET /api/spec`) e Sanitização Estrita contra Path Traversal

**Branch:** `feat/63-studio-navegavel-aprovacao-spec`
**Fase:** 63-01
**Tamanho:** `P`
**Depende de:** —
**Status:** `review_pending`

## Objetivo

Implementar no servidor HTTP do Forge Studio (`internal/studio/server.go`) o endpoint seguro `GET /api/spec?path=<path>`:
1. Receber o parâmetro de consulta `path` e realizar validação estrita na borda (`SEG-01`):
   - Sanitizar o caminho via `filepath.Clean`.
   - Resolver o caminho absoluto ancorado no diretório-base do projeto (`s.Dir`).
   - Bloquear tentativas de *Path Traversal* (`..`) e rejeitar caminhos fora do diretório do projeto com código HTTP `400 Bad Request` ou `403 Forbidden` (`SEG-02`).
   - Assegurar que apenas arquivos de especificações e documentos sob pastas SDD (`sdd/features/`, `sdd/discovery/`, `sdd/spec/`) possam ser lidos, impedindo a exposição acidental de segredos, tokens ou arquivos de sistema (`SEG-04`).
   - Retornar `404 Not Found` caso o arquivo especificado não exista no disco.
2. Formatar e retornar payload JSON determinístico contendo:
   - `path`: caminho relativo normalizado (ex: `sdd/features/feat-63-studio-navegavel-aprovacao-spec/discovery.md`).
   - `absolute_path`: caminho absoluto completo no sistema de arquivos local para possibilitar deep-links com a IDE.
   - `content`: conteúdo integral do arquivo Markdown em UTF-8.
   - `status`: status atual inferido ou mapeado da tarefa/spec (`pending_review`, `approved`, etc.).
   - `station`: estação responsável pela spec (default `pm-architect`).
   - `feature_id`: ID canônico da feature.
3. Desenvolver suíte de testes unitários determinísticos em `internal/studio/server_test.go` cobrindo cenários conformes, tentativas maliciosas de escape de diretório e caminhos inexistentes (`TST-01`, `TST-03`, `TST-04`).

## Modelo do módulo
- **Modelo:** `enxuto` dentro do pacote [`internal/studio`](file:///Users/nathanramorim/git/forge-sdd/internal/studio).
- **Justificativa:** Estende o router existente em `server.go` sem criar novos daemons ou dependências externas (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Servidor HTTP Studio | `internal/studio/server.go` | `Server` com `ServeMux`, `handleState` e `handleEvents` | `estender` | Adicionar handler `handleSpec` mapeado em `/api/spec` respeitando o middleware read-only existente |
| Testes do Servidor | `internal/studio/studio_test.go` | Testes de inicialização e rotas do Studio | `estender` / `criar` | Criar `server_test.go` com testes focados em sanitização, path traversal e retorno de spec (`TST-03`) |

## Dependências novas (`SEG-07`)
Nenhuma (100% Go stdlib: `net/http`, `os`, `path/filepath`, `strings`, `encoding/json` — `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-codigo-e-testes.md`
- **IDs:** `ARQ-01`, `ARQ-04`, `SEG-01`, `SEG-02`, `SEG-04`, `COD-01`, `COD-02`, `TST-01`, `TST-03`, `TST-04`, `TST-05`
- **Skills Reutilizadas:** —

## Arquivos Tocados
- `internal/studio/server.go`
- `internal/studio/server_test.go`

## Critérios de Aceitação Executáveis
1. **CA-03 (Leitura de Spec Legítima e Prevenção de Path Traversal):**
   ```bash
   go test -v -race -run TestAPISpec ./internal/studio/... && go vet ./...
   ```
2. **Conformidade Geral de Testes e Tipagem:**
   ```bash
   go test -v ./internal/studio/... && go vet ./...
   ```

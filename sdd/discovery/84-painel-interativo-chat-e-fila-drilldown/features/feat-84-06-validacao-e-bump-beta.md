# feat-84-06 — Validação Global, Documentação e Bump da Versão Beta (v4.7.0-beta)

**Iniciativa:** 84-painel-interativo-chat-e-fila-drilldown
**Branch:** feat/84-painel-interativo-chat-e-fila-drilldown
**Status:** ready

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Atualização de versão, documentação do LEIA-ME e suíte de testes de regressão (`ARQ-11`).

## Decisões do projeto aplicadas
- `ARQ-01`, `ARQ-04`: Binário autocontido e versionamento semântico sem breaking changes.
- `COD-02`, `COD-04`: `go vet` global e ausência de código descontinuado ou comentado.
- `COD-05`: Idioma pt-BR na documentação do LEIA-ME e mensagens de ajuda.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Declaração de versão do CLI | `cmd/forge-sdd/main.go` | Constante `version = "4.6.1"` | estender | Bump para `4.7.0-beta` refletindo a nova funcionalidade de painel interativo sem breaking changes. |
| Versão do wrapper npm | `package.json` | `"version": "4.6.1"` | estender | Alinhar versão para `4.7.0-beta`. |
| Documentação do usuário | `internal/scaffold/templates/.forge/LEIA-ME.md.tmpl` | Instruções do tmux e painéis | estender | Documentar o stream narrativo dos workers e o modal de drill-down da fila (`Esc` para fechar). |
| Testes golden de scaffold | `internal/scaffold/testdata/golden/` | Snapshots dos templates gerados | atualizar | Sincronizar testdata com as instruções atualizadas da versão beta. |

## Dependências novas
- Nenhuma. Stdlib Go pura (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-04`, `ARQ-06`, `ARQ-11`
- `COD-01`, `COD-02`, `COD-04`, `COD-05`
- `DUP-01`, `DUP-02`
- `TST-01`, `TST-02`, `TST-05`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Escopo
1. Atualizar a versão do projeto para `4.7.0-beta` em:
   - `cmd/forge-sdd/main.go`
   - `package.json`
2. Atualizar o `LEIA-ME.md.tmpl` e arquivos golden correspondentes descrevendo:
   - O stream conversacional nos workers (`forge-builder` e `forge-reviewer`).
   - O popup modal interativo de fila com drill-down (`Enter` para inspecionar, `Esc` para voltar).
3. Executar validação completa de compilação, testes e linting em todos os pacotes.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -v -race ./...
```
- `forge version` imprime `4.7.0-beta`.
- Todos os testes de scaffold, golden, pipeline, monitor e CLI passam com 100% de sucesso.

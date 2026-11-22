# Feature 113-02 — Resolvedor do gestor ativo e leitura de `mcps.md`

## Objetivo
Decidir qual gestor está ativo na concepção: desligado (padrão, custo zero), GitHub via `gh` ou painel externo via MCP, lendo `forge.json`, `sdd/modulos/project.json` e `sdd/memory/mcps.md`.

## Modelo do módulo
**Enxuto** (ARQ-11): função pura em `internal/config` (dono de `PMConfig`, DUP-01), sem I/O externo além de ler arquivos do projeto.

## Decisões do projeto aplicadas
- **ADR-0017** (`adr-0017-mcp-somente-leitura-com-saida-saneada.md`, existe): o conteúdo de `mcps.md` é tratado como entrada não confiável (SEG-01/SEG-08); a ADR em si trata do MCP do próprio forge, então vale só o princípio.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| `PMConfig` | `internal/config/forge.go:92-99` | `Provider`, `Supported`, `RequireExistingCard`, `StatusOnImplement`, `StatusOnPRApproved`, `EvidenceMaxAgeMinutes`; NÃO existe `CreateOnConception` | Estender | Campo novo opcional `create_on_conception` (`omitempty`, retrocompatível). Divergência menor: o plano o trata como "opção"; ela ainda não existe |
| Leitor de `sdd/memory/mcps.md` | grep `mcps\.md` em `internal/` (só scaffold/templates/goldens) | Nenhum código Go lê `mcps.md`; hoje é só texto/tabela (`MCP \| Status \| Usado por`) | Criar | Parser mínimo da tabela; status que não seja exatamente `ativo` (ex.: "não verificado") conta como inativo |
| Módulo `project` | `sdd/modulos/project.json` | Sem `internal/modulo` na `main` (ver 113-05); só checar a existência do arquivo | Criar (checagem de arquivo) | Não depende da 109 |
| `VCS` | `internal/vcs/vcs.go:83` (`LeVCS`) | Lê a Constituição | Reusar | Condição de GitHub exige `VCS = github` |

## Dependências novas
Nenhuma (SEG-07, ARQ-04): stdlib Go.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-10`, `DUP-01`, `DUP-06`, `SEG-01`, `TST-03`, `TST-04`.

## Critérios de Aceitação Executáveis
Testes com `t.TempDir()`, sem rede e sem `gh` (TST-04).
```bash
set -o pipefail
go build ./... && go vet ./...
go test ./internal/config -run 'TestResolverGestorConcepcaoDesligadoPorPadrao|TestResolverGestorConcepcaoGitHubQuandoConfigurado|TestResolverGestorConcepcaoMCPAtivo|TestResolverGestorConcepcaoMCPIndisponivelDesliga|TestResolverGestorConcepcaoProviderDesconhecidoRecusado|TestResolverGestorConcepcaoMcpsMdAusenteOuMalformado|TestLoadForgeConfigSemCreateOnConceptionRetrocompativel'
go test ./internal/config
```
Casos: sem `forge.json`/`pm_mcp`/módulo, resultado `desligado`; `VCS=github` sozinho NÃO ativa (risco do discovery); `project.json` presente ou `create_on_conception: true` ativa GitHub; `clickup`/`jira`/`azure-devops` só ativa com `ativo` em `mcps.md`; `indisponível` ou "não verificado" desliga; provider fora de `Supported` recusado (negativo); `mcps.md` ausente ou malformado não gera erro fatal; `forge.json` antigo sem o campo segue carregando.

## Depende de
Nenhuma.

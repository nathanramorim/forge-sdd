# Constituição — forge-sdd

## Missão
CLI Go de comando único (`forge-sdd init`) que scaffolda estruturas Forge-SDD em qualquer projeto de software.

## Stack
| Camada | Escolha | Motivo |
|--------|---------|--------|
| Runtime | Go 1.22+ | Binário estático, cross-platform, sem runtime externo |
| CLI framework | cobra v1.8 | Padrão Go, flags + help automático |
| Prompts interativos | huh v0.3 | Sem CGO, TUI moderno |
| Templates | embed.FS (stdlib) | Zero dependência para templates |
| Render | text/template (stdlib) | Suficiente, zero deps |
| Testes | testify v1.9 | Assertions claras |
| Release | goreleaser v2 | Multi-OS automatizado |

## Decisões resolvidas
| Decisão | Resolução |
|---------|-----------|
| Linguagem | Go — binário estático, sem runtime |
| Único comando público | `init` — upgrade/doctor/archive são chatmodes Copilot |
| Embed de templates | `embed.FS` — binário único, sem assets externos |
| Não sobrescrever arquivos | Erro com lista de conflitos — evitar perda de dados |

## Ferramentas e Integrações
| Campo | Valor |
|-------|-------|
| VCS / Work Item System | github |

Consulte `sdd/memory/mcps.md` para o status real de cada MCP configurado (`ativo`/`indisponível`) antes de assumir que ele responde. Se "VCS / Work Item System" for `azure-devops`, use `az repos pr create` (ou instrução equivalente documentada) em vez de `gh pr create`. Se `nenhum`, deixe a branch pronta e informe o usuário, sem tentar nenhum comando de VCS.

## Regras (máx. 15)
1. Sem commits diretos em main
2. Branch por feature (`feat/*`)
3. Templates embutidos via `embed.FS` (nunca arquivo externo em runtime)
4. Secrets nunca no binário ou repositório
5. Antes de qualquer `go get`, consultar context7 com versão exata — desde que `sdd/memory/mcps.md` o liste como `ativo`; se `indisponível`, usar a documentação oficial da lib
6. Toda feature tem critério executável (`go build` ou `go test`)
7. `go vet ./...` deve passar após cada task
8. Binário final sem dependências de runtime além da stdlib
9. `--dry-run` nunca cria arquivos — apenas imprime árvore
10. CLI expõe os comandos públicos `init` e `update`
11. Versões Beta (`-beta`) não podem ser mescladas na `main`; a tag é criada diretamente na branch de feature e o PR correspondente deve ser mantido aberto no GitHub para testes de estabilização antes de promover para oficial. Opcionalmente, uma mesma versão ou branch beta pode agrupar e acumular múltiplos fixes ou features antes de forçar um novo bump de versão ou tag.
12. Documentação de Comandos e Release Notes: Ao introduzir novos comandos, eles devem ser atualizados nos documentos do repositório. Os Release Notes devem conter prévias dos comandos CLI (ex: `forge-sdd doctor`) e identificar explicitamente o agente/tipo no caso de prompts de chat de IA (ex: `/status para Claude`).
13. Agrupamento em Subpastas: Quando houver quebra de features grandes em subfeatures/tasks, os arquivos individuais de tasks devem ser agrupados dentro de uma subpasta com o nome da feature (ex: `sdd/features/feat-XX-nome/task-YY.md`). Se a quebra decorrer de um plano de discovery, a subpasta de feature gerada deve refletir o nome do discovery original (ex: `sdd/features/feat-XX-nome-discovery/`).
14. Telemetria Incondicional de Sessão: As métricas de sessão devem ser obrigatoriamente gravadas na fase `Close` do Orquestrador (ex: `sdd/.metrics/session-<ISO8601>.json`), mesmo que a sessão seja inativa, cancelada ou interrompida (timeouts). A propriedade `"feature"` deve conter o caminho relativo completo da feature/task aninhada em subpasta para precisão do rastreamento.
15. Branch Única por Pasta de Feature Quebrada: Quando uma feature/fix é quebrada em subtarefas dentro de uma subpasta (`sdd/features/<prefixo>-ID-<nome>/`), a pasta inteira é a unidade de execução — uma única branch (`<prefixo>/ID-<nome>`) agrupa todas as subtarefas, nunca uma branch por subtarefa. Antes de criar/usar a branch, pergunte obrigatoriamente (a) qual branch usar como ponto de partida (default `main`) e (b) verifique (`git branch --list <prefixo>/ID-*`) se já existe uma branch da mesma feature/fix de sessão anterior a retomar, em vez de recriar do zero.

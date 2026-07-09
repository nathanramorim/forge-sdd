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

## Regras (máx. 12)
1. Sem commits diretos em main
2. Branch por feature (`feat/*`)
3. Templates embutidos via `embed.FS` (nunca arquivo externo em runtime)
4. Secrets nunca no binário ou repositório
5. Antes de qualquer `go get`, consultar context7 com versão exata
6. Toda feature tem critério executável (`go build` ou `go test`)
7. `go vet ./...` deve passar após cada task
8. Binário final sem dependências de runtime além da stdlib
9. `--dry-run` nunca cria arquivos — apenas imprime árvore
10. CLI expõe os comandos públicos `init` e `update`
11. Versões Beta (`-beta`) não podem ser mescladas na `main`; a tag é criada diretamente na branch de feature e o PR correspondente deve ser mantido aberto no GitHub para testes de estabilização antes de promover para oficial.
12. Documentação de Comandos e Release Notes: Ao introduzir novos comandos, eles devem ser atualizados nos documentos do repositório. Os Release Notes devem conter prévias dos comandos CLI (ex: `forge-sdd doctor`) e identificar explicitamente o agente/tipo no caso de prompts de chat de IA (ex: `/status para Claude`).
13. Agrupamento em Subpastas: Quando houver quebra de features grandes em subfeatures/tasks, os arquivos individuais de tasks devem ser agrupados dentro de uma subpasta com o nome da feature (ex: `sdd/features/feat-XX-nome/task-YY.md`). Se a quebra decorrer de um plano de discovery, a subpasta de feature gerada deve refletir o nome do discovery original (ex: `sdd/features/feat-XX-nome-discovery/`).
14. Identificação Flexível e Workitems: Todo novo discovery ou feature gerada deve solicitar o Workitem de referência ao usuário. Caso omitido ou inexistente, deve ser gerado de forma automatizada um hash hexadecimal aleatório de 4 caracteres (ex: `3ec4`) para compor o prefixo (ex: `feat-3ec4-nome.md`), permitindo o desenvolvimento paralelo assíncrono e prevenindo colisões de numerações sequenciais rígidas.

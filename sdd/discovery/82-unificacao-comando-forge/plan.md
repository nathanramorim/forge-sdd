# Plano de Implementação — Discovery 82: Unificação do Comando `forge`

## Visão Geral e Estratégia
A iniciativa unifica a nomenclatura do comando para `forge` em todo o ecossistema Forge-SDD em 3 frentes coordenadas:
1. Camada de Apresentação e CLI (Go + Cobra + Monitor de Processos).
2. Templates de Scaffold e Golden Files de Teste.
3. Prompts dos Agentes, Regras de Governança e Documentação.

## Quebra de Tarefas (Onda Única)

| ID | Nome | Tipo | Depende de | O que constrói | Rules / IDs | Paralelo? |
|---|---|---|---|---|---|---|
| **82-01** | Padronização no CLI Go e Monitor | P | — | Atualizar `rootCmd.Use = "forge"` em `cmd/forge-sdd/main.go`, padronizar strings de dicas, avisos e mensagens de erro em `cmd/` e `internal/monitor/ps.go`. Atualizar testes unitários. | `ARQ-01`, `COD-01`, `COD-03`, `TST-01` | Sim |
| **82-02** | Atualização de Templates e Goldens | P | 82-01 | Atualizar referências a `forge-sdd` nos templates em `internal/scaffold/templates/` (`LEIA-ME.md.tmpl`, `forge-up.sh`, etc.) e sincronizar os testes golden (`internal/scaffold/testdata/golden/`). | `DUP-10`, `TST-02` | Não |
| **82-03** | Prompts, Regras e Documentação | P | 82-01 | Atualizar `.agents/rules/`, `.agents/skills/`, `GEMINI.md`, `CLAUDE.md`, `docs/comandos.md` e `README.md` para instruir unicamente `forge`. Executar validação com `go vet` e `go test ./...`. | `COD-05`, `DUP-01`, `FLX-05`, `GOV-01`, `TST-05` | Sim (com 82-02) |

## Critério Final da Iniciativa
Todos os critérios de aceitação (CA-01 a CA-05 de `criteria.md`) satisfeitos e suíte de testes 100% verde sem regressões.

## Pós-Portão
Submissão para aprovação humana (Gate 1 — `FLX-01`). Após aprovação, executar `/split-features` para gerar as specs de cada task e entregá-las para a esteira do Builder.

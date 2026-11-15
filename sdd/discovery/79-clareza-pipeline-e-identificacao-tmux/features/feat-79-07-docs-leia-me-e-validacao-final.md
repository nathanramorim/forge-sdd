# feat-79-07 — Documentação, LEIA-ME e Validação Final

**Branch:** `feat/79-clareza-pipeline-e-identificacao-tmux`
**Fase:** 79-07
**Tamanho:** `P`
**Depende de:** 79-01, 79-03, 79-04, 79-05, 79-06
**Status:** `completed` (FLX-01 aprovado)

## Contexto & Motivação
Com a unificação do menu, identificação clara de projetos no tmux, mini-monitor de esteira, paralelismo de filas por DAG e monitor nativo de processos (`forge-sdd ps`), a documentação do usuário (`LEIA-ME.md`, guias do pipeline e conceitos) precisa ser atualizada. A validação de toda a suíte de testes e o Gate de governança garantem que não houve regressões.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Documentação técnica, atualização de templates markdown e execução de suíte de testes (`ARQ-11`).

## Decisões do projeto aplicadas
- `DUP-10`: Toda mudança de comportamento é refletida no guia e documentação do projeto.
- `GOV-01`: Governança Camada A e B sem violações.
- `FLX-05`: Verificação de critérios executáveis.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Guia do usuário do pipeline | `internal/scaffold/templates/.forge/LEIA-ME.md.tmpl` | Instruções de uso do tmux e menu | estender | Documentar a barra de status, monitor de processos `forge-sdd ps` e novo layout de menu. |
| Conceitos da arquitetura | `sdd/spec/concepts.md` | Descrição das 5 estações e pipeline | estender | Reforçar o princípio Pipeline First e o paralelismo por ondas na fila. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `DUP-10`, `FLX-05`, `GOV-01`, `TST-05`.
- Rules: `sdd-reuso-e-duplicacao.md`, `sdd-fluxo-e-portao.md`.

## Escopo
1. Atualizar `.forge/LEIA-ME.md.tmpl` com o comando `forge-sdd ps`, paralelismo da fila e novo menu.
2. Atualizar a documentação de conceitos em `sdd/spec/concepts.md`.
3. Executar `go build ./...`, `go vet ./...` e `go test -race ./...`.
4. Executar `forge-sdd gate` garantindo 0 violações na Camada A.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test ./...
forge-sdd gate
```
- Suíte completa verde com 0 violações.

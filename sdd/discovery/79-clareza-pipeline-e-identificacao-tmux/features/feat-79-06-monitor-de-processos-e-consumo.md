# feat-79-06 — Monitor de Processos, Workers e Consumo do Projeto (`forge-sdd ps`)

**Branch:** `feat/79-clareza-pipeline-e-identificacao-tmux`
**Fase:** 79-06
**Tamanho:** `M`
**Depende de:** 79-02
**Status:** `completed` (FLX-01 aprovado)

## Contexto & Motivação
O desenvolvedor precisa de visibilidade transparente sobre o que está rodando no computador relacionado ao projeto: quais workers, subagentes e CLIs estão ativos, quanto de CPU e memória RAM estão consumindo, há quanto tempo estão operando e se há processos zumbis ou travados gerando processamento oculto. Esta feature implementa um inspetor nativo de processos do projeto e o comando `forge-sdd ps`.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Inspeção de processos de sistema operacional (`ps` / `/proc` / `tmux list-panes`), parseamento de uso de CPU/RAM, manipulação segura de PIDs e interface de exibição no CLI (`ARQ-11`, `COD-01`, `SEG-01`).

## Decisões do projeto aplicadas
- `ADR-0006`: Visibilidade e inspeção no tmux.
- `ARQ-01`, `ARQ-04`: Sem dependências CGO ou bibliotecas pesadas de runtime além da stdlib Go e utilitários padrão de SO.
- `COD-01`, `COD-05`: Código modular e mensagens claras em pt-BR.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Mapeamento de painéis tmux | `internal/pipeline/tmux.go` | `TmuxSessionName`, `TmuxRunner` | estender | Usar `tmux list-panes` para extrair os PIDs base de cada estação do projeto. |
| Estrutura de subcomandos | `cmd/forge-sdd/` | Comandos cobra (`status.go`, `doctor.go`) | estender | Criar `cmd/forge-sdd/ps.go` para expor `forge-sdd ps`. |
| Layout de tabela no terminal | `internal/gate`, `status.go` | Formatação com alinhamento | reusar | Manter o mesmo padrão estético e limpo do Forge. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`). Usa comandos padrão de sistema (`ps -p`, `pgrep` ou stdlib Go `os/exec`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-04`, `COD-01`, `COD-03`, `SEG-01`, `TST-04`.
- Rules: `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`, `sdd-seguranca.md`.

## Escopo
1. Criar `internal/monitor/ps.go`:
   - Mapear os processos ligados à sessão tmux do projeto (`forge-<projeto>-<hash>`) e subprocessos filhos (CLIs dos agentes, despachantes, processos de teste).
   - Extrair métricas por processo: PID, Papel/Comando, % CPU, Memória RSS (MB), Tempo Ativo (Elapsed) e Estado (Ativo / Ocioso / Esperando).
   - Calcular totais agregados de CPU e Memória consumidos pelo ecossistema do projeto.
2. Criar subcomando `forge-sdd ps` (`cmd/forge-sdd/ps.go`) com:
   - Exibição de tabela limpa no terminal.
   - Flag `--json` para integração com dashboards e ferramentas.
   - Opção `--kill-zombies` para encerrar processos orfãos/presos com segurança mediante confirmação.
3. Expor o monitor de processos no menu interativo (`forge`) e fornecer resumo de consumo (CPU/RAM) para o mini-painel tmux.

## Critérios de Aceitação Executáveis
```bash
go test -v ./internal/monitor -run TestProjectProcessInspection
go build ./... && ./forge-sdd ps --dir .
```
- A listagem identifica os processos associados à sessão e diretório do projeto.
- Informações de PID, memória e CPU são exibidas corretamente sem travar se um processo terminar durante a leitura.
- `go build ./... && go vet ./...` sem erros.

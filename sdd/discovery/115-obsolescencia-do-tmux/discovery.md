# Discovery — 115 Obsolescência do tmux (o chat orquestra; as sessões de terminal saem)

## Problema

O forge-sdd nasceu com o pipeline em **três sessões de terminal** abertas em painéis do tmux: Discovery, Builder e Reviewer, cada uma um processo com sua CLI, conversando pelo barramento em `.forge/runtime/`. Esse desenho tem **1.585 linhas de código** só de tmux (abertura de painéis, instalador por gerenciador de pacotes, despachante que injeta prompts e lê a tela).

A **ADR-0015** já registrou o caminho novo: o pipeline acontece **no mesmo chat**, com cada papel como líder que coordena subagentes, e o Go como dono do laço. Com isso, o tmux deixa de ser necessário — ele era o transporte das sessões, e as sessões viraram subagentes.

**A decisão do dono (21/11/2026):** *"a verdade é que não usa mais; agora quem faz toda a orquestração é o próprio chat e escala os subagentes."*

## Para quem

- **O dono**, que não quer instalar nem manter tmux, e quer o fluxo no chat.
- **Quem instala o forge hoje em Windows**, onde o tmux não existe nativo e o modo `auto` já cai para janelas do sistema.
- **Quem mantém o projeto**, que carrega três implementações paralelas de terminal (tmux, janelas nativas, Windows Terminal) para o mesmo objetivo.

## Estado atual (levantado no código)

| Peça | O que faz | Tamanho |
|---|---|---|
| `internal/pipeline/tmux.go` | monta o plano de janelas/painéis e abre o tmux | 333 linhas |
| `internal/pipeline/tmuxinstall.go` | detecta gerenciador (brew, apt-get, dnf, pacman, apk, zypper) e instala o tmux | 292 linhas |
| `internal/pipeline/deliver.go` | despachante: `tmux load-buffer`, `paste-buffer`, `capture-pane` para injetar prompt e ler a tela | usa `exec.Command("tmux", ...)` |
| `internal/pipeline/status_live.go` | painel ao vivo, com o tmux como uma das fontes | consome tmux |
| `internal/pipeline/preflight.go` | avisa quando o tmux falta e sugere `--terminal native` | ~10 linhas |
| `cmd/forge-sdd/menu.go` | `terminalLabel`: "no tmux (3 painéis)" / "em janelas do sistema" | ~15 linhas |
| **Testes** | `tmux_test.go` (437), `tmuxinstall_test.go` (295), `tmux_real_test.go` (228) | 960 linhas |

Configuração existente: `terminal: auto | tmux | native` em `forge.json`; `auto` já cai para `native` quando o tmux falta.

## Jornada desejada

1. `forge init` e `forge` não perguntam sobre terminal nem oferecem instalar tmux.
2. O pipeline roda no chat: o `/implementar` e o `/revisar` escalam subagentes, sem abrir janela.
3. Quem tiver `terminal: tmux` gravado recebe um aviso claro de que o modo saiu, com o que fazer — não um erro seco.
4. O `forge doctor` não menciona tmux; menciona o que de fato importa para o modo chat.

## Escopo

**Dentro:**
- Descontinuar o modo `sessions`/tmux como fluxo de produto (ADR substituindo a 0008 e fechando a 0015 como decidida).
- Remover `tmux.go`, `tmuxinstall.go` e os testes correspondentes; adaptar `deliver.go` e `status_live.go` ao que sobra.
- Migrar a configuração: `terminal: tmux` deixa de ser aceito, com mensagem de transição.
- Atualizar README, `docs/`, `docs/comandos.md` e `/tutorial`, que hoje ensinam a instalar tmux.
- Uma nota de release explicando a saída e o que fazer.

**Fora:**
- Remover o barramento (`.forge/runtime/`), que continua sendo o estado dos jobs no modo chat.
- Mexer no modo `worker` headless, se ele não depender de tmux (a confirmar no critério).
- Reimplementar o despachante de prompts por outro transporte: no modo chat ele não existe (o chat entrega o prompt ao subagente).

## Riscos

- **R1 — Perda de quem usa o modo sessões.** Mitigação: uma versão de transição em que `terminal: tmux` avisa e não é removido ainda; só remove na versão seguinte.
- **R2 — `deliver.go` e `status_live.go` acoplados ao tmux de forma não óbvia.** Mitigação: o critério lista tudo que importa `tmux`, e o teste prova que o binário não executa `tmux` em nenhum caminho.
- **R3 — Documentação ensinando tmux espalhada.** Mitigação: um grep no critério (`grep -ri tmux docs/ README.md`) com lista fechada.

## Regras de negócio

- **RN-1** O produto tem **um** caminho de orquestração: o chat com subagentes. Modos alternativos de terminal saem; não são mantidos "por garantia".
- **RN-2** Configuração antiga não quebra em silêncio: `terminal: tmux` recebe aviso explícito com o que fazer.
- **RN-3** A saída é registrada em ADR, porque muda a arquitetura (ADR-0008 sai, ADR-0015 entra como decidida).
- **RN-4** Nenhuma versão estável publicada com o tmux como obrigatório: a remoção entra numa beta primeiro.

## Métricas de sucesso

- `grep -ri 'tmux' internal/ cmd/` retorna **zero** ocorrências fora de mensagem de transição.
- `go test ./...` verde sem `tmux_real_test.go` (que exige tmux instalado).
- Um projeto novo faz `forge init` → `forge` → pipeline no chat, **sem instalar nada de terminal**.

## A confirmar

- [ ] `deliver.go` e o modo `worker` headless dependem de tmux, ou só o modo `live`? (levantar no `criteria.md`)
- [ ] A transição fica **uma** versão (aviso) ou corta direto?
- [ ] O `forge doctor` ganha o quê no lugar do aviso de tmux?
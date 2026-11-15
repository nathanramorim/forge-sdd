# Discovery 68 — Sessões ACP Físicas Multi-Agente

## 1. Por quê (problema)
O chip "5 SESSÕES PRONTAS" do Forge Studio (`internal/studio/assets/index.html`, contador `sessionPoolCount`) é enganoso: `acp.SessionPool.Warmup` (`internal/acp/pool.go`) apenas registra as 5 sessões em memória com estado `standby`. Nenhum processo de agente existe. O operador acredita que há 5 agentes conectados quando não há nenhum.

## 2. Para quem
- Operador do Studio (precisa confiar no indicador de prontidão).
- Autor de agentes/estações (precisa plugar Copilot, Claude Code, agy, Gemini, Codex ou outro sem alterar o código do forge-sdd).

## 3. Decisões do humano (2026-11-01)
1. Um processo filho real por estação (5), comando configurável, com PID, stdin/stdout e health check.
2. Adaptador por agente (copilot, claude code, agy, gemini, codex, outros). Nada hardcoded no Go.
3. Falha: chip "x de 5 conectadas", estação em `error`, Studio segue em modo leitura sem bloquear o boot.
4. Zero dependências runtime (`ARQ-03`/`ARQ-04`), sem emojis, UI em pt-BR.

## 3.1 Decisões do Portão FLX-01 (2026-11-01)
1. Fallback "processo vivo + stdin aberto" aprovado para agentes sem ACP via stdio (`healthCheck.method` configurável).
2. Presets pesquisados na documentação pública; o que não foi confirmado fica "a confirmar" (ver `criteria.md`).
3. `forge-sdd config`: `list/get/set/unset` apenas para chaves `acp.*`; demais chaves do `.sddrc` ficam como evolução futura.
4. Restart automático com backoff exponencial (tudo configurável em `acp.*`); estação "reiniciando" e depois `error` ao esgotar. Nova task 68-05: painel de falhas no Studio.
5. Timeouts configuráveis, default herdado do warmup atual.

## 3.2 Decisão adicional do FLX-01
6. Preflight de requisitos: `init`/`update` perguntam quais agentes o usuário usará (multisseleção, sem instalar todos), detectam CLI/adaptador/Node/npx, mostram o comando exato de instalação e só executam com confirmação (`--yes` em CI; não interativo apenas reporta). Mesma checagem no `doctor` e no boot do Studio. Nova task 68-06.

## 4. Solução macro
- **Configuração:** `sdd/.sddrc` ganha o bloco `acp` com mapa de adaptadores (`command`, `args`, `env`, `healthCheck`) e a escolha do adaptador por estação. Gerenciado por um novo subcomando `forge-sdd config`.
- **Processo físico:** novo componente em `internal/acp` que inicia o filho com `os/exec` (argv explícito, sem shell), captura PID, liga stdin/stdout ao `AttachBridge` e executa health check periódico.
- **Pool honesto:** `Warmup` passa a iniciar os 5 processos; estado `standby` só após o filho responder ao health check. Falha de uma estação não aborta as demais.
- **Preflight:** pacote único de checagem de requisitos reutilizado por init, update, doctor e Studio; grava `acp.enabledAgents`.
- **Resiliência:** queda do filho dispara restart com backoff exponencial e jitter; histórico de falhas por estação exposto em endpoint de leitura e painel no Studio.
- **Studio:** chip mostra "x de 5 conectadas"; estação com falha aparece em erro; boot nunca bloqueia.

## 5. Fora de escopo
Execução de tarefas reais pelos agentes além do handshake/health check, autenticação dos agentes, PR (proibido até o humano retornar).

## 6. Métricas de sucesso
- Chip só exibe "5 de 5 conectadas" com 5 PIDs vivos verificados.
- Matar um filho muda a estação para `error` em até 1 ciclo de heartbeat.
- Studio sobe com agente ausente, em modo leitura.
- Filho morto é reiniciado dentro do backoff; ao esgotar tentativas a estação vai a `error` e o painel mostra causa, exit code e tentativas.

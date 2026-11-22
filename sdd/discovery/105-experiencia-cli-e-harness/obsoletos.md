# Obsoletos — proposta (nada removido nesta discovery)

Modelo atual: 3 sessões (Discovery, Builder, Reviewer) mais aprovação humana. Fonte da lista anterior: `obsoletos.md` da 104 em `bbd26d6` (não está na `main`). Aqui, cada item tem evidência, quem ainda usa, ação e risco. Ações: **remover**, **avisar** (depreciar com mensagem), **manter**, **corrigir texto**. Toda remoção depende do dono (`A confirmar`).

| # | Item | Evidência | Quem ainda usa | Ação proposta | Risco |
|---|---|---|---|---|---|
| O1 | Texto "5 estações" em `CLAUDE.md.tmpl`, `GEMINI.md.tmpl`, `.claude/agents/lider.md(.tmpl)`, golden `copilot-instructions.md`, `docs/acp-ide-setup.md:3`, `.agents/rules/sdd-acp.md:2`, `attach.go:21` (busca por "5 estações") | geram texto que o usuário recebe no `init` | Todo projeto novo | **Corrigir texto** para "3 sessões"; atualizar goldens no mesmo PR (`TST-02`) | Médio: muda arquivo entregue; revisar `update` |
| O2 | Agente `lider` | `.claude/agents/lider.md` ("coordena as 5 estações"); `internal/pipeline/launcher.go:125`, `tmux.go:133`, `preflight.go:165` o usam como agente da sessão Discovery; testes `launcher_test`, `tmux_test`, `preflight_test`, `up_sessions_test` citam | `forge up` | **Avisar e redirecionar**: manter o arquivo, trocar descrição; renomear para `discovery` só com teste e ADR; **não remover** | Alto se removido: quebra `forge up` |
| O3 | `forge queue stations` e `internal/queue/stations.go` | `forge queue stations` imprime 5 filas vazias (rodado) | `attach.go` e `internal/queue/activity.go` citam as estações (busca no código) | **Avisar** (cobra `Deprecated`); remover só após ADR que altera `ARQ-06` | Médio: lista de `ARQ-06` |
| O4 | `forge attach` (sessões ACP por estação) | `attach.go:21` fala em 5 estações; `forge --help` lista | ACP em IDEs (`docs/acp-ide-setup.md`) | **Avisar**; decidir com o dono se ACP segue (a 104 reaproveita `internal/acp` para MCP) | Médio |
| O5 | `queue handoff` | `forge queue --help`: "entre duas estações" | skills `sdd-*` | **Avisar** | Baixo |
| O6 | Skills `sdd-concepcao`, `-planejamento`, `-build`, `-governanca`, `-operacao` | `.agents/skills/` (cada `SKILL.md` tem de 1,2 a 2 KB); ADR-0003; comandos `.claude/commands/*` apontam para elas | As 3 sessões, via comandos | **Manter** (apoio das 3 sessões); renomear só em ADR própria | Baixo se mantidas |
| O7 | `/implement` e `/proxima-feature` (aliases de `/implementar`) | `.claude/commands/implement.md` (131 B) e `implementar.md` (133 B) apontam para o mesmo arquivo | Quem digita o alias | **Avisar** no texto do alias; remover depois | Baixo |
| O8 | Cabeçalhos de versão antigos: `docs/pipeline.md:1` (v4.5.0-beta), `sdd/spec/flows.md:1` (v4.7.7-beta), `npm/package.json` 4.9.1 (segundo a 104) | leitura dos cabeçalhos | Quem lê a doc | **Corrigir texto**; `npm` entra no checklist de release | Baixo |
| O9 | Texto "pronta para uso com GitHub Copilot" e `init --agent` padrão `copilot` (`main.go:22-23`, `init --help`) | rodado | Usuário novo | **Corrigir texto**; padrão do agente é decisão do dono | Baixo (texto); médio (padrão) |
| O10 | `forge doctor` "em sincronia" contra `.sdd-version` e lista de 126 caminhos | `doctor.go:90-96` | Todos | **Corrigir** (comparar com o binário; lista só com `--verbose`) | Baixo |
| O11 | Comando `forge review` | não existe; é `forge bus review` (`review.go`, `busReviewCmd`) | Doc que cita `forge review` | **Corrigir texto** na doc; nada a remover | Baixo |
| O12 | ADRs `proposto` em uso (0009 a 0014) | cabeçalho de cada arquivo | Código | Pedir **aprovação** ao dono; não é remoção | Baixo |
| O13 | 5 pareceres antigos da 97 que não decodificam | `doctor --harness` em `bbd26d6`: 5 de 17 | Medição de proficiência | **Avisar** (arquivar a iniciativa 97); não regravar | Baixo |
| O14 | Iniciativas 79 a 88 com fila concluída aparecendo como ativas | `forge status` | Todos | **Corrigir** o filtro; `/archive` é decisão do dono | Baixo |

**Mantém (não obsoleto):** `archivist`, `migrator`, `/toca-ai`, `/inbox`, `/status`, `/doctor`, `/tutorial`, `/archive`, `/nova-feature`, `/novo-fix`, `/c4-architecture`, `/constitution`, `/upgrade-sdd`, skill `novo-subcomando-cli-sdd`, modo `sessions` (tmux).

**Não verificado:** quem usa `attach`/ACP fora deste repositório; se `forge-sdd` (alias do npm) ainda é usado; comportamento de `update` ao regravar os templates corrigidos.

**Gate 1 (2026-11-20):** aprovado só corrigir texto (O1, O8, O9, O10, O11, O14) e avisar (O3, O4, O5, O7, O13); `lider` fica (só descrição), skills `sdd-*` e a lista "Mantém" ficam; nada removido. O9: `init` passa a perguntar o agente, sem padrão fixo.

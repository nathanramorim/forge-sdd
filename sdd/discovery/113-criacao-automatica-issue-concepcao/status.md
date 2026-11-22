# Status — 113 Criação automática de issue/card na concepção

**Gate 1:** aprovado
- **Iniciativa:** `113-criacao-automatica-issue-concepcao`
- **Branch:** `feat/113-criacao-automatica-issue-concepcao`
- **Status:** `spec-para-revisao` (Discovery, Criteria e Plan concluídos; aguardando Gate 1)
- **Próximo passo:** Aprovação humana no Gate 1 (`[Seguir]` / `[Ajustar]`) e posterior `/split-features`.
- **A confirmar (Gate 1):**
  1. Interruptor de ativação no GitHub: ativar quando o módulo `project` (109) estiver configurado ou `pm_mcp` habilitar `github-projects` / `create_on_conception: true`, mantendo desligado (custo zero) por padrão.
  2. Granularidade: 1 issue-mãe no `/discovery`, 1 issue por `/nova-feature` e 1 issue por `/novo-fix`.
  3. Escala para outros painéis (`clickup`, `jira`, `azure-devops`): criação pelo agente via servidor MCP ativo em `sdd/memory/mcps.md` + registro determinístico em `card-evidence.json`.

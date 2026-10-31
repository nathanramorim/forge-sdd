# Catálogo de Conceitos e Contratos (`DUP-01`) — forge-sdd

Regra `DUP-01`: **Um conceito = uma casa.** Antes de criar uma nova struct, pacote ou utilitário, consulte esta tabela para verificar qual pacote já é dono do conceito e reutilize ou estenda seu contrato público (`DUP-02`).

| Conceito | Pacote/Domínio Dono | Entidade/Struct Mestra | Contrato Público | Observação |
| --- | --- | --- | --- | --- |
| Configuração do Projeto (`.sddrc`) | `internal/config` | `config.Config` | `LoadSddrc`, `SaveSddrc`, constantes `Agent*` | Fonte única para stack, agentes configurados e versão SDD |
| Motor de Scaffold e Migração | `internal/scaffold` | `templatesFS` (`embed.FS`) | `Walk`, `Run`, `RunAgents`, `UpdateSddrc` | Materializa templates e preserva domínio (`shouldPreserve`) |
| Diagnóstico de Integridade | `internal/doctor` | `doctor.Report` | `forge-sdd doctor` | Verifica saúde da estrutura SDD, naming convention e resumo de métricas |
| Telemetria de Sessão | `internal/telemetry` | `telemetry.SessionRecord` | `Record`, `Summary`, `Report` (`forge-sdd session record`, `forge-sdd report`) | Gravação determinística em `sdd/.metrics/session-<ISO8601>.json` |
| Lições Aprendidas | `internal/lessons` | `lessons` | `Add` (`forge-sdd lessons add`) | Mantém `sdd/memory/lessons.md` dentro do orçamento |
| Fila Determinística do Builder | `internal/queue` | `queue.Queue`, `queue.TaskItem` | `Enqueue`, `Next`, `Complete`, `Review`, `List` (`forge-sdd queue`) | Persiste estado das tasks em `sdd/.queue/queue.json` e valida parecer JSON Camada B |
| Portão de Governança (Camada A) | `internal/gate` | `gate.CheckResult`, `gate.Report` | `RunChecks` (`forge-sdd gate`) | Executa verificações determinísticas `P-xx` em modo `observacao` ou `bloqueio` |
| Motor de Auto-Skills | `internal/skills` | `skills.Skill` | `Propose`, `Approve`, `Sync`, `List` (`forge-sdd skills`) | Gerencia `.agents/skills/<slug>/SKILL.md` (`DUP-05`, `GOV-03`) e sincroniza adaptadores |
| Aplicação Estática Web & SEO (`/site`) | `site/` | Astro 5 SSG (`dist/`) | `npm --prefix site run build / test`, Cloudflare Workers (`wrangler.jsonc`) | Hospedada em `https://forge-sdd.org` com SSG pré-renderizado, SEO técnico, Schema.org e motor Three.js cliente isolado (`ARQ-04`) |

# Critérios — Iniciativa 115 (pipeline e seleção de modelos)

## Restrições (Constitution)

ARQ-01, ARQ-02, ARQ-04, ARQ-06, ARQ-11, COD-03, DUP-01, DUP-02, FLX-01, FLX-02, FLX-04, FLX-05, GOV-01, SEG-01, SEG-03, TST-01, TST-03, TST-04 e TST-05.

## Modelo do módulo

**Enxuto** (ARQ-11): sem pacote novo. Estender internal/config para tiers/mapeamentos, internal/bus.Job para complexidade/tier, internal/pipeline para despacho e resolução pelo worker, e as skills comuns para orientar o orquestrador. cmd/ só encaminha flags e comandos.

## Decisões do projeto aplicadas

- ADR-0004: worker Go reivindica jobs via barramento e controla lease; Reviewer é papel separado.
- ADR-0005: cada papel pode usar CLI diferente; CLIFor é resolvedor único e os perfis usam argv explícito.
- ADR-0014 D-5: já há modelo e esforço fixos por papel; esta entrega acrescenta seleção por task.
- ADR-0002 e ADR-0016: estado e identidade pertencem à iniciativa; workers respeitam esse escopo.
- ADR-0023: estado das tasks e PRs paralelos é isolado.
- ADR-0024 proposta: registrar orquestração, escopo por iniciativa, tiers e tmux opcional antes do código (FLX-05, GOV-01).
- ADR-0015 está proposta e não é tratada como decisão aprovada.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Modelo estático por papel | feat-90-07; internal/config/forge_profiles.go | models/efforts; worker usa SessionModelArgs | Estender e manter como fallback | Não duplicar feat-90-07 |
| Modelo por job | internal/bus/job.go | Job não tem complexidade/tier | Estender com metadados por task | Builder recebe escolha sem reclassificar |
| Despacho | internal/pipeline/dispatch.go | Valida spec/plano/evidência antes de enfileirar Builder | Preservar e validar tier | Spec antes do código |
| Handoff para Reviewer | internal/pipeline/worker.go, internal/pipeline/pr.go | PR aciona Camada A e VerifyJobHandoff enfileira Reviewer | Preservar e testar | Código chega validado |
| Estimativa histórica | iniciativa 110, E-09 | Recomenda modelo de forma descritiva se houver amostra | Reusar como evidência, sem substituir decisão | Evita duplicar estimativa |
| Skills PM-Architect/Planner | .agents/skills/sdd-concepcao/, .agents/skills/sdd-planejamento/ | Não há rubrica nem Advisor condicional | Estender fonte comum | Mesma política em todas as CLIs |

## Dependências novas

Nenhuma (SEG-07). Não consultar catálogo externo durante despacho.

## Regras de complexidade e tiers

- **Baixa:** escopo isolado, arquitetura e critérios explícitos, sem dependência externa/migração. Tier fast.
- **Média:** mudanças relacionadas, mas arquitetura conhecida e critérios claros. Tier balanced.
- **Alta/incerta:** fronteira arquitetural, segurança/migração, dependência externa ou múltiplas interpretações. PM-Architect consulta Advisor no tier reasoning.
- Complexidade da iniciativa não se propaga às tasks: split claro continua fast. Advisor não escolhe automaticamente um modelo caro para Builder.

## Mapeamento por CLI

Tiers fast, balanced e reasoning são neutros de provedor e configuráveis. Claude recomenda haiku, sonnet e opus nessa ordem. Outras CLIs recebem nomes próprios configurados; só enviar flags/parâmetros aceitos pela CLI ativa. Sem mapeamento, usar modelo fixo por papel com aviso ou falhar quando o Advisor for obrigatório; nunca usar nome de outro provedor.

Referências oficiais consultadas: [Claude Code model configuration](https://code.claude.com/docs/en/model-config), [Codex CLI](https://developers.openai.com/codex/cli), [Gemini CLI model selection](https://geminicli.com/docs/cli/model/), [GitHub Copilot CLI](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference).

## C4 — Contexto

    C4Context
      title Pipeline SDD com seleção de modelo
      Person(owner, "Dono", "Define a demanda e aprova gates")
      System(forge, "Forge SDD", "Orquestra agentes, filas e verificações")
      System_Ext(clis, "CLIs de agentes", "Claude, Codex, Gemini, Copilot, agy")
      Rel(owner, forge, "Descreve demanda e aprova")
      Rel(forge, clis, "Aciona agentes com tier resolvido")

## C4 — Contêineres

    C4Container
      title Seleção e despacho
      System_Boundary(forge, "Forge SDD") {
        Container(core, "Pipeline Go", "internal/pipeline", "Valida gate, resolve tier e executa jobs")
        Container(config, "Configuração", "internal/config", "Mapeia tier lógico para modelo e CLI")
        Container(queue, "Barramento", "internal/bus", "Transporta jobs com metadados")
        Container(skills, "Skills", ".agents/skills", "Instrui liderança, concepção, planejamento e build")
      }
      System_Ext(clis, "CLIs", "Processos de agentes com modelo configurado")
      Rel(skills, core, "Define complexidade e tier")
      Rel(core, config, "Resolve modelo pela CLI")
      Rel(core, queue, "Enfileira job com tier")
      Rel(core, clis, "Executa Builder/Reviewer/Advisor")

## Fluxo — Sequência

    sequenceDiagram
      actor Dono
      participant Lider
      participant PM as PM-Architect
      participant Advisor
      participant Go as Forge Go
      participant Builder
      participant Reviewer
      Dono->>Lider: demanda
      Lider->>PM: conceber e classificar
      PM-->>Lider: spec e tiers por task
      opt alta incerteza
        Lider->>Advisor: consulta no tier reasoning
        Advisor-->>Lider: parecer técnico
      end
      Dono->>Lider: Gate 1 aprovado
      Lider->>Go: dispatch da spec validada
      Go->>Builder: job com model_tier
      Builder->>Go: código e validações; PR após Camada A
      Go->>Reviewer: PR validado
      Reviewer-->>Lider: parecer

## Critérios de Aceitação Executáveis

1. Testes classificam baixa/média/alta com justificativa e provam que uma task simples e clara recebe fast.
2. O JSON do job leva complexity/model_tier; jobs antigos sem tier mantêm models.builder ou inherit.
3. O worker resolve o tier usando a própria CLI. Testes de argv cobrem Claude, Codex, Gemini, Copilot e agy sem supor nomes iguais.
4. Claude mapeia fast/balanced/reasoning para haiku/sonnet/opus; Builder não recebe Sonnet/Opus para task fast sem justificativa.
5. Advisor só é acionado para alta incerteza/decisão arquitetural, não escreve código e registra modelo/tier na telemetria.
6. CLI sem modelo configurado não recebe flag inventada nem modelo de outro provedor; despacho informa fallback ou falta de config.
7. Reviewer só recebe job depois de forge pipeline pr concluir Camada A e registrar PR; falha não enfileira Reviewer.
8. Workers só reivindicam a iniciativa alvo, mantêm review/rework da mesma iniciativa e encerram quando as filas Builder e Reviewer dela drenam.
9. Jobs ativos terminam normalmente; mudança não interrompe processos ou janelas existentes.
10. go vet ./... e go test ./... passam.

## Riscos

| Risco | Mitigação |
|---|---|
| Nomes/flags mudam entre CLIs e contas | Tiers configuráveis, validação por perfil, sem catálogo embutido |
| Advisor acionado em tasks rotineiras | Rubrica exige alta incerteza; tasks claras ficam fast |
| Campo novo quebra job antigo | Campo opcional e fallback por papel |
| Fila global mistura iniciativas | Filtro por iniciativa, incluindo review/rework |
| Reviewer recebe código incompleto | Preservar forge pipeline pr e VerifyJobHandoff |

## A confirmar no Gate 1

A interpretação de “enviar com código pronto e validado” como handoff ao Reviewer/PR está registrada em discovery.md.

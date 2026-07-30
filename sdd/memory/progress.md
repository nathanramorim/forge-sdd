# Progress — forge-sdd

## Status
```
Fase 0 — Foundation           [x] done
Fase 1 — Templates embed      [x] done
Fase 2 — Init interativo      [x] done
Fase 3 — Init flags           [x] done
Fase 4 — Dry-run              [x] done
Fase 5 — Versioning           [x] done
Fase 6 — Self-test            [x] done
Fase 7 — Release              [x] done
Fase 8 — npx                  [x] done
Fase 9 — Multi-agent          [x] done
Fase 10 — Landing page         [x] done
Fase 11 — Unify agent structures [x] done
Fase 14 — Discovery command        [x] done
Fase 16 — Constitution command     [x] done
Fase 17 — Unified onboarding docs  [x] done
Fase 18 — SDD Handoff Protocol     [x] done
Fase 23 — Agent-specific MCP config   [x] done
Fase 24 — Automated handoff flow    [x] done
Fase 30 — Fix branch & PR flow      [x] done
Fase 31 — Upgrade domain preservation [x] done
Fase 32 — Product release notes       [x] done
Fase 33 — Automated wiki sync         [x] done
Fase 34 — Update agent prompts         [x] done
Fase 35 — Remove install skill         [x] done
Fase 36 — Structured documentation      [x] done
Fase 37 — Publish stable 1.6.0          [x] done
Fase 38 — CLI smart upgrade/beta        [x] done
Fase 39 — CLI doctor and destroy cmd       [x] done
Fase 40 — CLI init directory refinement     [x] done
Fase 41 — Nested features/discovery grouping [x] done
Fase 42 — Consolidação e Lançamento v1.7.0 [x] done
Fase 1.9.0-beta — Atalho forge e Boas-vindas    [x] done
Fase 5ae2-08 — Nomenclatura, Telemetria e Tokens [x] done
Fase 47 — Fix flag naming-convention no update  [x] done
Fase 48 — Fix /novo-fix ausente no Copilot      [x] done
```

## Features ativas
| Feature | Branch | Status |
|---------|--------|--------|
| feat-00-foundation | feat/foundation | done |
| feat-01-templates-embed | feat/templates-embed | done |
| feat-02-init-interactive | feat/init-interactive | done |
| feat-03-init-flags | feat/init-flags | done |
| feat-04-dry-run | feat/dry-run | done |
| feat-05-versioning | feat/versioning | done |
| feat-06-self-test | feat/self-test | done |
| feat-07-release | feat/release | done |
| feat-08-npx | feat/npx | done |
| feat-09-multi-agent | feat/multi-agent | done |
| feat-10-landing-page | feat/landing-page | done |
| feat-11-unify-agent-structures | feat/unify-agent-structures | done |
| feat-14-discovery-command | feat/discovery-command | done |
| feat-16-constitution-command | feat/constitution-command | done |
| feat-17-unified-onboarding-docs | feat/unified-onboarding-docs | done |
| feat-21-install-skill-from-url | feat/install-skill-from-url | done |
| feat-22-smart-upgrade-detection | feat/smart-upgrade-detection | done |
| feat-23-agent-specific-mcp-config | feat/agent-specific-mcp-config | done |
| feat-24-automated-handoff-flow | feat/automated-handoff-flow | done |
| feat-30-fix-branch-pr-flow | feat/fix-branch-pr-flow | done |
| feat-31-upgrade-preserves-domain | feat/upgrade-preserves-domain | done |
| feat-32-product-release-notes | feat/product-release-notes | done |
| feat-33-lp-wiki-sync | feat/sync-wiki-documentation | done |
| feat-34-update-agent-prompts | feat/update-agent-prompts | done |
| feat-35-remove-install-skill | feat/remove-install-skill | done |
| feat-36-structured-documentation | feat/structured-documentation | done |
| feat-37-publish-1.6.0 | feat/publish-1.6.0 | done |
| feat-38-cli-detect-existing | feat/cli-detect-existing | done |
| feat-39-cli-doctor-destroy | feat/cli-doctor-destroy | done |
| feat-40-cli-init-refinement | feat/cli-init-refinement | done |
| feat-41-nested-features-discovery | feat/cli-nested-features-discovery | done |
| feat-42-release-v1.7.0 | feat/release-v1.7.0 | done |
| feat-43-cli-shortcut-and-welcome | feat/cli-shortcut-and-welcome | done |
| feat-43a2-telemetry-metrics-nested | fix/telemetry-metrics-nested | done |
| feat-5ae2-01-cheat-sheet-init | feat/5ae2-cheat-sheet-init | done |
| feat-5ae2-02-status-prescritivo | feat/5ae2-status-prescritivo | done |
| feat-5ae2-03-doctor-deriva-convencao | feat/5ae2-doctor-deriva-convencao | done |
| feat-5ae2-04-tutorial-guiado | feat/5ae2-tutorial-guiado | done |
| feat-5ae2-05-gate-graduacao-autopilot | feat/5ae2-gate-graduacao-autopilot | done |
| feat-5ae2-06-spike-subagentes-nativos | feat/5ae2-spike-subagentes-nativos | done |
| feat-5ae2-07-modo-iniciante | feat/5ae2-modo-iniciante | done |
| fix-5ae2-08-naming-telemetry-tokens | fix/naming-telemetry-tokens | done |
| fix-45-update-beta-version-detection | fix/update-beta-version-detection | done |
| fix-46-doctor-metrics-path | fix/update-beta-version-detection | done |
| fix-47-naming-convention-update-flag | fix/naming-convention-not-applied-on-update | done |
| fix-48-novo-fix-missing-copilot-agent | fix/novo-fix-missing-copilot-agent | done |

## Próximo passo
**Iniciar:** Nenhuma feature `todo` pendente. fix-47 e fix-48 agrupadas na v1.9.2.
**Bloqueios:** —

## Handoff da última sessão
- fix-48-novo-fix-missing-copilot-agent concluída: `/novo-fix` nunca existiu para o agente Copilot (só Claude/Gemini tinham o template) — corrigido criando `internal/scaffold/templates/.github/prompts/novo-fix.prompt.md.tmpl`, adicionando `"novo-fix"` a `commandOrder` (`cheatsheet.go`) e citando o comando em `CLAUDE.md.tmpl`/`GEMINI.md.tmpl`. OpenAI ficou de fora — gap pré-existente maior (nenhum prompt implementado), não específico do `novo-fix`. Golden fixtures regeneradas.
- fix-47-naming-convention-update-flag concluída: `forge-sdd update`/`init` (redirecionado para update em projeto existente) ignorava completamente a flag `--naming-convention` — `updateCmd` nem registrava a flag, e `runUpdateFlow` tinha dois early-returns que abortavam antes de persistir a mudança quando nenhuma outra flag (`--agent`/`--upgrade`/`--version`) era combinada. Corrigido em `cmd/forge-sdd/main.go`: flag registrada em `updateCmd`, lida e aplicada em `rc.NamingConvention` logo após `ReadSddrc`, com os early-returns ajustados para não descartar a mudança. Validado manualmente com `forge-sdd update --yes --naming-convention workitem`.
- v1.9.1-beta promovida a estável (v1.9.1): fix-45-update-beta-version-detection e fix-46-doctor-metrics-path incorporadas via cherry-pick — `--upgrade --yes` agora resolve a versão via `config.ResolveUpgradeTarget` (consultando `FetchNpmVersions`) em vez da constante `version` do binário; `--upgrade` sem `--yes` pré-seleciona a opção de upgrade no prompt interativo; falha de rede reporta erro/aviso explícito em vez de fallback silencioso; timeout de `FetchNpmVersions` elevado de 2s para 8s; prompts `/doctor` corrigidos para verificar `sdd/.metrics/schema.json` em vez da raiz. Testes novos em `internal/config/config_test.go` e `cmd/forge-sdd/commands_test.go`.
- feat-5ae2-07-modo-iniciante concluída: prompts `/constitution` (3 agentes) ganharam a pergunta opcional de nível de linguagem (`padrão`/`iniciante`), persistida em `constitution.md`; `/discovery` ganhou instrução para simplificar jargão quando `iniciante`. Discovery 5ae2 100% implementada (7/7), alvo do pacote v1.9.0.
- feat-5ae2-06-spike-subagentes-nativos concluída (spike, sem código): mapeado suporte nativo a subagentes com contexto isolado por agente — Claude tem primitivo maduro (`.claude/agents/*.md`), Gemini/Copilot não confirmados/ausentes hoje. Recomendação: piloto restrito ao Claude, sem migrar os três agentes de uma vez (quebraria paridade de comportamento).
- feat-5ae2-05-gate-graduacao-autopilot concluída: novo comando `forge-sdd autopilot` bloqueia a criação de `sdd/.sdd-auto-pilot` até N ciclos completos (`outcome: approved`) em telemetria, com bypass consciente via `--skip-graduation`. Independente do autopilot em si (que segue só na branch `feat/cli-autopilot-autonomy`, ainda em teste).
- feat-5ae2-04-tutorial-guiado concluída: novo prompt `/tutorial` (Claude, Gemini, Copilot) guia um ciclo SDD fictício isolado; `forge-sdd init --tutorial` sugere rodá-lo ao final do scaffold.
- feat-5ae2-03-doctor-deriva-convencao concluída: `forge-sdd doctor` agora detecta e avisa quando convenções sequencial (`feat-NN`) e hash (`feat-xxxx`) coexistem em `sdd/features/` e `sdd/discovery/` — confirmado usando o próprio repositório do Forge-SDD como caso real.
- feat-5ae2-02-status-prescritivo concluída: os três prompts `/status` (Claude, Gemini, Copilot) agora encerram sempre com "Próximo comando sugerido: <comando>", calculado a partir do estado de `sdd/discovery/` e `sdd/features/`.
- feat-5ae2-01-cheat-sheet-init concluída: `forge-sdd init` agora imprime a lista dos comandos SDD disponíveis para os agentes selecionados ao final de uma execução bem-sucedida (fora de `--dry-run`).
- Discovery 5ae2 (curva de aprendizado do Forge-SDD na era dos agentes de IA) quebrada em 7 features via `/split-features`, agrupadas em `sdd/features/feat-5ae2-curva-aprendizado-agentes-ia/`, alvo do pacote v1.9.0.
- Feature 42 concluída (consolidação de todas as versões beta na main, bump de versão para 1.7.0, atualização de golden files e publicação estável v1.7.0 no NPM).

## Última sessão
- 2026-07-20 — fix: implementadas, testadas e validadas as fix-45 (detecção de versão beta no `update --upgrade`) e fix-46 (caminho incorreto de `.metrics/schema.json` nos prompts `/doctor`), agrupadas na mesma branch/PR.
- 2026-07-09 — feat: concluída a Fase 42 (consolidação e publicação oficial da versão estável v1.7.0).
- 2026-07-06 — feat: concluída a Fase 41 (agrupamento em subpastas, CLI doctor recursivo e lançamento da v1.6.1-beta.3).
- 2026-07-05 — feat: concluída a Fase 40 (refinamento do init, criação de subpastas, relatórios e lançamento da v1.6.1-beta.2).
- 2026-07-05 — feat: concluída a Fase 39 (implementação de comandos doctor/destroy e lançamento da v1.6.1-beta.1).
- 2026-07-04 — feat: concluída a Fase 38 (detecção inteligente de projeto e lançamento da v1.6.1-beta.0).
- 2026-07-03 — feat: concluída a Fase 37 (publicação oficial da versão estável v1.6.0).
- 2026-06-30 — feat: concluída a Fase 36 (criação da documentação estruturada com diagramas conceituais Mermaid detalhados).
- 2026-06-29 — feat: concluída a Fase 35 (remoção da funcionalidade /install-skill nos três agentes e documentação).
- 2026-06-29 — feat: concluída a Fase 34 (migração dos prompts do Copilot para nova especificação).
- 2026-06-28 — feat: concluída a Fase 33 (sincronização automática da wiki no GitHub Actions).
- 2026-06-28 — feat: concluída a Fase 23 (MCP e Habilidades Específicas por Agente) e finalizado o roadmap do forge-sdd.

> Histórico completo em `progress-log.md`


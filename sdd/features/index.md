# Index de Features — forge-sdd

## Dependency graph

```
main
└─ feat/foundation
   └─ feat/templates-embed
      └─ feat/init-interactive
         └─ feat/init-flags
            └─ feat/dry-run
               └─ feat/versioning
                  └─ feat/self-test
                     └─ feat/release
            └─ feat/npx
               └─ feat/multi-agent
                  └─ feat/unify-agent-structures
                     └─ feat/npm-release-and-deploy
                        └─ feat/update-readmes-multiagent
                           └─ feat/discovery-command
                              └─ feat/sdd-guardrails
                           └─ feat/constitution-command
                              └─ feat/unified-onboarding-docs
                                 └─ feat/sdd-handoffs
                                 └─ feat/c4-model-mermaid
                                    └─ feat/advanced-discovery-and-c4-refinement
                                       └─ feat/install-skill-from-url
                                          └─ feat/smart-upgrade-detection
                                          └─ feat/agent-specific-mcp-config
                                          └─ feat/automated-handoff-flow
                                        └─ feat/landing-page  (paralela)
                                        └─ feat/open-source-readme (paralela)
                                        └─ feat/upgrade-node-ci (paralela)
                                        └─ feat/dev-flow-guide (paralela)
                                           └─ feat/remove-homebrew
                                              └─ feat/fix-branch-pr-flow
                                                  └─ feat/upgrade-preserves-domain
                                                     └─ feat/product-release-notes
                                                         └─ feat/sync-wiki-documentation
                                                             └─ feat/update-agent-prompts
                                                                 └─ feat/publish-1.6.0
                                                                     └─ feat/cli-detect-existing
                                                                         └─ feat/cli-doctor-destroy
                                                                             └─ feat/cli-init-refinement
                                                                                 └─ feat/cli-nested-features-discovery
                                                                                     └─ feat/release-v1.7.0
                                                                                         └─ fix/telemetry-metrics-nested
                                                                                             └─ feat/cli-shortcut-and-welcome
                                                                                                 └─ feat/5ae2-cheat-sheet-init
                                                                                                     └─ feat/5ae2-status-prescritivo
                                                                                                     └─ feat/5ae2-doctor-deriva-convencao
                                                                                                     └─ feat/5ae2-tutorial-guiado (paralela)
                                                                                                     └─ feat/5ae2-gate-graduacao-autopilot (paralela)
                                                                                                     └─ feat/5ae2-spike-subagentes-nativos (paralela)
                                                                                                     └─ feat/5ae2-modo-iniciante (paralela)
                                                                                                         └─ feat/01-telemetria-code-enforced
                                                                                                             └─ feat/01-agregador-telemetria
                                                                                                             └─ feat/01-lessons-artefato
                                                                                                                 └─ feat/01-consulta-lessons-builder-revisor
                                                                                                             └─ feat/01-ferramentas-configuraveis-constitution (paralela)
                                                                                                             └─ feat/01-lifecycle-unico
                                                                                                             └─ feat/01-reducao-duplicacao-naming
                                                                                                             └─ feat/01-auditoria-comandos-sobrepostos
            ```

                                       ## Janelas de paralelismo
                                       feat-5ae2-04, feat-5ae2-05, feat-5ae2-06 e feat-5ae2-07 podem ser desenvolvidas em paralelo entre si após feat-5ae2-03 (não têm dependência direta entre elas).
                                       feat-01-05 (ferramentas configuráveis) é independente da cadeia de telemetria e pode ser desenvolvida em paralelo a feat-01-01..04. feat-01-06, feat-01-07 e feat-01-08 não dependem entre si e podem ser paralelas, mas são priorizadas por último (maior superfície de revisão).

                                       ## Índice

                                       | # | Arquivo | Branch | Fase | Status |
                                       |---|---------|--------|------|--------|
                                       | 00 | feat-00-foundation.md | feat/foundation | 0 | done |
                                       | 01 | feat-01-templates-embed.md | feat/templates-embed | 1 | done |
                                       | 02 | feat-02-init-interactive.md | feat/init-interactive | 2 | todo |
                                       | 03 | feat-03-init-flags.md | feat/init-flags | 3 | todo |
                                       | 04 | feat-04-dry-run.md | feat/dry-run | 4 | todo |
                                       | 05 | feat-05-versioning.md | feat/versioning | 5 | todo |
                                       | 06 | feat-06-self-test.md | feat/self-test | 6 | todo |
                                       | 07 | feat-07-release.md | feat/release | 7 | done |
                                       | 08 | feat-08-npx.md | feat/npx | 8 | done |
                                       | 09 | feat-09-multi-agent.md | feat/multi-agent | 9 | done |
                                       | 10 | feat-10-landing-page.md | feat/landing-page | 10 | done |
                                       | 11 | feat-11-unify-agent-structures.md | feat/unify-agent-structures | 11 | done |
                                       | 14 | feat-14-discovery-command.md | feat/discovery-command | 14 | done |
                                       | 16 | feat-16-constitution-command.md | feat/constitution-command | 16 | done |
                                       | 17 | feat-17-unified-onboarding-docs.md | feat/unified-onboarding-docs | 17 | done |
                                       | 18 | feat-18-sdd-handoffs.md | feat/sdd-handoffs | 18 | done |
                                       | 19 | feat-19-c4-model-mermaid.md | feat/c4-model-mermaid | 19 | done |
                                       | 21 | feat-21-install-skill-from-url.md | feat/install-skill-from-url | 21 | done |
                                       | 22 | feat-22-smart-upgrade-detection.md | feat/smart-upgrade-detection | 22 | done |
                                       | 23 | feat-23-agent-specific-mcp-config.md | feat/agent-specific-mcp-config | 23 | done |
| 24 | feat-24-automated-handoff-flow.md | feat/automated-handoff-flow | 24 | done |
| 26 | feat-26-open-source-readme.md | feat/open-source-readme | 26 | done |
| 27 | feat-27-upgrade-node-ci.md | feat/upgrade-node-ci | 27 | done |
| 28 | feat-28-dev-flow-guide.md | feat/dev-flow-guide | 28 | done |
                                         | 29 | feat-29-remove-homebrew.md | feat/remove-homebrew | 29 | done |
                                         | 30 | feat-30-fix-branch-pr-flow.md | feat/fix-branch-pr-flow | 30 | done |
                                         | 31 | feat-31-upgrade-preserves-domain.md | feat/upgrade-preserves-domain | 31 | done |
| 32 | feat-32-product-release-notes.md | feat/product-release-notes | 32 | done |
| 33 | feat-33-lp-wiki-sync.md | feat/sync-wiki-documentation | 33 | done |
| 34 | feat-34-update-agent-prompts.md | feat/update-agent-prompts | 34 | done |
| 35 | feat-35-remove-install-skill.md | feat/remove-install-skill | 35 | done |
| 36 | feat-36-structured-documentation.md | feat/structured-documentation | 36 | done |
| 37 | feat-37-publish-1.6.0.md | feat/publish-1.6.0 | 37 | done |
| 38 | feat-38-cli-detect-existing.md | feat/cli-detect-existing | 38 | done |
| 39 | feat-39-cli-doctor-destroy.md | feat/cli-doctor-destroy | 39 | done |
| 40 | feat-40-cli-init-refinement.md | feat/cli-init-refinement | 40 | done |
| 41 | feat-41-nested-features-discovery.md | feat/cli-nested-features-discovery | 41 | done |
| 42 | feat-42-release-v1.7.0.md | feat/release-v1.7.0 | 42 | done |
| 43a2 | feat-43a2-telemetry-metrics-nested.md | fix/telemetry-metrics-nested | 43a2 | doing |
| 43 | feat-43-cli-shortcut-and-welcome.md | feat/cli-shortcut-and-welcome | 1.9.0-beta | done |
| 5ae2-01 | feat-5ae2-curva-aprendizado-agentes-ia/feat-5ae2-01-cheat-sheet-init.md | feat/5ae2-cheat-sheet-init | 5ae2-01 | done |
| 5ae2-02 | feat-5ae2-curva-aprendizado-agentes-ia/feat-5ae2-02-status-prescritivo.md | feat/5ae2-status-prescritivo | 5ae2-02 | done |
| 5ae2-03 | feat-5ae2-curva-aprendizado-agentes-ia/feat-5ae2-03-doctor-deriva-convencao.md | feat/5ae2-doctor-deriva-convencao | 5ae2-03 | done |
| 5ae2-04 | feat-5ae2-curva-aprendizado-agentes-ia/feat-5ae2-04-tutorial-guiado.md | feat/5ae2-tutorial-guiado | 5ae2-04 | done |
| 5ae2-05 | feat-5ae2-curva-aprendizado-agentes-ia/feat-5ae2-05-gate-graduacao-autopilot.md | feat/5ae2-gate-graduacao-autopilot | 5ae2-05 | done |
| 5ae2-06 | feat-5ae2-curva-aprendizado-agentes-ia/feat-5ae2-06-spike-subagentes-nativos.md | feat/5ae2-spike-subagentes-nativos | 5ae2-06 | done |
| 5ae2-07 | feat-5ae2-curva-aprendizado-agentes-ia/feat-5ae2-07-modo-iniciante.md | feat/5ae2-modo-iniciante | 5ae2-07 | done |
| 5ae2-08 | feat-5ae2-curva-aprendizado-agentes-ia/fix-5ae2-08-naming-telemetry-tokens.md | fix/naming-telemetry-tokens | 5ae2-08 | done |
| 45 | fix-45-update-beta-version-detection.md | fix/update-beta-version-detection | 45 | done |
| 46 | fix-46-doctor-metrics-path.md | fix/update-beta-version-detection | 46 | done |
| 47 | fix-47-naming-convention-update-flag.md | fix/naming-convention-not-applied-on-update | 47 | done |
| 48 | fix-48-novo-fix-missing-copilot-agent.md | fix/novo-fix-missing-copilot-agent | 48 | done |
| 50 | fix-50-telemetry-recording-gemini-only.md | fix/telemetry-recording-gemini-only | 50 | done |
| 01-01 | feat-01-simplificacao-e-aprendizado-continuo/feat-01-01-telemetria-code-enforced.md | feat/01-telemetria-code-enforced | 01-01 | done |
| 01-02 | feat-01-simplificacao-e-aprendizado-continuo/feat-01-02-agregador-telemetria.md | feat/01-agregador-telemetria | 01-02 | todo |
| 01-03 | feat-01-simplificacao-e-aprendizado-continuo/feat-01-03-lessons-artefato.md | feat/01-lessons-artefato | 01-03 | todo |
| 01-04 | feat-01-simplificacao-e-aprendizado-continuo/feat-01-04-consulta-lessons-builder-revisor.md | feat/01-consulta-lessons-builder-revisor | 01-04 | todo |
| 01-05 | feat-01-simplificacao-e-aprendizado-continuo/feat-01-05-ferramentas-configuraveis-constitution.md | feat/01-ferramentas-configuraveis-constitution | 01-05 | todo |
| 01-06 | feat-01-simplificacao-e-aprendizado-continuo/feat-01-06-lifecycle-unico.md | feat/01-lifecycle-unico | 01-06 | todo |
| 01-07 | feat-01-simplificacao-e-aprendizado-continuo/feat-01-07-reducao-duplicacao-naming.md | feat/01-reducao-duplicacao-naming | 01-07 | todo |
| 01-08 | feat-01-simplificacao-e-aprendizado-continuo/feat-01-08-auditoria-comandos-sobrepostos.md | feat/01-auditoria-comandos-sobrepostos | 01-08 | todo |

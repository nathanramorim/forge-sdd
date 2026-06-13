# feat/sdd-guardrails

**Branch:** `feat/sdd-guardrails`
**Fase:** 15
**Depende de:** `feat-11-unify-agent-structures`, `feat-14-discovery-command`
**Status:** `todo`

## Objetivo

Implementar mecanismos de "Guardrails" (travas de segurança e qualidade) em cada etapa do ciclo de vida SDD. O objetivo é garantir que o agente não prossiga para a próxima fase se requisitos críticos (como aderência à Constitution, limites de tokens ou critérios técnicos) não forem atendidos.

## Critério de conclusão

```bash
# 1. Simular uma falha de Guardrail (ex: progress.md > 1KB)
# O Orquestrador deve bloquear a sessão e exigir a execução do /archive.

# 2. Verificar novos campos nos templates de prompt
grep "Guardrails" internal/scaffold/templates/agents/gemini/.gemini/prompts/*.prompt.md.tmpl
```

## Tarefas

- [ ] **15-1** Definir a lista canônica de Guardrails por fase (Plan, Act, Review, Close)
- [ ] **15-2** Atualizar `sdd/memory/constitution.md` para incluir a seção de Guardrails obrigatórios
- [ ] **15-3** Implementar verificações de Guardrails nos prompts do Gemini (`.gemini/prompts/`)
- [ ] **15-4** Implementar verificações de Guardrails nos prompts do Copilot (`.github/prompts/`)
- [ ] **15-5** Atualizar as skills de `Orquestrador` e `Revisor` para serem os "enforcers" dos guardrails
- [ ] **15-6** Documentar o protocolo de Guardrails na `docs/metodologia-sdd.md`

## Guardrails Propostos

1. **Pre-Flight (Plan):** Validar se `progress.md` está dentro do budget antes de iniciar.
2. **Implementation (Act):** Builder deve validar `go vet` ou linter equivalente antes de reportar conclusão.
3. **Quality (Review):** Revisor deve validar se apenas arquivos declarados na feat foram alterados.
4. **Post-Flight (Close):** Orquestrador deve validar se as métricas foram gravadas antes de encerrar.

## Arquivos gerados/modificados

```
internal/scaffold/templates/
docs/metodologia-sdd.md
sdd/memory/constitution.md
.gemini/prompts/
.github/prompts/
```

## Skills relevantes

- `orquestrador.chatmode.md` (Workflow Control)
- `revisor.chatmode.md` (Quality Assurance)

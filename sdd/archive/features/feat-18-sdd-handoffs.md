# feat/sdd-handoffs

**Branch:** `feat/sdd-handoffs`
**Fase:** 18
**Depende de:** `feat-14-discovery-command`, `feat-17-unified-onboarding-docs`
**Status:** `done`

## Objetivo

Implementar um Protocolo de Handoff (passagem de bastão) entre as etapas do SDD. O objetivo é garantir que a saída de uma fase (ex: Discovery) seja entregue de forma estruturada para a próxima (ex: Especificação de Features), evitando perda de contexto e garantindo que o agente seguinte saiba exatamente por onde começar.

## Critério de conclusão

- [x] Prompts `/discovery`, `/nova-feature` e `/revisar` terminam com uma seção clara de "Handoff".
- [x] O arquivo `sdd/memory/progress.md` possui um campo de `Handoff da última sessão` para persistir o contexto entre turnos.
- [x] Metodologia atualizada com o "Protocolo de Handoff (§10)".

## Tarefas

- [x] **18-1** Definir o template de Handoff para cada transição.
- [x] **18-2** Atualizar `progress.md` para incluir a seção `Handoff Context`.
- [x] **18-3** Implementar instruções de Handoff nos templates de prompt do Gemini, Copilot e Claude.
- [x] **18-4** Atualizar as skills de `Specifier`, `Builder` e `Revisor` para gerar o handoff ao final de suas ações.
- [x] **18-5** Documentar o Protocolo de Handoff na `docs/metodologia-sdd.md`.

## Fluxo de Handoff (Exemplo)

1. **Agente A (Discovery)** termina a análise.
2. **Handoff:** Gera um resumo estruturado: "Foco: MVP de Pagamentos. Arquivo base: discovery-01.md. Próximo passo: /nova-feature para criar as tarefas de integração Stripe."
3. **Agente B (Specifier)** lê o Handoff e já sabe qual arquivo de discovery processar.

## Arquivos gerados/modificados

```
docs/metodologia-sdd.md
internal/scaffold/templates/sdd/memory/progress.md.tmpl
internal/scaffold/templates/agents/*/prompts/
```

## Skills relevantes

- `orquestrador.chatmode.md` (Context Management)
- `specifier.chatmode.md` (Transition logic)

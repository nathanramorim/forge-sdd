# feat/advanced-discovery-and-c4-refinement

**Branch:** `feat/advanced-discovery-and-c4-refinement`
**Fase:** 20
**Depende de:** `feat-14-discovery-command`, `feat-19-c4-model-mermaid`
**Status:** `todo`

## Objetivo

Refinar a visualização arquitetural e expandir as capacidades de planejamento do SDD.
1. **Refinamento C4:** Garantir que o agente use *conceitos* C4 mas renderize diagramas usando apenas `flowchart` e `sequence diagram` padrão do Mermaid (evitando a sintaxe experimental `C4Context`).
2. **Auto-Planning:** O comando `/discovery` deve gerar automaticamente um arquivo `sdd/discovery/plan-XX.md` com o roadmap de tarefas sugerido.
3. **Novo Comando `/split-features`:** Criar um comando para quebrar um plano de discovery em múltiplas features independentes.
4. **Preservação do `/nova-feature`:** Manter o comando atual focado em incrementos pontuais.

## Critério de conclusão

- [ ] Skill `c4-architecture` instrui explicitamente o uso de `flowchart TD/LR` e `sequenceDiagram`.
- [ ] Comando `/discovery` gera: `discovery-XX.md`, `criteria-XX.md` e `plan-XX.md`.
- [ ] Novo comando `/split-features` implementado em todos os agentes.
- [ ] Teste de integração validando a geração do arquivo de plano.

## Tarefas

- [ ] **20-1** Ajustar skill `c4-architecture.chatmode.md` para proibir sintaxe `C4Context` e exigir `flowchart`/`sequenceDiagram`.
- [ ] **20-2** Atualizar prompts de `/discovery` para incluir a geração do `plan-XX.md`.
- [ ] **20-3** Criar prompt `/split-features` para Gemini, Copilot e Claude.
- [ ] **20-4** Atualizar a lógica do `Specifier` para gerenciar a quebra de planos em múltiplas `feat-XX.md`.
- [ ] **20-5** Atualizar documentação universal (READMEs e Methodology).

## Arquivos gerados/modificados

```
internal/scaffold/templates/agents/*/prompts/
internal/scaffold/templates/agents/*/skills/
docs/metodologia-sdd.md
README.md
```

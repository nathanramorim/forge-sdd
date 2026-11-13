# feat/c4-model-mermaid

**Branch:** `feat/c4-model-mermaid`
**Fase:** 19
**Depende de:** `feat-14-discovery-command`, `feat-16-constitution-command`
**Status:** `done`

## Objetivo

Acrescentar ao SDD a habilidade de descrever arquiteturas complexas utilizando o framework **C4 Model** representado em **Mermaid**. O foco é utilizar `flowchart` para os níveis 1 (Contexto) e 2 (Container), e `sequence diagram` para detalhar interações críticas no nível 3 (Componente), garantindo uma visualização clara e bem distribuída.

## Critério de conclusão

- [x] Nova skill `c4-architecture.chatmode.md` adicionada ao scaffold.
- [x] O comando `/discovery` e `/constitution` são instruídos a sugerir diagramas C4 quando a complexidade exigir.
- [x] Template `sdd/spec/overview.md.tmpl` atualizado com placeholders para diagramas Mermaid.

## Tarefas

- [x] **19-1** Criar a skill `c4-architecture.chatmode.md` detalhando as regras de mapeamento C4 em Mermaid.
- [x] **19-2** Atualizar o template `sdd/spec/overview.md.tmpl` para incluir seções de arquitetura visual.
- [x] **19-3** Instruir o `Specifier` a utilizar diagramas de sequência Mermaid para fluxos de dados complexos (C4 Nível 3).
- [x] **19-4** Adicionar suporte a diagramas C4 nos prompts de `/discovery` e `/constitution`.
- [x] **19-5** Atualizar `sdd/skills/index.md` para incluir a nova habilidade de arquitetura visual.

## Regras de Diagramação (Guardrails)

1. **Nível 1 (Contexto):** Flowchart mostrando sistemas externos e usuários.
2. **Nível 2 (Container):** Flowchart detalhando aplicações, bancos e serviços.
3. **Nível 3 (Componente):** Sequence Diagram para fluxos de API/Eventos críticos.
4. **Nível 4 (Código):** Apenas se solicitado explicitamente (Markdown estruturado).

## Arquivos gerados/modificados

```
internal/scaffold/templates/sdd/spec/overview.md.tmpl
internal/scaffold/templates/agents/gemini/.gemini/skills/c4-architecture.chatmode.md.tmpl
internal/scaffold/templates/sdd/skills/index.md.tmpl
docs/metodologia-sdd.md
```

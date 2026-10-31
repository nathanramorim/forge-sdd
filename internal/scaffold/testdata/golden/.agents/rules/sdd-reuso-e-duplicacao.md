---
description: Como executar o Reuse Check, manter o catálogo de conceitos, aplicar a Regra de Três para Auto-Skills e preservar fonte única da verdade.
globs: ["sdd/**", "internal/**", ".agents/**"]
regras: [DUP-01, DUP-02, DUP-05, DUP-10, GOV-04]
---

# Guia Prático — Reúso, Anti-Duplicação e Fonte Única

Passos práticos para cumprir `DUP-01`, `DUP-02`, `DUP-05`, `DUP-10` e `GOV-04` definidos em `sdd/memory/constitution.md`.

## 1. Catálogo de Conceitos e Reuse Check Obrigatório (`DUP-01`, `DUP-02`)

- **Consulta ao catálogo (`DUP-01`):**
  - Antes de criar uma nova struct de configuração, estado ou comando, consulte `sdd/spec/concepts.md` para identificar qual pacote em `internal/` já é dono daquele conceito.
  - Se um novo conceito de domínio nascer na entrega, adicione sua linha correspondente em `sdd/spec/concepts.md` apontando o pacote dono e o contrato público.
- **Preenchimento de `## O que já existe` (`DUP-02`):**
  - Toda spec (`criteria-ID-*.md` e `feat-XX-*.md`) deve conter uma tabela com as colunas `| Procurei | Onde | Achei | Decisão (reusar, estender, criar) | Justificativa |`.
  - A decisão `criar` só é aceita quando acompanhada de justificativa objetiva explicando por que os pacotes listados em `Achei` não podem ser reusados ou estendidos.

## 2. Fonte Única para Artefatos de Agentes e Rules (`DUP-10`, `GOV-04`)

- **Corpo canônico em `.agents/` (`DUP-10`):**
  - Comandos residem em `.agents/commands/<cmd>.md`, papéis em `.agents/roles/<papel>.md`, regras de domínio em `.agents/rules/<regra>.md` e skills em `.agents/skills/<slug>/SKILL.md`.
  - Os arquivos em `.claude/`, `.gemini/` e `.github/` devem funcionar apenas como adaptadores finos que apontam para o arquivo canônico em `.agents/`, evitando editar o mesmo prompt em três lugares diferentes.
- **Referência por ID nas Rules (`GOV-04`):**
  - Ao criar ou editar qualquer arquivo em `.agents/rules/*.md`, declare no frontmatter YAML a chave `regras: [ID-01, ID-02]` com IDs presentes em `sdd/memory/constitution.md`.
  - Escreva no corpo apenas procedimentos práticos ("como fazer/verificar"), sem copiar as frases declarativas da Constituição.

## 3. Extração de Padrões e Auto-Skills pela Regra de Três (`DUP-05`)

- Na 1ª ocorrência de um procedimento, resolva diretamente na task.
- Na 2ª ocorrência similar, observe a repetição e mantenha o código/fluxo alinhado.
- Na **3ª ocorrência** de um mesmo bloco de código ou roteiro operacional:
  - Se for lógica em Go, extraia uma função ou pacote interno compartilhado.
  - Se for procedimento operacional de agente (ex.: sequência padrão de criar subcomando + template + golden test), proponha uma Auto-Skill em `.agents/skills/<slug>/SKILL.md` via `forge-sdd skills propose` (ou no campo `candidate_skill` da revisão semântica).

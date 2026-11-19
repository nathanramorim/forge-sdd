# fix-88-06 — Veredito autoexplicado: descrição da regra inline

**Branch:** `fix/88-06-veredito-com-descricao-da-regra`
**Origem:** Iniciativa 88, `.agents/skills/sdd-governanca/references/veredito.md`
**Tamanho:** `S`
**Status:** `done`

## Sintoma

O contrato do veredito (`veredito.md`) tinha `"regra": "FLX-08"` mas nenhum campo para a descrição. O Reviewer citava o ID isolado e o usuário no terminal não sabia o que `FLX-08`, `GOV-04` ou `SEG-01` significavam sem abrir `constitution.md` ou rodar `forge rules`. O veredito não era autoexplicado.

## Feature Correlacionada

- **Arquivos alterados:**
  - `.agents/skills/sdd-governanca/references/veredito.md` — schema do JSON de veredito
  - `.agents/skills/sdd-governanca/SKILL.md` — instruções do Revisor
  - `internal/scaffold/templates/.agents/skills/sdd-governanca/references/veredito.md.tmpl`
  - `internal/scaffold/testdata/golden/.agents/skills/sdd-governanca/references/veredito.md`

## Decisões do projeto aplicadas

- GOV-04 (rules explicam o como cumprir)
- COD-01 (responsabilidade única — veredito autoexplicado)

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Campo `regra_descricao` | `veredito.md` | Ausente | **adicionado** |
| Instrução para popular descrição | `SKILL.md` governança | Ausente | **adicionada** |

## Escopo implementado

1. `veredito.md`: campo `"regra_descricao"` adicionado ao schema + nota de obrigatoriedade
2. `SKILL.md`: regra de ouro atualizada para exigir `regra_descricao` com texto completo
3. Template e golden atualizados em sincronia

## Critérios de Aceitação

```bash
grep "regra_descricao" .agents/skills/sdd-governanca/references/veredito.md        # ✓
grep "regra_descricao" .agents/skills/sdd-governanca/SKILL.md                      # ✓
grep "regra_descricao" internal/scaffold/templates/.agents/skills/sdd-governanca/references/veredito.md.tmpl  # ✓
grep "regra_descricao" internal/scaffold/testdata/golden/.agents/skills/sdd-governanca/references/veredito.md # ✓
go test ./...  # ✓ verde
```

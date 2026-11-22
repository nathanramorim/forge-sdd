# Feature 115-03-03 — Rubrica de complexidade e Advisor

## Modelo do módulo

Enxuto (`ARQ-11`): atualizar skills comuns e formato de task; sem código de runtime ou dependências novas.

## Decisões do projeto aplicadas

- ADR-0003: skills por estação e contexto controlado.
- ADR-0014 D-5: respeitar modelo configurado por papel quando task antiga não tem tier.
- ADR-0024 (proposta): classificar por task; raciocínio avançado condicional; policy neutra de CLI.
- Iniciativa 110, E-09: estimativas históricas podem apoiar, mas não substituem a rubrica.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Concepção | `.agents/skills/sdd-concepcao/` | Sem rubrica operacional nem Advisor condicional | Acrescentar critérios e saída estruturada | Tornar tier explícito no split |
| Planejamento e build | `.agents/skills/sdd-planejamento/`, `.agents/skills/sdd-build/` | Handoff não exige preservar tier | Atualizar contrato compartilhado | Tier acompanha task ao job |
| Orquestração | skills de operação/liderança disponíveis | Variação por CLI | Usar tiers lógicos neutros de provedor | Consistência geral |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`COD-01`, `COD-05`, `DUP-01`, `FLX-01`, `SEG-01`, `TST-03`.

## Critérios de Aceitação Executáveis

1. Skills descrevem low/fast, medium/balanced e high/balanced para Builder, com Advisor reasoning só diante de incerteza arquitetural e impacto.
2. Cada task produz complexidade, justificativa e model_tier; task clara não herda complexidade da iniciativa.
3. Exemplos cobrem Claude (Haiku/Sonnet/Opus) e instruem cada outra CLI a usar seus modelos configurados sem assumir equivalência de nomes.
4. Referências de modelo apontam para documentação oficial vigente e a instrução respeita gates humanos.

```sh
git diff --check
```

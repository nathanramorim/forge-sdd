# ADR-0003: Skills por Estação e Orçamento de Contexto

- **Data:** 2026-11-13
- **Status:** `aprovado`
- **Regras afetadas (IDs da constitution):** `DUP-10`, `ARQ-06`, `FLX-02`, `FLX-06`, `DUP-06`, `DUP-01`, `GOV-01` (este ADR)
- **Autor / Responsável:** Líder (decisão do dono do produto no discovery 69)

## Contexto

Um `/discovery` obrigava a ler cerca de 65 KB (o `index.md` sozinho tinha 27 KB) e o Builder cerca de 35 KB por task. As instruções dos agentes tinham quatro camadas (adaptador, comando, papel, rules), replicadas em 63 arquivos de adaptadores, com 13 comandos e 7 papéis (cerca de 65 KB em `.agents/`). A Constitution escrevia cada regra duas vezes (Regras 0 a 15 e IDs canônicos) e mantinha um mapa rule para IDs que repetia o frontmatter das próprias rules. Referência: `sdd/discovery/69-otimizacao-estrutura-tokens-e-paralelismo/`, `feat-69-06` e `feat-69-07`.

## Opções Consideradas

1. **Só agrupar as auto-skills geradas:** mantém comandos e papéis como estão; não ataca o custo fixo.
2. **Consolidar comandos e papéis em skills por estação:** cinco skills curtas com `references/` carregadas sob demanda, e adaptadores finos.
3. **As duas coisas:** inclui o agrupamento de auto-skills por domínio, que ainda não tem massa crítica (há uma auto-skill).

## Decisão

Opção 2, com orçamentos verificáveis:
- Cinco skills em `.agents/skills/` (`sdd-concepcao`, `sdd-planejamento`, `sdd-build`, `sdd-governanca`, `sdd-operacao`), `scope: core`: o `skills sync` não as copia para `.claude/skills` nem `.gemini/skills`, e `skills approve` as recusa. `.agents/commands/` e `.agents/roles/` deixam de existir; o `update` remove os antigos apenas quando idênticos ao template anterior (ADR 0002). `DUP-10` passa a nomear `.agents/skills/` e `.agents/rules/` como fontes únicas.
- Os adaptadores mantêm o formato e o local de cada ferramenta (inclusive o `tools:` somente leitura da Governança, `SEG-06`) e apontam para a skill. Novo comando `/implementar [feature]`, com `/proxima-feature` como alias.
- Orçamentos medidos pelo check `P-24` do gate (Camada A) e por testes do scaffold: instruções raiz até 1.536 bytes, `SKILL.md` até 2.048, adaptadores de comando até 300, definições de subagente e chatmode até 420, Constitution até 9.216. Conjunto de leitura medido no scaffold: `/discovery` 17.004 bytes (limite 22 KB) e `/implementar` 5.359 bytes (limite 15 KB).
- `forge-sdd context <task>` entrega o contexto mínimo: a spec, só as linhas dos IDs que ela cita, as rules que os cobrem e as skills aplicáveis.
- Constitution: cada regra numerada vira uma linha que remete ao ID; os IDs ganham redação enxuta; o mapa de rules deixa de ser tabela e passa a apontar o frontmatter das rules (`globs`, `regras`), fonte única (`DUP-06`). Nenhum ID foi criado ou removido (45 antes e depois); `ARQ-06` agora lista todos os subcomandos existentes (`acp`, `attach`, `autopilot`, `governance`, `destroy`, `next-id`, `context`), e `FLX-02` e as Regras 13 e 15 citam a pasta da iniciativa (ADR 0002).
- Divergências de documentação corrigidas: `concepts.md` passa a apontar o dono real de diagnóstico, telemetria e lições (`cmd/forge-sdd`, dívida de `ARQ-01` registrada) e a citar `acp.HTTPHandler`; o mapa de rules deixou de omitir `sdd-acp.md` por deixar de existir.

## Medição (bytes, não tokens)

Soma dos tamanhos dos arquivos que o comando manda ler, no commit `ec9c4fe` (antes) e neste repositório (depois). Não usa `forge-sdd report`: as sessões de telemetria desta série não registraram tokens, então o relatório não permite comparar. Como 1 token vale cerca de 4 caracteres, a razão em bytes é uma aproximação razoável da razão em tokens.

| Conjunto | Antes | Depois |
|---|---|---|
| `/discovery` | 65.934 B: `CLAUDE.md` 4.548, adaptador 110, comando 3.120, papéis `pm-architect` 3.049, `planner` 2.123 e `lider` 4.020, `sdd/features/index.md` 27.150, Constitution 14.023, `clarify` 1.748, `naming-convention` 1.373, `concepts` 3.121, `lessons` 945, `progress` 604 | 21.897 B: `CLAUDE.md`, adaptador, `SKILL.md` e `references/discovery.md` da concepção, Constitution (9.210), `clarify`, `naming-convention`, `concepts`, `lessons` |
| Builder por task | 31.680 B (sem a spec): `CLAUDE.md` 4.548, adaptador 122, comando 3.853, papel 2.349, `lessons` 945, Constitution 14.023, `progress` 604, 2 rules (2.854 e 2.382) | 6.007 B de base (`CLAUDE.md`, adaptador, `SKILL.md`, `references/implementar.md`, `lessons`) mais cerca de 7 KB de `forge-sdd context` (spec da task, IDs citados, caminhos de rules e skills) |

Leitura honesta: o `/discovery` cai cerca de 67%, em grande parte porque o `index.md` de 27 KB sai do fluxo. O Builder cai de 31.680 para cerca de 13 KB (base mais `context`), ou seja cerca de 59%, comparando com o "antes" sem a spec; com a spec incluída nos dois lados a razão é parecida.

## Reescrita dos IDs da Constitution

Os 45 IDs seguem presentes e um teste (`TestConstitutionDoRepositorioMantemTodosOsIDs`) impede ganho ou perda sem ADR. O texto de todos foi condensado; mudanças de sentido registradas aqui (`GOV-01`): `SEG-01` passou a citar refs git entre as entradas externas; `SEG-05` deixou de listar variáveis de ambiente e stderr (permanecem logs, saída padrão e telemetria); `FLX-03` deixou de enumerar os campos do template de PR (`.github/pull_request_template.md` é a fonte); `GOV-03` fala em "fim de uma onda"; `GOV-04` deixou de dizer "sem jamais copiar" (o check `P-23` continua reprovando cópia literal); `FLX-06` deixou de dizer "toda task passa". As Regras 11 e 12 foram condensadas sem perder exigência.

## Consequências

- **Positivas:** o custo fixo por comando cai de dezenas de KB para poucos KB (medição abaixo); uma instrução tem um único lugar; o orçamento é verificado pelo portão e não depende de disciplina.
- **Negativas / Trade-offs:** os adaptadores de comando dependem de a ferramenta seguir o ponteiro para a skill; o `update` só remove os artefatos antigos idênticos à versão imediatamente anterior; os comandos ainda não executam o orçamento em tempo de execução, apenas o medem; os arquivos `.agents/skills/sdd-*` são gerenciados pelo forge-sdd e sobrescritos a cada `update`, então personalizações devem virar skills próprias (fora de `sdd-*`) ou rules.
- **Edição de regra (`GOV-01`):** o texto de `DUP-10` foi alterado nesta série (agora cita `.agents/skills/` e `.agents/rules/`) e está coberto por este ADR; o bloco gerenciado `forge-sdd-governanca` mantém `versao=3.0.0-beta` por não haver mecanismo de bump automático.
- **Ações de acompanhamento:** extrair doctor, telemetria e lições de `cmd/` para `internal/` (dívida de `ARQ-01`); decidir o destino de `acp.HTTPHandler` (ADR 0001); avaliar o agrupamento de auto-skills quando houver mais de uma.

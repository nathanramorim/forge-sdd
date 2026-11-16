# Discovery 88 — Modo Pipeline Avançado (experimental)

## Problema
**Segundo problema (mesma iniciativa):** `pm_mcp.require_existing_card` nasce `true` por padrão (`internal/config/forge.go`), então toda esteira passa a exigir card/ID de tarefa no gerenciador de projeto sem nunca perguntar. Quem não usa ClickUp, Jira etc. precisa descobrir o campo e editar o JSON.

No pipeline multi-sessão a esteira para a cada decisão humana: Gate 1 da spec, "Abrir mesmo assim?", escalação após `max_rework_attempts`, confirmações de ferramenta e merge. Quem quer deixar a esteira rodando por dias precisa ficar respondendo.

## Para quem
Pessoa dona do projeto, com harness bem configurado (Constitution, rules e ADRs maduras), em ambiente isolado.

## Jornada
1. No menu, em "Iniciar esteira (up)", pressiona **→**: liga o modo avançado (sem rótulo no menu; só a linha de confirmação mostra `avançado`). **←** volta ao normal.
2. Confirma que o ambiente é isolado (o modo implica `permissions: full`).
3. Discovery gera a spec. O **Reviewer** valida a spec contra Constitution, rules, ADRs e arquitetura (Portão de Spec).
4. Aprovada, a fila é montada e o Builder constrói; o Reviewer confere a entrega contra a spec (Portão de Entrega).
5. Violação volta ao Builder; no limite de tentativas a task vira `blocked` e a fila segue nas independentes.
6. A pessoa volta dias depois e lê o painel, o inbox e os PRs abertos.

## Regras de negócio
- **RN-1** Modo `normal` é o padrão e não muda de comportamento. O avançado é experimental; se der certo, vira padrão para todos (decisão futura, outra iniciativa).
- **RN-2** No avançado, o Reviewer substitui a aprovação humana do Gate 1 (FLX-01). Exige a ADR-0013.
- **RN-3** O Reviewer nunca é autor da spec; roda em sessão própria, contexto limpo, somente leitura (SEG-06). O parecer é gravado pelo Go em `review-spec.md`.
- **RN-4** Parecer da spec: `aprovado` | `ajustar` (volta à Discovery com evidência, limite de tentativas) | `escalar` (vira `blocked`, nunca espera em silêncio).
- **RN-5** Nunca commita em `main`; o merge do PR fica manual (padrão adotado).
- **RN-6** Salvaguardas obrigatórias: teto de tokens e de tempo por job, teto de jobs por dia, parada limpa por `forge down` ou arquivo `.forge/runtime/STOP`.
- **RN-8** O card de gerenciador de projeto **não é exigido por padrão**: a pergunta "Você usa gerenciador de projeto?" é feita no `forge pipeline init` (e no menu quando o pipeline é ativado) e a resposta fica em `forge.json`. Sem terminal interativo, vale `false`.
- **RN-9** Existe um menu **Settings** (nome do item no menu principal) com chaves liga/desliga (`ativado`/`desativado`). Por enquanto contém só a chave de card obrigatório; novas chaves entram aqui, não em prompts soltos. O item de card sai de "Ativar ou reconfigurar o pipeline" (DUP-01).
- **RN-7** Exceção a spec/ADR/Constitution nunca é decidida pela máquina: vira `blocked` + inbox (GOV-02).

## Escopo
**Dentro:** Settings com a chave de card; default `false` + pergunta no init; campo `autonomy` em `forge.json`; gesto →/← no `up`; job `spec_review`; parecer do Reviewer; `blocked` com continuação; salvaguardas; ADR-0013; docs.
**Fora:** outras chaves no Settings (entram quando fizerem sentido); migrar `forge.json` existente (quem já tem o valor gravado mantém); auto-merge; tornar o avançado padrão; mudar o modo normal; novos agentes/CLIs.

## Riscos
| Risco | Mitigação |
|---|---|
| Reviewer aprova spec ruim e a máquina constrói por dias | Checklist fixa (Constitution, rules, ADRs, `O que já existe`, critérios executáveis) + Camada A (`forge gate`) obrigatória antes do parecer |
| `permissions: full` em máquina errada | Confirmação de ambiente isolado + `forge doctor` avisa |
| Custo/tokens descontrolados | RN-6 + auto-reset de contexto (ADR-0009) |
| Mesmo modelo escreve e revisa | Sessão e contexto separados; recomenda-se agente diferente em `sessions` (aviso, não bloqueio) |
| Gesto →/← não suportado pelo `huh` | Task 88-02 valida; fallback documentado: `forge pipeline autonomy advanced|normal` |

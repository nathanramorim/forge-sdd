# fix-106-07 — Toda decisão é componente de ask, não texto do chat

**Branch:** `docs/106-specs-worktree-modo-ask` (PR único das três specs 05, 06 e 07)
**Origem:** Iniciativa 106, `feat-106-26a` (comandos de concepção) e `sdd/memory/clarify.md`
**Tamanho:** `M`
**Depende de:** `fix-106-04` (reescreve a seção de visualização e o Gate 1 do `clarify.md`)
**Status:** `spec` (aguarda Gate 1)

## Sintoma

O dono pediu: "sempre que houver interação não deixar a decisão no texto do chat, sempre abrir um componente de ask". Nesta sessão, em 22/11/2026, o agente deixou o Gate 1 e várias perguntas em prosa ("Gate 1: aprova a spec?", "Quer que eu faça essa mudança?") até o dono corrigir. A regra existe em pedaços, mas nada a garante.

## Apuração

| Pergunta | Evidência | Conclusão |
|---|---|---|
| A regra existe? | `sdd/memory/clarify.md:23`: "use `AskUserQuestion` para toda pergunta de múltipla escolha ou sim/não" | **Sim**, mas na seção "Como perguntar" do **clarify**, que fala de esclarecer a descrição |
| Cobre o Gate 1? | `.agents/skills/sdd-concepcao/references/discovery.md`: "**Gate 1 (`FLX-01`):** submeta ao humano" | **Não**: texto, sem menu |
| As references seguem? | Das **12** references com palavras de decisão (`pergunt`, `confirm`, `aprova`, `Gate 1`), só **5** citam `AskUserQuestion` ou o `clarify.md`; **7 não citam**: `constitution`, `implementar`, `pr-review`, `archive`, `toca-ai`, `upgrade-sdd`, `lider` | **Cobertura 5/12** |
| Há teste? | Nenhum teste varre as references por esse contrato | **Sem garantia** |

## Feature Correlacionada

- **Arquivo de origem:** `sdd/memory/clarify.md` (seção "Como perguntar") e as 7 references sem o ponteiro.
- **Causa-raiz:** a regra foi escrita como **instrução de um passo** (clarify), não como **contrato geral** que toda reference herda. Quem escreve uma reference nova não é obrigado a citar nada.
- **Evidência do defeito:** o levantamento acima e as perguntas em prosa desta sessão.

## Modelo do módulo

**Enxuto** (`ARQ-11`): texto em um arquivo, ponteiros nas references, uma linha nos templates dos agentes e um teste de varredura. Sem pacote novo.

## Decisões do projeto aplicadas

- `DUP-01` — o contrato mora em **um** lugar (`clarify.md`); as references **apontam**, não repetem.
- `GOV-04` — a explicação fica em `references/`/`clarify.md`, a Constituição não muda.
- RN-9 da discovery 104 — todo handoff é menu; esta spec generaliza para **toda decisão**.
- `FLX-01` — o Gate 1 é decisão do humano; o menu é a forma de ele decidir.
- **Não usar hook de `Stop` para impor a regra.** O dono viveu isso com o `/goal` (o hook concorre com o agente e força o fim depois de 9 bloqueios). A garantia é **texto único + varredura em teste**, não bloqueio em tempo de execução.
- `COD-05` — pt-BR, nível padrão.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Regra do ask | `sdd/memory/clarify.md` ("Como perguntar") | uma frase, escopo de esclarecimento | **promover** a contrato geral, com seção própria |
| Reescrita do Gate e da visualização | `fix-106-04` | reescreve a seção "Visualização pós-spec" do mesmo arquivo | **depender**: esta spec entra depois, sem editar as mesmas linhas |
| Referências de comando | `.agents/skills/*/references/*.md` | 12 com decisão, 7 sem ponteiro | **acrescentar** o ponteiro nas 7 (`DUP-01`) |
| Instrução nos agentes | `CLAUDE.md`, `GEMINI.md`, `AGENTS.md`, `copilot-instructions.md` (+ templates) | sem a regra | **uma linha** apontando para o contrato |
| Varredura determinística | `internal/gate` (testes que leem as skills) | há precedente (`P-23`, `P-24`) | **criar** teste de varredura das references |
| Hook de bloqueio | `Stop` | concorre com o agente | **não usar** |

## Dependências novas (`SEG-07`)

`Nenhuma`.

## Rules e IDs da Constituição Aplicáveis

`COD-04`, `COD-05`, `DUP-01`, `FLX-01`, `GOV-04`, `TST-01`, `TST-03`.

Rules: `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`.

## Escopo

**Dentro:**
1. **Contrato único** em `clarify.md`, seção "Toda decisão é componente de ask":
   - Vale para: sim/não, escolha entre opções, confirmação de ação externa (push, merge, fechar PR, remover), Gate 1, esclarecimento.
   - O texto do chat **informa** (resultado, evidência, resumo); **nunca pergunta**.
   - Recomendada primeiro; no máximo 4 opções por pergunta e 4 perguntas por chamada.
   - Em agentes sem o componente (Gemini, Copilot), lista numerada; no CLI do forge, `Select`/`Confirm` do `huh`.
   - Depois de qualquer ação de visualização, **voltar no mesmo turno** com o menu da decisão (liga com o `fix-106-04`).
2. As 7 references sem ponteiro passam a citar o contrato.
3. Uma linha nos templates e arquivos dos agentes: "Toda decisão abre componente de ask; ver `sdd/memory/clarify.md`".
4. Teste de varredura: toda reference com palavras de decisão cita `clarify.md` ou `AskUserQuestion`.

**Fora:**
- Bloquear o agente em tempo de execução (hook de `Stop`).
- Mudar o comportamento dos agentes além do texto do contrato.
- A memória pessoal do dono (fica fora do repositório).

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# 1. O contrato existe e é a fonte única.
grep -q 'Toda decisão é componente de ask' sdd/memory/clarify.md
# 2. Toda reference de decisão cita o contrato (hoje 5 de 12; deve ser 12 de 12).
go test ./internal/gate/... -run 'TestReferenciasDeDecisaoCitamOContrato'
# 3. Os arquivos dos agentes carregam a linha.
test "$(grep -l 'componente de ask' CLAUDE.md GEMINI.md AGENTS.md .github/copilot-instructions.md | wc -l | tr -d ' ')" = 4
# 4. Goldens em dia.
go test ./internal/scaffold/...
# 5. Nenhum hook de Stop foi criado para impor a regra.
! grep -rn '"Stop"' internal/scaffold/hooks.go || exit 1
go vet ./... && go test ./...
```

- **Teste negativo obrigatório (`TST-03`):** uma reference de teste com a palavra "aprova" e **sem** o ponteiro faz `TestReferenciasDeDecisaoCitamOContrato` **falhar**; com o ponteiro, passa.
- **Prova do estado atual:** o teste, rodado antes da correção, lista exatamente as 7 references sem ponteiro.

## Riscos

- **R1 — Conflito de edição com o `fix-106-04`** no mesmo arquivo. Mitigação: ordem fixa (04 antes de 07) e seções diferentes.
- **R2 — A varredura por palavra dá falso positivo** (uma reference que só cita "confirmar" no sentido de teste). Mitigação: a lista de palavras fica em um lugar, e uma exceção explícita exige comentário.
- **R3 — O contrato vira texto morto de novo.** Mitigação: o teste falha quando uma reference nova de decisão não o cita.

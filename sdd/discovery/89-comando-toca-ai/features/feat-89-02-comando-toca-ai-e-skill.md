# feat-89-02 — Comando `/toca-ai`: o Líder planeja e gera todas as specs, o motor executa

**Iniciativa:** 89-comando-toca-ai
**Branch:** feat/89-comando-toca-ai
**Status:** ready
**Depende de:** feat-89-01

## Problema
A pessoa colava um prompt enorme ("Goal set: você é responsável por fazer acontecer…") para deixar um trabalho longo rodando. Quer um comando pronto: informa só o objetivo e continua conversando no chat (`/rc`).

## Comportamento (`/toca-ai <objetivo>`)
1. **Specs primeiro, execução depois.** O Líder gera **todas** as specs antes de executar: `plan.md` (etapas e `**Depende de:**`), uma spec por etapa (objetivo, entradas, saídas, critério de aceitação, o que ler) e `status.md` até 1 KB. Nada fica só na conversa: qualquer etapa é retomável lendo o próprio arquivo.
2. **Pré-voo com a pessoa, uma vez só.** Ao gerar as specs, o Líder grava `requisitos.json` com tudo que as etapas vão precisar: comandos exatos, ferramentas, arquivos de entrada e consentimentos. Roda `forge pipeline toca-ai --preflight` e mostra **uma única lista** do que falta (ex.: "liberar este comando", "confirmar que a voz é sua"). Só depois que a pessoa responde essa lista o trabalho começa; dali em diante **não há mais pergunta de rotina**.
3. **Regra de ouro: o agente nunca manda a pessoa rodar algo.** Todo comando é executado pelo próprio agente ou subagente. "Rode no seu terminal" e "libere a permissão" são proibidos no meio da execução: ou entraram no pré-voo, ou viram etapa `blocked` com o item exato, enquanto as etapas independentes seguem.
4. **Entrega ao motor e para.** Roda `forge pipeline toca-ai <iniciativa>` (feat-89-01) e informa "entregue à esteira". O Líder não executa as etapas nem carrega o trabalho longo; o chat fica livre.
5. **Impasses:** decide sozinho na ordem Constitution, ADRs e rules, depois a opção mais conservadora e reversível, e registra em `decisoes.md` (o que, por quê, alternativa descartada). Só chama a pessoa para: ação irreversível fora do repositório, segredo, gasto externo, merge em `main` (Regra 1) e exceção a regra (`GOV-02`).
6. **Contexto, `/clear`, `/goal clear` e archive:** quem faz é o motor (feat-89-01). O Líder **não** depende do `/goal`; se existir uma meta do Claude Code ativa, ao terminar avisa "pode digitar `/goal clear`", pois comando com barra só a pessoa digita.
7. **Acompanhar e retomar:** `/status` e `forge status` mostram o andamento; rodar `/toca-ai` de novo no mesmo objetivo continua de onde parou.
8. **Linguagem:** segue o `Nível de Linguagem` da Constitution (ex.: `iniciante` = palavras simples, sem termos em inglês); o comando não embute vocabulário.
9. **Fim:** relatório final (feito, decisões tomadas, pendências humanas) e telemetria por etapa.

## Premissas (a confirmar no Gate 1)
- Nome do comando: `/toca-ai`, sem apelido.
- Sem o pipeline ativo, o `/toca-ai` orienta `forge pipeline init` em vez de executar no chat.

## Modelo do módulo
**Enxuto** (ARQ-11): só instrução de agente (skill, comandos e cheatsheet).

## Decisões do projeto aplicadas
ADR-0003 (skills por estação e orçamento de contexto), ADR-0004, ADR-0013. Rules: `sdd-fluxo-e-portao`, `sdd-arquitetura`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Padrão de comando novo | `inbox`: `.agents/skills/sdd-operacao/references/inbox.md`, `.claude/commands/inbox.md`, `.gemini/prompts/inbox.prompt.md`, `.github/prompts/inbox.prompt.md`, templates, golden e `internal/scaffold/cheatsheet.go` | Referência + 3 adaptadores + templates + golden | **reusar o padrão** | DUP-01: mesma forma de todo comando |
| Papel do Líder | `sdd-operacao/references/lider.md`, `.claude/agents/lider.md` | Painel, despacho e fechamento | **estender** com `references/toca-ai.md` | Uma casa para a orquestração |
| Motor | feat-89-01 | `forge pipeline toca-ai` | **reusar** | O Líder só chama |
| Registro de comandos | `sdd-operacao/SKILL.md`, `CLAUDE.md`, `GEMINI.md`, `AGENTS.md`, `.forge/LEIA-ME.md` | Tabelas e listas | **estender** | Fonte única |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`FLX-01`, `FLX-06`, `ARQ-01`, `DUP-01`, `COD-01`, `COD-05`, `TST-01`.
Rules: `.agents/rules/sdd-fluxo-e-portao.md`, `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`.

## Escopo
1. `references/toca-ai.md` (skill `sdd-operacao`) + linha na tabela do `SKILL.md`: specs primeiro, entrega ao motor, ordem de decisão, `decisoes.md`, limites de interrupção, `/goal clear`, relatório final.
2. Adaptadores para Claude, Gemini e Copilot, templates em `internal/scaffold/templates/` e golden.
3. Registro em `cheatsheet.go` e nas listas de comandos.
4. `docs/pipeline.md`: seção curta (como usar, quando você é interrompido, como retomar).
5. Respeitar o orçamento de contexto (P-24, ADR-0003): referência curta.

## Fora do escopo
Subcomando de nível raiz; apelido de comando; mudar o modo avançado.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/scaffold -run 'TestCheatsheet|TestGolden|TestCommands'
go vet ./... && go test ./...
forge gate
```
- `forge update` instala `/toca-ai` nos três agentes e o golden bate.
- Se `requisitos.json` listar comando, ferramenta, arquivo ou consentimento que falta, o `/toca-ai` mostra tudo numa lista e **não começa** até a pessoa responder.
- A referência traz a regra "nunca mande a pessoa rodar" e o exemplo do `voice-studio` (consentimento de voz) como caso de pré-voo.
- A referência descreve: specs de todas as etapas antes de executar, entrega ao motor, impasses, limites de interrupção, `/goal clear` e retomada (com exemplo).
- O comando não embute vocabulário de linguagem.
- O orçamento de instruções (P-24) continua dentro do limite.

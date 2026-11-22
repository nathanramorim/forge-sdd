# Discovery — 105 Experiência do CLI e proficiência do harness

## Pedido do dono (literal)
"considere melhorar a experiência do CLI do forge SDD, e forma de navegar ser mais fluida, mais com a temática do design system, considere também avaliar a proficiência do harness, proponha, reveja o que está obsoleto, a ideia principal aqui é facilitar a vida de quem usa esse boilerplate sdd".

## Decisões do dono (não reabertas)
1. Escopo total em 3 ondas, por valor para quem usa o boilerplate: **1** obsoletos e limpeza; **2** navegação e tema do CLI; **3** medição de proficiência do harness.
2. Tema: paleta **permanente** do Forge (`#0d0f14` fundo, `#4f8ef7` azul, `#3ecf8e` verde, `#f5b04a` âmbar, `#f2677a` coral, `#e9ecf4` texto). Sem Edição Missões, sem personagens com arma (a edição termina em 23/11/2026 00:00).
3. Produto = 3 sessões (Discovery, Builder, Reviewer) com aprovação humana. As 5 estações são obsoletas.

## Clarify (`sdd/memory/clarify.md`)
Sem pergunta nova: critério (ondas e métricas), escopo (3 ondas) e dependência (104) estão claros. Duas lacunas viraram `A confirmar` no `status.md`: base da 105 versus merge da 104 (ver Riscos) e o que remover de fato.

## Base de evidência e limites
- Binário usado: `go build ./cmd/forge-sdd` da branch (`/tmp/forge105`, versão `5.0.0`), **não** o do PATH. Para o que só existe na 104 (`doctor --harness`, `eval`, `gate --criterios`) usei um binário compilado de `bbd26d6` (branch `docs/constitution-alinhamento-104`, que contém a 104 mesclada nas ondas 1 a 3 e as ADRs 0016 a 0020).
- **Fato que muda o plano:** a `main` (fab5af4) só tem a *discovery* da 104. Não tem `internal/harness`, `internal/doctor`, `doctor --harness`, `eval`, `mcp`, ADR-0016 a 0020, `obsoletos.md` nem `constitution-propostas.md`; tudo isso está em `bbd26d6`. A onda 3 depende de a 104 chegar à `main`.
- Nada abaixo foi executado de ponta a ponta com agentes reais; marco `não verificado` onde cabe.

## Como o CLI navega hoje (medido)
`forge` sem argumentos com TTY abre o menu `huh` (`cmd/forge-sdd/menu.go`, ADR-0005); sem TTY imprime ajuda curta e sai 0 (`menu.go:52`, medido com `< /dev/null`). 30 entradas no mesmo nível (inclui `help` e `completion`) (`forge --help`).

| # | Atrito | Evidência | Quem sente |
|---|---|---|---|
| A1 | Subcomando mal digitado **sai com 0** e só imprime ajuda: `forge bus foo` rc=0, `forge queue claim` rc=0 (provável confusão entre `queue` e `bus claim`). `forge foo` sai 1, mas imprime o erro duas vezes ("Error: unknown command..." e a mesma frase de novo). | comandos rodados, rc sem pipe | scripts e quem copia comando de doc |
| A2 | `forge --help` é lista plana de 30 entradas em ordem alfabética, em inglês parcial ("Available Commands", "help for"), sem separar o caminho feliz (`init`, `pipeline init`, `up`, `status`) do interno (`bus`, `worker`, `acp`, `queue`, `governance`, `skills`). A frase de `Long` diz "pronta para uso com GitHub Copilot" (`main.go:22-23`). | `forge --help`; `main.go` | usuário novo |
| A3 | `forge --version` não existe (`unknown flag: --version`); só `forge version`. `init --version` existe com outro sentido (versão do Forge-SDD a instalar). | rodado | todo mundo |
| A4 | `forge doctor` diz "Versão registrada em sincronia: v4.5.0-beta" com o binário em 5.0.0: compara `.sddrc` com `.sdd-version`, não com o binário (`doctor.go:90-96`). Imprime os 126 caminhos de features sem pedir. O `doctor --harness` da 104 acusa o descompasso ("projeto 4.5.0-beta mais antigo que o binário 5.0.0"), mas é outra flag. | `forge doctor`; `bbd26d6 doctor --harness --json` | quem atualizou o binário |
| A5 | `forge status` mostra "Iniciativas ativas: 7", incluindo 79 a 82 com fila `completed` e "próximo: Revisão humana ... PR #74" antigo: não distingue concluída de ativa. | `forge status` | quem volta ao projeto |
| A6 | Caminho do `init` ao primeiro PR tem 6 passos em 3 lugares, sem guia único: `forge init` (11 campos em `internal/survey/survey.go`, ou `--yes`), `/constitution` no agente, `forge pipeline init`, `forge up --discovery X --builder Y --reviewer Z`, `/discovery` + aprovação (Gate 1), `/split-features` e a entrega ao Builder, PR e aprovação final. `docs/pipeline.md` descreve "3 passos" mas o cabeçalho diz v4.5.0-beta. `init --agent` tem padrão `copilot`. | `forge init --help`; `docs/pipeline.md:1,21-40` | usuário novo |
| A7 | Sem cor nem tema próprio: nenhum `lipgloss` direto, nenhuma leitura de `NO_COLOR`; só ANSI cru de tela alternativa (`queue.go:276`, `status_live.go:411`). Os menus `huh` usam o tema padrão da biblioteca, não a paleta do Forge. | busca no código | quem quer identidade visual e acessibilidade |
| A8 | Legado exposto no `--help`: `queue stations` ("5 filas de estações"), `attach` ("5 estações canônicas", `attach.go:21`), `queue handoff`. | `forge queue --help`; `forge queue stations` | usuário novo |

Não foi medido: contagem de toques por tarefa no menu; largura estreita (as telas não foram renderizadas em 40 colunas); comportamento em Windows.

## Problema
Quem adota o boilerplate encontra um CLI com 30 entradas na mesma lista, erros que passam como sucesso, mensagens de um produto de 5 estações, e nenhum número que diga se o harness (contratos, rules, gate, esteira) está funcionando. O custo é perder tempo antes do primeiro PR e não saber se o setup está saudável.

## Para quem
Quem usa o boilerplate: (a) pessoa nova, que roda `forge init` e precisa chegar ao primeiro PR sem ler 109 linhas de `docs/pipeline.md`; (b) pessoa recorrente, que volta ao projeto e quer saber "o que faço agora"; (c) mantenedor, que precisa de um número para dizer "o harness está proficiente".

## Jornada alvo
`forge` abre menu com a opção recomendada marcada ("próximo passo") conforme o estado do projeto (sem `sdd/` → `init`; constituição rascunho → `/constitution`; sem `forge.json` → `pipeline init`; spec aprovada e fila → `up`; PR pronto → aprovar). Cada subcomando termina oferecendo o próximo menu (RN-9). Comandos de baixo nível ficam sob "Avançado".

## Regras de negócio
- **RN-1** Subcomando desconhecido ou grupo sem ação (`bus`, `queue`, `pipeline`) sai com código diferente de zero e uma única mensagem, com sugestão (`Você quis dizer ...`), em pt-BR.
- **RN-2** Cor só com TTY e sem `NO_COLOR`; sem TTY a saída é texto puro e estável (quem lê é agente ou script). Largura estreita (40 colunas) quebra linha, não corta.
- **RN-3** Mensagem segue a paleta: azul destaque, verde ok, âmbar aviso, coral erro, texto padrão; ícone ou palavra acompanha a cor (não depender só de cor).
- **RN-4** Nenhuma remoção sem decisão do dono e ADR quando mexer em `ARQ-06`; nesta iniciativa o máximo da onda 1 é **avisar** (depreciação) e corrigir texto.
- **RN-5** Métrica sem dado é "a medir", nunca 0 nem estimativa disfarçada (`medicao` real versus `estimado` precisa aparecer).
- **RN-6** Dado de arquivo, processo, ambiente e git é validado na borda; saída é saneada (`notify.Sanitize`); nunca executar caminho vindo de dado externo; escrita por `writeFileSafe`; testes não dependem de identidade do git; `--dry-run` sem efeito colateral (`ARQ-05`).
- **RN-7** Todo handoff à pessoa é menu (RN-9 da 104).

## Escopo
**Dentro, onda 1 (obsoletos e limpeza):** texto e ajuda do raiz; `--version`; `doctor` compara com o binário; `status` esconde concluída; depreciar `queue stations`/`attach` (só aviso); trocar "5 estações" em templates e goldens; renomear o papel `lider` no launcher; ADR de obsolescência. Detalhe em `obsoletos.md`.
**Dentro, onda 2 (navegação e tema):** `internal/ui` (paleta, `NO_COLOR`, TTY, largura); tema `huh`; erro único e sai !=0; `--help` agrupado em pt-BR; menu raiz com próximo passo recomendado; `doctor` e `status` temáticos e curtos; fim de `init` com menu do próximo passo.
**Dentro, onda 3 (proficiência):** painel de métricas, proposta e limiares em `harness.md`; `forge report --proficiencia` (somente leitura).
**Fora:** remover comandos do `ARQ-06`; mudar o laço do pipeline ou o modo `chat` (104); tema Missões; TUI nova com `bubbletea` (só `huh`/`lipgloss`); medir qualidade de modelos; ligar telemetria a serviço externo.

## Riscos
- **R1 Base:** a 105 nasce na `main`, mas a onda 3 e parte da onda 1 (`doctor`) dependem de código só em `bbd26d6`. Mitigação: ondas 1 e 2 independentes da 104; onda 3 só começa depois do merge da 104 (`A confirmar`).
- **R2 Tema e acessibilidade:** `huh` v0.3.0 pode não expor tema ou não respeitar `NO_COLOR`; verificar no context7 antes da task 2.1. Fallback: `internal/ui` decide.
- **R3 Goldens:** trocar "5 estações" em `CLAUDE.md.tmpl`, `GEMINI.md.tmpl`, `lider.md.tmpl` e goldens muda arquivos que o usuário recebe no `init`/`update` (`TST-02`).
- **R4 `lider`:** está fixo como agente da sessão Discovery em `launcher.go:125`, `tmux.go:133`, `preflight.go:165`; remover o arquivo quebra `forge up`.
- **R5 Telemetria:** 88 de 97 arquivos de `sdd/.metrics/session-*.json` têm `tokens_input: 0`; tokens por task não é confiável hoje.
- **R6 ADRs `proposto`:** 0009 a 0014 estão em uso mas sem aprovação formal.
- **R7 Numeração:** nova ADR desta iniciativa deve ser a próxima livre depois da 0020; se a 104 não mesclar antes, haverá disputa de número.

## Gate 1 (FLX-01)
Aguardando aprovação humana. Decisões abertas em `status.md` → `A confirmar`.

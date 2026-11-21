# Plano — Iniciativa 107

Ordem: de baixo para cima (dado e contrato, depois comportamento, depois instrução e documentação). Até 7 tasks por onda. Alinhado à 106 redesenhada: o split é do Builder, uma branch e um PR em draft por feature, estado por feature (o que a 106 ainda não implementou está marcado "a verificar"). Os arquivos de feature nascem no `/split-features` da onda 1. Sem dependência de 104-08/09 (decisão do dono).

## Onda 1 — Tabela, estado, WIP humano e instrução (01 a 04 independem da 106; a 12 depende da 106/0.5 e do PR único de Constituição)
- **feat-107-01** Tabela de roteamento como dado validável (`internal/pipeline/roteamento.go`): intenção, papel, modo, isolamento, portão; fluxo da 106 (spec, Conferência 1, [Seguir], split pelo Builder, feature por branch, Conferência 2, [Continuar] por onda, arquivar da principal); papel conferido contra `.claude/agents/`. **Depende de:** nenhuma. Rules: ARQ-01, ARQ-02, COD-01, DUP-01, TST-03.
- **feat-107-02** Estado de delegações de runtime (`.forge/runtime/orquestrador/<iniciativa>.json`): leitura estrita, gravação atômica, lock da iniciativa, contador de 3 tentativas por delegação (reusa `max_rework_attempts`); sem arquivo compartilhado no git. **Depende de:** feat-107-01. Rules: SEG-01, SEG-02, SEG-05, COD-03, DUP-06.
- **feat-107-03** Política de entrada não confiável (relatório e mensagem de sessão par como dado; saneamento; permissão negada não contornada) com testes de injeção. **Depende de:** feat-107-02. Rules: SEG-06, SEG-08, FLX-08.
- **feat-107-04** Reference do papel Orquestrador em `sdd-operacao` (tabela, briefing, retomar, menu, "abrir pasta"), com templates e goldens. **Depende de:** feat-107-01. Rules: DUP-10, TST-02, COD-05.
- **feat-107-12** WIP humano, **só código** (a 107 não edita a Constituição): leitura do campo `WIP humano` pelo leitor compartilhado da tabela (da 106/0.5, 106 primeiro), validação (inteiro positivo ou `desligado`; recusa valor inválido), limite 3 por padrão, parada no limite com pergunta em menu, aviso de riscos antes de reduzir ou desligar, modo no `status.md`; distinto de `builder.max_parallel_tasks`. Contagem de pendências por interface `FonteDePendencias` (**decisão do dono em 2026-11-21**): `Pendentes(ctx) (int, error)`; padrão baseado em arquivo (itens `review_pending` e specs aguardando [Seguir]); a 108 (task 3.4) acrescenta depois uma segunda fonte, só com o espelho do quadro ligado; a 107 não depende da 108. **Depende de:** feat-107-02, do PR único de Constituição (task nova da 106, para o campo existir) e do leitor da tabela da 106/0.5. Rules: SEG-01, SEG-02, FLX-05, DUP-06, TST-03.

Paralelo: 01 sozinha; 02 e 04 após 01; 03 após 02; 12 após 02, 106/0.5 e o PR único de Constituição.

## Onda 2 — Verificação antes do handoff (só depois de a 106 fixar o script, RN-12)
- **feat-107-05** `forge pipeline verificar --ponto spec|entrega --feature <sid> --json`, sobre o contrato de checagens do script da 106: gate, commit, branch, diff no escopo da feature, `status.md` até 1 KB, imagens dos diagramas, testes reexecutados. **Depende de:** feat-107-02 e do script da 106. Rules: ARQ-01, ARQ-05, ARQ-06, SEG-03, FLX-06.
- **feat-107-06** Consumo do adaptador de VCS da 106 (`github`, `azure-devops`, `nenhum`) para o estado do PR por feature; sem adaptador próprio. **Depende de:** feat-107-05 e do adaptador da 106 (a verificar). Rules: SEG-03, SEG-01, DUP-06.
- **feat-107-07** Retomada: reconstruir delegações abertas de runtime e o estado por feature da 106 depois de `/compact` ou reinício. **Depende de:** feat-107-02 e da decisão xvi da 106 (a verificar). Rules: FLX-02, DUP-01.

## Onda 3 — Tetos, ferramentas e avaliação
- **feat-107-09** Teto técnico de paralelismo (`max_parallel_tasks`) e de tentativas (3 por delegação, `max_rework_attempts`); escalada em menu no teto. **Depende de:** feat-107-02. Rules: DUP-06, TST-03.
- **feat-107-08** Fallback para o modo `sessions` com aviso onde falta equivalente de subagente, com a matriz mínima da própria 107 (sem `Capabilities` da 104-09). **Depende de:** feat-107-04. Rules: DUP-03, ARQ-02.
- **feat-107-10** Casos de avaliação do orquestrador (injeção, loop, permissão negada, WIP no limite) no `forge eval` da 104-11. **Depende de:** feat-107-03, feat-107-09, feat-107-12 e feat-104-11 (já mesclada: a verificar). Rules: TST-01, TST-04.
- **feat-107-11** `sdd/spec/concepts.md`, `docs/pipeline.md`, `flows.md` e o texto final da ADR (número no merge). **Depende de:** ondas anteriores. Rules: FLX-05, GOV-01, DUP-01.

## Onda 4 — Início automático do Builder (decisão do dono em chat, 20/11/2026; aprovada em 2026-11-21)
- **feat-107-13** Início automático do Builder local, **no core** (veio da 108/3.7, sem depender do quadro nem do módulo da 109). Comando sob `forge pipeline observar` (`ARQ-06`; **não** `forge project`, que é do módulo). Processo local do forge **só enquanto a sessão estiver aberta**: polling com intervalo mínimo de 30 s (padrão 60 s; a confirmar contra o limite de taxa do VCS, a verificar); só lê o estado por feature (106/0.7), o marcador `Gate 1: aprovado` do `status.md` (106/2.2) e o merge da spec **pelo adaptador de VCS da 106** (nunca `gh` direto, nunca texto de cartão, título ou comentário). Regra Todo para Building: spec na principal, **WIP humano 3** (feat-107-12), **teto de 3 tentativas** (reusa o contador `Attempt` da 106/1.6; não cria outro, `DUP-06`), modo e atestado da 106; não grava nada que seja aprovação sem menu; encerra limpo em Ctrl-C ou fim da sessão; `--dry-run` não escreve nem inicia nada. Só roda com o campo `Início automático do Builder` = `local` (valores `desligado`, padrão, ou `local`; a linha `| Início automático do Builder | desligado |` já está na proposta da 106/0.8, **a 107 não edita a Constituição**); `desligado` recusa com mensagem e código diferente de zero. Funciona com `Módulos permitidos = nenhum` e com VCS `nenhum` (sem chamar `gh`). **Depende de:** feat-107-12, 106/0.5, 106/0.8, 106/0.7, 106/2.4; transitivas pela 2.4: 106/0.2, 2.1, 2.2, 1.6 (a 108/3.7 as listava como duras). Rules: FLX-01, SEG-03, SEG-08, DUP-06, TST-03.

Paralelo: nenhum (fecha a cadeia da 106/2.4). A 108/3.7 não será mais construída no core: a regra mora aqui e o módulo do quadro só soma a fonte de pendências (109, capacidade `pendentes`).

Colisões a sequenciar: templates e goldens com a 105 e a 106 (`TST-02`); número de ADR no merge; a 107 não edita a Constituição (o campo `WIP humano` vai no PR único da 106, notas em `.agents/rules/sdd-modos-de-projeto.md`).

## Fecho da 108 (prioridade 0)
Decisão do dono (20/11/2026): a 108 (quadro opcional do GitHub Project) é prioridade 0 e todas as suas dependências são construídas. Da 107, a 108 exige só 107-01, 107-02 e 107-12.
- **Onda 1 do fecho: 107-01** (tabela de roteamento). Depende de nada.
- **Onda 2 do fecho: 107-02** (estado em `.forge/runtime/orquestrador/`). Depende só da 107-01; pelo plano acima, não usa o estado por feature da 106/0.7 (só a 107-07 usa).
- **Onda 3 do fecho: 107-12** (WIP humano, só código). Depois de 106/0.5 (leitor da tabela) e do PR único de Constituição da 106.
- 107-01 e 107-02 **não dependem de nada da 106** (confirmado pelas linhas `Depende de` do plano).
- **Adiadas: 107-03 a 107-11.** A 108 não as exige; entram na ordem normal da 107, sem bloquear o fecho. Cada onda do fecho tem 1 task (limite de 7).

## Ordem das ondas do core (2026-11-20)
A ordem de merge do core (105, 106, 107, 109 e fixes) foi refeita na iniciativa 109, seção `## Ordem das ondas do core (substitui o fecho da prioridade 0)` de `sdd/discovery/109-modulos-externos/plan.md`. O fecho da 108 acima deixa de ser o fecho obrigatório; o escopo da 107 não foi cortado e a 107-13 entra pela regra do início automático.

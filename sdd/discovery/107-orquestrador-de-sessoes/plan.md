# Plano — Iniciativa 107

Ordem: de baixo para cima (dado e contrato, depois comportamento, depois instrução e documentação). Até 7 tasks por onda. Nada é implementado antes do Gate 1 (`FLX-01`). Os arquivos de feature em `features/` nascem no `/split-features`.

## Onda 1 — Tabela, estado e instrução (independe da 106)
- **feat-107-01** Tabela de roteamento como dado validável (`internal/pipeline/roteamento.go`): intenção, papel, modo, isolamento, portão; papel conferido contra `.claude/agents/`. **Depende de:** nenhuma. Rules: ARQ-01, ARQ-02, COD-01, DUP-01, TST-03.
- **feat-107-02** Estado de delegações (`.forge/runtime/orquestrador/<iniciativa>.json`): leitura estrita, gravação atômica, lock da iniciativa, contador de reentregas. **Depende de:** feat-107-01. Rules: SEG-01, SEG-02, SEG-05, COD-03, DUP-06.
- **feat-107-03** Política de entrada não confiável (relatório e mensagem de sessão par como dado; saneamento; permissão negada não contornada) com testes de injeção. **Depende de:** feat-107-02. Rules: SEG-06, SEG-08, FLX-08.
- **feat-107-04** Reference do papel Orquestrador em `sdd-operacao` (tabela, briefing, retomar, menu, "abrir pasta"), com os templates e goldens. **Depende de:** feat-107-01. Rules: DUP-10, TST-02, COD-05.

- **feat-107-12** WIP humano configurável e opcional: limite de itens aguardando o humano, parada com menu no limite, aviso de riscos antes de reduzir ou desligar, modo gravado no `status.md`; distinto de `builder.max_parallel_tasks`. Local e nome do campo seguem a decisão xv da 106 (padrão proposto: ligado, limite 3). **Depende de:** feat-107-02. Rules: SEG-01, SEG-02, DUP-06, TST-03.

Paralelo: 01 sozinha; 02 e 04 após 01; 03 e 12 após 02.

## Onda 2 — Verificação antes do handoff (depende do script da 106, RN-12)
- **feat-107-05** Contrato de checagens e `forge pipeline verificar --iniciativa <id> --rota <intenção> --json`: gate, commit, branch, diff no escopo, `status.md` até 1 KB, imagens dos diagramas, testes reexecutados. **Depende de:** feat-107-02 e do desenho da 106 (RN-12). Rules: ARQ-01, ARQ-05, ARQ-06, SEG-03, FLX-06.
- **feat-107-06** Adaptador de VCS (`github`, `azure-devops`, `nenhum`) lido da Constituição para o estado do PR. **Depende de:** feat-107-05. Rules: SEG-03, SEG-01, DUP-06.
- **feat-107-07** Retomada: gravar `Delegado:` no `status.md` e reconstruir o estado depois de `/compact` ou reinício. **Depende de:** feat-107-02. Rules: FLX-02, DUP-01.

Paralelo: 06 e 07 após 05 e 02.

## Onda 3 — Independência de ferramenta, tetos e avaliação
- **feat-107-08** Consumir `Capabilities` (104-09): rota cai para `sessions` com aviso onde falta equivalente de subagente. **Depende de:** feat-104-09 e feat-107-04. Rules: DUP-03, ARQ-02.
- **feat-107-09** Tetos técnicos de paralelismo e reentrega vindos de `forge.json` (`max_parallel_tasks`, `max_rework_attempts`; 2 ou 3 em `A confirmar`); escalada em menu no teto. **Depende de:** feat-107-02. Rules: DUP-06, TST-03.
- **feat-107-10** Casos de avaliação do orquestrador (injeção, loop, permissão negada) no `forge eval` da 104-11. **Depende de:** feat-107-03, feat-107-09 e feat-104-11. Rules: TST-01, TST-04.
- **feat-107-11** ADR (número no merge), `sdd/spec/concepts.md`, `docs/pipeline.md` e `flows.md`. **Depende de:** ondas anteriores. Rules: FLX-05, GOV-01, DUP-01.

Colisões a sequenciar: templates e goldens com a 105 e a 106 (`TST-02`); número de ADR no merge.

# Plano — Iniciativa 107

Ordem: de baixo para cima (dado e contrato, depois comportamento, depois instrução e documentação). Até 7 tasks por onda. Alinhado à 106 redesenhada: o split é do Builder, uma branch e um PR em draft por feature, estado por feature (o que a 106 ainda não implementou está marcado "a verificar"). Os arquivos de feature nascem no `/split-features` da onda 1. Sem dependência de 104-08/09 (decisão do dono).

## Onda 1 — Tabela, estado, WIP humano e instrução (independe da 106)
- **feat-107-01** Tabela de roteamento como dado validável (`internal/pipeline/roteamento.go`): intenção, papel, modo, isolamento, portão; fluxo da 106 (spec, Conferência 1, [Seguir], split pelo Builder, feature por branch, Conferência 2, [Continuar] por onda, arquivar da principal); papel conferido contra `.claude/agents/`. **Depende de:** nenhuma. Rules: ARQ-01, ARQ-02, COD-01, DUP-01, TST-03.
- **feat-107-02** Estado de delegações de runtime (`.forge/runtime/orquestrador/<iniciativa>.json`): leitura estrita, gravação atômica, lock da iniciativa, contador de 3 tentativas por delegação (reusa `max_rework_attempts`); sem arquivo compartilhado no git. **Depende de:** feat-107-01. Rules: SEG-01, SEG-02, SEG-05, COD-03, DUP-06.
- **feat-107-03** Política de entrada não confiável (relatório e mensagem de sessão par como dado; saneamento; permissão negada não contornada) com testes de injeção. **Depende de:** feat-107-02. Rules: SEG-06, SEG-08, FLX-08.
- **feat-107-04** Reference do papel Orquestrador em `sdd-operacao` (tabela, briefing, retomar, menu, "abrir pasta"), com templates e goldens. **Depende de:** feat-107-01. Rules: DUP-10, TST-02, COD-05.
- **feat-107-12** WIP humano: leitura do campo `WIP humano` da Constituição na branch base, limite 3 por padrão, parada com menu no limite, aviso de riscos antes de reduzir ou desligar, modo no `status.md`; distinto de `builder.max_parallel_tasks`; ADR e mudança da Constituição por `GOV-01` nesta task (a Constituição só é editada aqui, com ADR). Leitor da tabela compartilhado com o da 106: **a verificar**. **Depende de:** feat-107-02. Rules: SEG-01, SEG-02, GOV-01, FLX-05, DUP-06, TST-03.

Paralelo: 01 sozinha; 02 e 04 após 01; 03 e 12 após 02.

## Onda 2 — Verificação antes do handoff (só depois de a 106 fixar o script, RN-12)
- **feat-107-05** `forge pipeline verificar --ponto spec|entrega --feature <sid> --json`, sobre o contrato de checagens do script da 106: gate, commit, branch, diff no escopo da feature, `status.md` até 1 KB, imagens dos diagramas, testes reexecutados. **Depende de:** feat-107-02 e do script da 106. Rules: ARQ-01, ARQ-05, ARQ-06, SEG-03, FLX-06.
- **feat-107-06** Consumo do adaptador de VCS da 106 (`github`, `azure-devops`, `nenhum`) para o estado do PR por feature; sem adaptador próprio. **Depende de:** feat-107-05 e do adaptador da 106 (a verificar). Rules: SEG-03, SEG-01, DUP-06.
- **feat-107-07** Retomada: reconstruir delegações abertas de runtime e o estado por feature da 106 depois de `/compact` ou reinício. **Depende de:** feat-107-02 e da decisão xvi da 106 (a verificar). Rules: FLX-02, DUP-01.

## Onda 3 — Tetos, ferramentas e avaliação
- **feat-107-09** Teto técnico de paralelismo (`max_parallel_tasks`) e de tentativas (3 por delegação, `max_rework_attempts`); escalada em menu no teto. **Depende de:** feat-107-02. Rules: DUP-06, TST-03.
- **feat-107-08** Fallback para o modo `sessions` com aviso onde falta equivalente de subagente, com a matriz mínima da própria 107 (sem `Capabilities` da 104-09). **Depende de:** feat-107-04. Rules: DUP-03, ARQ-02.
- **feat-107-10** Casos de avaliação do orquestrador (injeção, loop, permissão negada, WIP no limite) no `forge eval` da 104-11. **Depende de:** feat-107-03, feat-107-09, feat-107-12 e feat-104-11 (já mesclada: a verificar). Rules: TST-01, TST-04.
- **feat-107-11** `sdd/spec/concepts.md`, `docs/pipeline.md`, `flows.md` e o texto final da ADR (número no merge). **Depende de:** ondas anteriores. Rules: FLX-05, GOV-01, DUP-01.

Colisões a sequenciar: templates e goldens com a 105 e a 106 (`TST-02`); número de ADR no merge; edição da Constituição (a 106 também altera `FLX-01`, `FLX-02`, Regra 15 e os campos de modo) na mesma tabela.

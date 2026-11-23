# fix-118-03 — Template do `veredito.md` sem os campos do Parecer: `init` novo falha no `doctor`

**Branch:** `fix/118-03-template-do-veredito` (spec e correção no mesmo PR, por urgência, com aval do dono em 23/11/2026)
**Origem:** Iniciativa 118, teste da 5.2.2-beta no `hub-busize`: o `update` trocou o `veredito.md` por uma versão sem os campos e o `forge doctor --harness` passou a falhar.
**Tamanho:** `S`
**Status:** `spec`

## Sintoma

Em projeto novo (`forge init`) ou atualizado pela 5.2.2-beta, `forge doctor --harness` sai com código 3: `[falha] contrato-skill-binario: campos do contrato Parecer ausentes no exemplo da skill: head_sha, ponto, modo, feature`. Como o `doctor --harness` é o hook `SessionStart`, o Claude Code mostra erro a cada abertura. O `update` também **rebaixa** o arquivo de quem já o tinha correto.

## Feature Correlacionada

- **Arquivo de origem:** `internal/scaffold/templates/.agents/skills/sdd-governanca/references/veredito.md.tmpl` e o golden `testdata/golden/.agents/skills/sdd-governanca/references/veredito.md`; origem do contrato: feat-106-11 (#159), que acrescentou os campos só à cópia do repositório.
- **Causa-raiz:** o template embutido nunca recebeu os campos `head_sha`, `ponto`, `modo` e `feature`. O teste de "doctor saudável" lê o `veredito.md` do repositório, não o que o scaffold grava.
- **Evidência:** `forge init` em diretório vazio + `forge doctor --harness` → `[falha] contrato-skill-binario`; `diff` entre a cópia do repo e o template: 5 linhas.

## Modelo do módulo

**Enxuto** (`ARQ-11`): troca de conteúdo de template e golden, mais um teste.

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`TST-02` (template e golden no mesmo PR), `TST-03`, `DUP-04` (definição única do contrato: o template passa a ser igual à cópia do repositório).

## Decisões do projeto aplicadas

- Template e golden passam a ser **idênticos** à cópia do repositório (fonte única, `DUP-04`).
- O teste novo exercita o que o `init` grava, não o que o repositório tem.
- Vale uma beta nova (5.2.3), porque a 5.2.2-beta quebra o `doctor` em todo projeto.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Conteúdo certo | `.agents/skills/sdd-governanca/references/veredito.md` | tem os 4 campos | **reusar** como fonte do template |
| Teste do doctor saudável | `TestDoctorHarnessSaudavelNaoEscreveEmStderr` | lê o arquivo do repo | **estender** com um teste sobre o `init` |

## Comportamento esperado

1. `forge init` grava um `veredito.md` com `head_sha`, `ponto`, `modo` e `feature`.
2. `forge doctor --harness` num projeto recém-criado não tem nenhuma `[falha]`.
3. O `update` deixa de rebaixar o arquivo.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
go test ./cmd/forge-sdd -run '^TestInitNovoPassaNoDoctorHarness$' -count=1
[ "$(go test ./cmd/forge-sdd -list '^TestInitNovoPassaNoDoctorHarness$' | grep -c '^Test')" -eq 1 ]
go test ./... -count=1
```

## Fora do escopo

- Outros arquivos de skill que possam estar fora de sincronia entre repositório e template: o teste novo cobre só o contrato do doctor.

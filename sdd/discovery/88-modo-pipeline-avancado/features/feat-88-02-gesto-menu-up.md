# feat-88-02 — Gesto →/← no menu para ligar o modo avançado

**Iniciativa:** 88-modo-pipeline-avancado
**Branch:** feat/88-modo-pipeline-avancado
**Status:** ready
**Depende de:** feat-88-01

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** `cmd/forge-sdd/menu.go` e `internal/survey` (ARQ-01): só adapta a CLI; a regra fica em `config`.

## Decisões do projeto aplicadas
- ADR-0005 (menu é o mesmo código do subcomando, D-7), ADR-0013.
- Rules abaixo; nenhuma ADR depreciada usada como base.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Seleção no menu | `internal/survey/select.go` (`SelectOne`) | Select do `huh` | estender | Captura das teclas →/← no item `up`; validar suporte no `huh`. |
| Fluxo do up | `menuEnv.startSessions` | Escolhe agentes, permissões, preflight e abre | estender | Insere a confirmação de ambiente isolado só quando `advanced`. |

## Dependências novas
- Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-06`, `DUP-01`, `COD-03`, `TST-01`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`.

## Escopo
1. No item "Iniciar esteira (up)": → liga e ← desliga o modo avançado; nenhum rótulo do menu o menciona antes de ligado.
2. Ao ligar, pergunta se o ambiente é isolado; sem confirmação nada é gravado.
3. A linha "Abrir agora?" mostra `avançado`.
4. Fallback se o `huh` não capturar as setas: subcomando `forge pipeline autonomy normal|advanced` (ARQ-06: subcomando de `pipeline`).

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./cmd/forge-sdd -run 'TestMenu_(Autonomy|UpGesto)'
```
- `TestMenu_UpGesto_DireitaLiga`, `TestMenu_UpGesto_EsquerdaDesliga`, `TestMenu_Autonomy_SemConfirmacaoNaoGrava`.
- `go vet ./...` limpo.

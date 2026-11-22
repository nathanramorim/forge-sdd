# Feature 115-04 — O tmux volta como opção

## Objetivo

Decisão do dono: na 115 o tmux **continua opcional**. A 115-02 (#224) o removeu do código; esta feature o reintroduz como opção, sem voltar a ser requisito. O padrão continua sendo as janelas do sistema em headless, e quem quiser o painel vivo em tmux o pede explicitamente.

## Modelo do módulo

Enxuto (`ARQ-11`): restaurar em `internal/pipeline` o que a 115-02 apagou e religar `cmd/forge-sdd`; sem pacote novo, sem dependência Go nova. O ponto de partida é o código do commit pai do #224, não uma reescrita:

```sh
git show 938d5b8^:internal/pipeline/tmux.go
git show 938d5b8^:internal/pipeline/tmuxinstall.go
git show 938d5b8^:internal/pipeline/deliver.go
```

## Decisões do projeto aplicadas

- ADR-0024 (proposta): tmux é opção de interface, não requisito para a continuidade do pipeline.
- ADR-0004 e ADR-0005: jobs trafegam pelo bus; CLI e argv por papel.
- Contrato atual da `main` (pós-#224): `forge up` abre janelas do sistema; `--terminal` está obsoleta e ignorada; o modo `live` cai para headless.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código do tmux | `git show 938d5b8^:internal/pipeline/tmux*.go`, `deliver.go` | Removido pela 115-02, com testes | Restaurar e adaptar ao contrato atual | Evita reescrever e herda os testes |
| Flag `--terminal` | `cmd/forge-sdd/up.go` | Registrada, obsoleta e ignorada | Reativar com `auto`, `native` e `tmux`; `auto` = janelas do sistema | Opcional: sem a flag nada muda |
| Modo `live` | `cmd/forge-sdd/up.go`, `menu.go` | Cai para headless; opção removida do menu | Voltar só com `terminal: tmux` | Live depende de painel |
| Validação de `terminal` | `internal/config/forge_profiles.go` | `ValidTerminal` ainda aceita `tmux` | Manter | Config legada segue válida |
| Painel vivo | `internal/pipeline/status_live.go`, `status.go` | `status --tmux` obsoleta | Religar a flag quando o tmux existir | Opção, não requisito |

## Dependências novas

Nenhuma dependência Go (`SEG-07`). O tmux é programa externo opcional, detectado em runtime; sem ele, `forge up` avisa e segue em janelas do sistema.

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-11`, `DUP-01`, `FLX-01`, `SEG-01`, `TST-01`, `TST-03`, `TST-05`.

## Critérios de Aceitação Executáveis

1. **Padrão inalterado:** sem `--terminal` e sem `terminal: tmux` na config, `forge up` abre as janelas do sistema, como na `main` atual.
2. **Opção explícita:** com `--terminal tmux` (ou `terminal: tmux`) e o tmux instalado, `forge up` abre as sessões em tmux; `--dry-run` imprime os comandos sem abrir nada.
3. **Sem tmux, sem falha:** com `--terminal tmux` e o tmux ausente, `forge up` avisa e cai para as janelas do sistema, sem erro.
4. **Valor inválido** em `--terminal` é recusado; `auto`, `native` e `tmux` são aceitos.
5. **Live só no tmux:** o modo `live` só vale com `terminal: tmux`; fora dele cai para headless e avisa, como hoje.
6. **O pré-requisito não volta:** `Preflight` não exige o tmux; só avisa quando ele foi escolhido e está ausente.
7. **Windows e macOS em CI:** os testes de tmux real são ignorados quando o tmux não existe.

```sh
go build ./...
go test ./cmd/forge-sdd/... ./internal/pipeline/... ./internal/config/...
go vet ./...
```

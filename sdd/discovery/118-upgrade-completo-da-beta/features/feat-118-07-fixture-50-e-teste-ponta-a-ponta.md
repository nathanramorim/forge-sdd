# feat-118-07 — Fixture 5.0 e teste ponta a ponta `TestUpgradeDe50` para os 4 agentes

**Depende de:** fix-106-14a, fix-106-14b e feat-118-01 a 118-06c (por último)
**Achados:** todos (verificação), mais os itens (a) a (e) do escopo. **Tamanho:** `M`. **Status:** `spec`

## Objetivo

Prova, num projeto 5.0 congelado, que `forge update --upgrade` deixa tudo coerente para Claude, Gemini, OpenAI/Codex e Copilot. Implementa o furo 6 do fix-106-09 ("não há projeto-fixture antigo nos testes"), sem reimplementá-lo (`DUP-03`); o fix-106-09 é marcado "implementado pela 118" no fim.

## Entradas

- `internal/scaffold/testdata` (hoje só `golden`, a versão atual).
- Um projeto 5.0 gerado com o binário 5.0 (instruções com stack e banco, `.sddrc` sem os campos novos, rules e `schema.json` da 5.0, `settings.json` com permissões e hooks do usuário, `.mcp.json` com servidor próprio, `ci.yml` do usuário, `CLAUDE.md` editado pelo usuário).

## Saídas

- `internal/scaffold/testdata/legado-50/` (fixture, sem segredos) por agente ou com os quatro agentes na mesma árvore.
- `TestUpgradeDe50` em `./internal/scaffold` (camada de arquivos) e em `./cmd/forge-sdd` (`update --upgrade --yes` de ponta a ponta), verificando: stack e banco intactos; instrução editada intocada; hook só com script; nenhum hook sem script (`TestNenhumHookSemScript`, 14a); rules não editadas atualizadas e editadas preservadas com `.novo`; campos de modo acrescentados sem alterar valores; `ci.yml` byte a byte igual; sem `forge-pronto.yml` sem `--ci`; `forge.json` não sobrescrito (a); `.sdd-version` e `.sddrc` com a mesma versão (c); layout "nenhuma migração necessária" (d); idempotência na segunda execução; `doctor` sem erro no fim.
- `TestUpdateNaoPerdeDadoDoUsuario` completo (stack e instrução).

## Modelo do módulo

**Completo** (`ARQ-11`): testes sobre `scaffold`, `config`, `doctor` e `update`, sem código de produção novo. Justificativa: o defeito atravessa as camadas.

## Decisões do projeto aplicadas

- `TST-03`: casos conformes e negativos por agente (tabela com os quatro).
- `TST-02`: fixture e goldens no mesmo PR.
- `DUP-03`: reusa o que 01 a 06c entregam; falha aqui aponta a feature responsável, não cria código paralelo.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Fixture antiga | `internal/scaffold/testdata` | só `golden` | **criar** `legado-50/` | fix-106-09 furo 6 |
| Teste de `update` | `forge_templates_test.go:56` `TestUpdate_NaoSobrescreveForgeJsonAtivo`, `scaffold_test.go:286,343` | casos pontuais | **reusar** como base de helper | `DUP-03` |
| Helper de árvore | `doctor_versao_test.go:99` `snapshotArquivos` | snapshot de diretório | **reusar** | idempotência |
| Teste de hook sem script | fix-106-14a `TestNenhumHookSemScript` | spec pronta | **reusar** | RN-1 |
| Teste ponta a ponta do CLI | `claude_integration_test.go` | cobre a integração do Claude | **estender** | `cmd/forge-sdd` |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`TST-03`, `TST-02`, `DUP-03`, `DUP-02`, `SEG-02` (a fixture prova a preservação), `ARQ-11`.

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**; `TestUpgradeDe50` só fica verde depois das features 118-01 a 118-06c.

```bash
set -o pipefail
go test ./internal/scaffold ./cmd/forge-sdd -run '^TestUpgradeDe50$' -count=1
go test ./internal/scaffold -run '^(TestUpgradeDe50Idempotente|TestUpgradeDe50NaoTocaNoCiYml|TestNenhumHookSemScript|TestUpdateNaoPerdeDadoDoUsuario)$' -count=1
[ "$(go test ./internal/scaffold ./cmd/forge-sdd -list '^(TestUpgradeDe50|TestUpgradeDe50Idempotente|TestUpgradeDe50NaoTocaNoCiYml|TestNenhumHookSemScript|TestUpdateNaoPerdeDadoDoUsuario)$' | grep -c '^Test')" -eq 6 ]
go test ./internal/scaffold ./cmd/forge-sdd -count=1
```

A contagem é 6 porque `TestUpgradeDe50` existe nos dois pacotes (`./internal/scaffold` e `./cmd/forge-sdd`).

## Fora do escopo

- Hooks para Gemini, Copilot e Codex (não têm hook).
- Achado 11 (reserialização do `settings.json`).

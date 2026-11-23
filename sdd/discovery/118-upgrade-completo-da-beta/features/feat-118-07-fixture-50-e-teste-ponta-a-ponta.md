# feat-118-07 — Fixture 5.0 e teste ponta a ponta `TestUpgradeDe50` para os 4 agentes

**Depende de:** fix-106-14a e feat-118-01 a 118-06c (por último). **Não** depende da 14b (implementação pendente do aceite da ADR-0027); quando a 14b chegar, o teste dela, `TestHookPRGuardMigraComandoAntigo`, entra nesta suíte em PR próprio.
**Achados:** todos (verificação), mais os itens (a) a (e) do escopo. **Tamanho:** `M`. **Status:** `spec`

## Objetivo

Prova, num projeto 5.0 congelado, que `forge update --upgrade` deixa tudo coerente para Claude, Gemini, OpenAI/Codex e Copilot. Implementa o furo 6 do fix-106-09 ("não há projeto-fixture antigo nos testes"), sem reimplementá-lo (`DUP-03`); o fix-106-09 **não** é marcado "implementado pela 118": o furo 3 (esquema de versão do `queue.json` e dos estados) fica nele.

## Entradas

- `internal/scaffold/testdata` (hoje só `golden`, a versão atual).
- Um projeto 5.0 gerado com o binário 5.0 (instruções com stack e banco, `.sddrc` sem os campos novos, rules e `schema.json` da 5.0, `settings.json` com permissões e hooks do usuário, `.mcp.json` com servidor próprio, `ci.yml` do usuário, `CLAUDE.md` editado pelo usuário).

## Saídas

- `internal/scaffold/testdata/legado-50/` (fixture, sem segredos) por agente ou com os quatro agentes na mesma árvore.
- `TestUpgradeDe50` em `./internal/scaffold` (camada de arquivos) e em `./cmd/forge-sdd` (`update --upgrade --yes` de ponta a ponta), verificando: stack e banco intactos; instrução editada intocada; hook só com script; nenhum hook sem script (teste **novo desta feature**, `TestUpgradeDe50NenhumHookSemScript`; os da 14a, como `TestUpdateEntregaPRGuardAntesDoHook`, cobrem a unidade); rules não editadas atualizadas e editadas preservadas com `.novo`; campos de modo acrescentados sem alterar valores; `ci.yml` byte a byte igual; sem `forge-pronto.yml` sem `--ci`; `forge.json` não sobrescrito (a); `.sdd-version` e `.sddrc` com a mesma versão (c); layout "nenhuma migração necessária" (d); idempotência na segunda execução; `doctor` sem erro no fim.
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
| Testes do hook e do script | fix-106-14a (`TestUpdateEntregaPRGuardAntesDoHook`, `TestUpdateReparaHookSemScript`) | spec pronta | **reusar** a ideia; **criar** o ponta a ponta | RN-1 |
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
go test ./internal/scaffold -run '^(TestUpgradeDe50Idempotente|TestUpgradeDe50NaoTocaNoCiYml|TestUpgradeDe50NenhumHookSemScript|TestUpdateNaoPerdeDadoDoUsuario)$' -count=1
[ "$(go test ./internal/scaffold ./cmd/forge-sdd -list '^(TestUpgradeDe50|TestUpgradeDe50Idempotente|TestUpgradeDe50NaoTocaNoCiYml|TestUpgradeDe50NenhumHookSemScript|TestUpdateNaoPerdeDadoDoUsuario)$' | grep -c '^Test')" -eq 6 ]
go test ./internal/scaffold ./cmd/forge-sdd -count=1
```

A contagem é 6: `TestUpgradeDe50` existe nos dois pacotes (2) mais `TestUpgradeDe50Idempotente`, `TestUpgradeDe50NaoTocaNoCiYml`, `TestUpgradeDe50NenhumHookSemScript` e `TestUpdateNaoPerdeDadoDoUsuario` (1 cada, em `./internal/scaffold`).

## Fora do escopo

- Hooks para Gemini, Copilot e Codex (não têm hook).
- Achado 11 (reserialização do `settings.json`).

## Nota da Conferência 1

`snapshotArquivos` está em `cmd/forge-sdd` (`doctor_versao_test.go:99`): o `TestUpgradeDe50` de `internal/scaffold` não pode reusá-lo; o Builder cria um helper próprio em `internal/testutil` (ou equivalente) e o `cmd/forge-sdd` pode passar a usá-lo.

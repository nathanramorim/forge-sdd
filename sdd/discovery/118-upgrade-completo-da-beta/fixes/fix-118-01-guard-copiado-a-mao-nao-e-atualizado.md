# fix-118-01 — `scripts/pr_guard.py` copiado à mão fica "do usuário" e nunca é atualizado

**Branch da spec:** `fix/118-spec-guard-e-hook-de-sessao`; **da correção:** `fix/118-01-impl-adocao-do-guard`, a partir da principal
**Origem:** Iniciativa 118, teste da 5.2.1-beta no projeto `hub-busize` (23/11/2026); o dono notou que, sem atualizar o guard, o plano da 106 fica sem efeito.
**Tamanho:** `M`
**Status:** `spec`

## Sintoma

`forge update --upgrade` (5.2.1-beta) imprime `mantido scripts/pr_guard.py (sem o marcador forge-sdd:generated; é do usuário e não é atualizado)`. O arquivo não foi editado por ninguém: é cópia byte a byte do guard do repositório, criada às 08:10 pelo fluxo antigo (`/upgrade-sdd` no agente). Como o hook `forge hook pr-guard` executa esse script (`internal/hook/prguard.go`, `uv run scripts/pr_guard.py --hook`), toda correção futura do guard deixa de chegar a esse projeto, e o aviso não diz o que fazer.

## Feature Correlacionada

- **Arquivo de origem:** `internal/scaffold/pr_guard_script.go` (`IsGeneratedPRGuard`, `PRGuardUserOwned`) e a regra de `shouldPreserve` em `scaffold.go`, feat-118-02/04 e fix-106-14a.
- **Causa-raiz:** o marcador `# forge-sdd:generated v<versão>` só é injetado na gravação pelo instalador; o script que o agente copiou do repositório (5.2.0-beta e 5.2.1-beta) não o tem, e a adoção por hash existe só para os arquivos da 5.0 (`internal/scaffold/v50/`), não para o guard.
- **Evidência:** sha256 de `scripts/pr_guard.py` em `hub-busize` = `80ada8e16d14…` = o do repositório em `v5.2.1-beta`; o de `v5.2.0-beta` é `ae69dab97db8…`.

## Modelo do módulo

**Enxuto** (`ARQ-11`): tabela de hashes e uma checagem na decisão de preservar; sem pacote novo. Justificativa: estende `shouldPreserve` e o plano do update, como a 118-04.

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`SEG-02` (nunca sobrescrever arquivo do usuário; só adotar por igualdade byte a byte), `SEG-01` (hash calculado sobre o arquivo lido, sem seguir link simbólico), `TST-02`, `TST-03` (caso negativo: byte diferente continua do usuário), `DUP-02`.

## Decisões do projeto aplicadas

- Adoção só por **igualdade byte a byte** com uma versão que o forge publicou (mesma regra da 118-04); um byte diferente = do usuário = intocado.
- Guard do usuário (editado de verdade) não é trocado, mas o `update` grava `scripts/pr_guard.novo.py` ao lado e avisa que a barreira de PR **não** receberá correções enquanto ele não for revisado (mesmo padrão dos `.novo.md`).
- Tabela `prGuardPublicados` (sha256 → versão) em `internal/scaffold/`, com os hashes de `v5.2.0-beta` e `v5.2.1-beta`; cada release nova acrescenta o seu.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Adoção por hash | `internal/scaffold/v50/` e `decideDestino` | só arquivos da 5.0 | **estender** com a tabela do guard |
| Marcador | `prGuardMarkerVersion`, `installPRGuardScript` | só marca na gravação | **reusar**: adotado = regravar com o marcador |
| Aviso de arquivo mantido | `.novo.md` das instruções | existe | **reusar** o padrão para `.novo.py` |

## Comportamento esperado

1. `scripts/pr_guard.py` sem marcador e com sha256 na tabela é tratado como gerado: o `update` o regrava com o marcador da versão atual (`update --dry-run` mostra `atualizar scripts/pr_guard.py (igual ao entregue pelo forge na 5.2.x)`).
2. Sem marcador e com hash fora da tabela: continua intocado, o `update` grava `scripts/pr_guard.novo.py` e imprime o aviso da barreira sem correções.
3. Com marcador e versão menor: atualizado (já existe).
4. `forge doctor` diz, para o guard do usuário, que ele é do usuário e não recebe atualização.

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**. A aprovação é o `-run`, com nomes exatos.

```bash
set -o pipefail
go test ./internal/scaffold -run '^(TestGuardCopiadoAMaoIgualAoPublicadoEAdotado|TestGuardEditadoFicaIntactoEGravaNovoPy|TestTabelaDeHashesDoGuardBateComAsTags)$' -count=1
go test ./cmd/forge-sdd -run '^TestDoctorAvisaGuardDoUsuario$' -count=1
[ "$(go test ./internal/scaffold ./cmd/forge-sdd -list '^(TestGuardCopiadoAMaoIgualAoPublicadoEAdotado|TestGuardEditadoFicaIntactoEGravaNovoPy|TestTabelaDeHashesDoGuardBateComAsTags|TestDoctorAvisaGuardDoUsuario)$' | grep -c '^Test')" -eq 4 ]
go test ./... -count=1
```

**Manual do Builder (relatar no PR):** em `/tmp`, projeto com o `pr_guard.py` de `v5.2.0-beta` copiado à mão: `forge update` o adota e o `head -3` mostra o marcador.

## Fora do escopo

- Arquivos de instrução reescritos pela 5.2.0-beta (CLAUDE.md etc.): ver "Recuperação" na seção da wiki; não há hash fixo porque dependem do projeto.

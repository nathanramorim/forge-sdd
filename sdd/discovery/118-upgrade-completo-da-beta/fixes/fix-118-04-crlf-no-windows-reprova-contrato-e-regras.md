# fix-118-04 — CRLF no Windows reprova o contrato da Governança e marca as rules como editadas

**Branch:** `claude/project-thread-wkj3hi` (linha 5.3) e `claude/project-thread-wkj3hi-6.0` (beta 6.0). Spec e correção no mesmo PR, com aval do dono em 27/11/2026, como no fix-118-03.
**Origem:** Iniciativa 118, migração do `hub-busize` para o Windows (27/11/2026). Fecha os follow-ups "aviso falso tem versão nova" e "CRLF" do `status.md`.
**Tamanho:** `S`
**Status:** `implementado`

## Sintoma

1. `forge doctor --harness` sai com `[falha] contrato-skill-binario: bloco ```json não encontrado` num `veredito.md` correto, só porque o checkout gravou CRLF.
2. `forge doctor` avisa que as 5 rules `sdd-*` e o `schema.json` "têm versão nova e você o editou; revise o .novo ao lado", mas nada foi editado e o `.novo` não existe.

## Feature Correlacionada

- **Arquivo de origem:** `internal/harness/skill.go` (regex `"```json\n"`), `internal/scaffold/regras.go` (`PlanoRegras` e `arquivoDoForge` com `bytes.Equal`), `internal/scaffold/instrucao.go` (`adotavelDe50`), `cmd/forge-sdd/doctor.go` (texto do aviso). Origem: feat-118-04.
- **Causa-raiz:** o forge lê arquivos do disco e os compara ou interpreta byte a byte em LF. No Windows, `core.autocrlf=true` grava CRLF no checkout: a regex não casa e toda rule gerenciada difere do template. O doctor ainda manda revisar um `.novo` que só o `forge update` grava.
- **Evidência:** no `hub-busize` em Windows, 6 de 6 arquivos gerenciados marcados como editados; o mesmo `veredito.md` passa no Mac.

## Modelo do módulo

**Enxuto** (`ARQ-11`): um helper `internal/eol` reusado nos pontos de leitura, mais o bloco do `.gitattributes`.

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`TST-02`, `TST-03`, `DUP-04` (uma definição de normalização de EOL), `SEG-02` (o `.gitattributes` é gravado com `writeFileSafeMode`).

## Decisões do projeto aplicadas

- A correção real é ler aceitando CRLF; o `.gitattributes` é prevenção, porque um clone existente só muda com `git add --renormalize`.
- O `.gitattributes` fixa `eol=lf` só em `sdd/**`, `.agents/**`, `.claude/**` e `.forge/**`. Nada de `* text=auto`: o código do usuário não é tocado. O bloco é acrescentado uma vez (marcador `# forge-sdd:eol`) e preserva o que já existe.
- Sem comando novo: o guia visual não muda.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Normalização de CRLF | `internal/skills/skills.go:276`, `internal/pipeline/deliver.go:512` | `ReplaceAll` inline | **novo helper** `internal/eol`; os dois pontos antigos ficam para outro PR |
| Caminho do `.novo` | `novoPath` em `instrucao.go` | privado | **exportar** como `NovoPath` para o doctor |

## Comportamento esperado

1. `veredito.md` com CRLF passa em `contrato-skill-binario`.
2. Rules e `schema.json` iguais ao template a menos de CRLF não entram em "manter" e não geram `.novo`.
3. Sem `.novo` ao lado, o doctor manda rodar `forge update`, em vez de revisar um arquivo que não existe.
4. `forge init` e `forge update` acrescentam o bloco `eol=lf` ao `.gitattributes`, uma vez só.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
go test ./internal/eol -count=1
go test ./internal/doctor -run '^TestDiagnoseContratoComCRLFPassa$' -count=1
go test ./internal/scaffold -run '^(TestRegrasComCRLFNaoViramEditadas|TestInitGravaGitattributesComLF|TestGitattributesDoUsuarioPreservado)$' -count=1
go test ./... -count=1
```

## Fora do escopo

- Detecção do `forge.CMD` pelo hub (Python): é do `hub-busize`, não do forge-sdd. O Go já resolve PATHEXT com `exec.LookPath`.
- Trocar os `ReplaceAll` inline de `skills.go` e `deliver.go` pelo helper.

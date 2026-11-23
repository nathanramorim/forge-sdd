# feat-118-04 — Rules e `schema.json`: atualização por hash entregue (preserva editadas, avisa)

**Depende de:** feat-118-02 (reusa a decisão única "gerado ou do usuário"); onda 3, em paralelo com 118-03 e 118-05
**Achado:** 6. **Tamanho:** `M`. **Status:** `spec`

## Objetivo

Rule não editada e `sdd/.metrics/schema.json` não editado recebem a versão nova no `update`; arquivo **editado** pelo usuário é preservado, o conteúdo novo vai para `<nome>.novo.<ext>` ao lado e o `update` e o `doctor` avisam. Implementa o **furo 1** do fix-106-09 (rule alterada não chega ao projeto antigo; rules com dono) sem reimplementar o contrato (`DUP-03`). Reusa a tabela `embed.FS` e o hash criados na 118-02 e os estende a rules e `schema.json`.

## Entradas

- Rules `.agents/rules/sdd-*.md` (embed: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`, `sdd-reuso-e-duplicacao`, `sdd-seguranca`, e a de modos quando existir) e `templates/sdd/.metrics/schema.json.tmpl`.
- Hash do que o forge entregou, gravado no `.sddrc` (`entregues`); para projeto 5.0, tabela embed com os hashes conhecidos da 5.0.

## Saídas

- `.sddrc` com o mapa `entregues` (caminho para sha256); atualizado a cada gravação do forge.
- `shouldPreserve` para essas rules e o schema: preserva só quando o arquivo difere do hash entregue; igual = regrava.
- Arquivo `.novo` e aviso "rule X tem versão nova; revise" no `update` e no `doctor`. Rules próprias do usuário (nomes fora do embed) nunca entram.

## Modelo do módulo

**Completo** (`ARQ-11`): `scaffold`, `config` (`.sddrc`), `doctor` (aviso) e testes. Justificativa: atravessa gravação, estado e diagnóstico.

## Decisões do projeto aplicadas

- `SEG-02`: nunca sobrescreve edição; `.novo` gravado de forma atômica e sem seguir link simbólico.
- `DUP-06`: o mesmo mecanismo de dono do 118-02 e do contrato do fix-106-09 (hash em `.sddrc`), sem segundo mecanismo.
- `ARQ-03`: a tabela de hashes da 5.0 vem por `embed.FS`.
- Conflito `.sddrc`: o `UpdateSddrc` (118-01) e esta feature gravam o mesmo arquivo; o `entregues` é montado antes da gravação final, uma só vez.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Preservação de rules | `scaffold.go:330-333` | toda rule existente é preservada | **estender** | causa do achado 6 |
| Contrato de atualização | fix-106-09 §2 (rules com dono, hash no `.sddrc`) | spec sem código | **reusar** o desenho | `DUP-03` |
| Marcador/dono dos adaptadores | feat-118-02 | decisão "gerado ou do usuário" | **reusar** | uma regra só |
| Schema de métricas | `templates/sdd/.metrics/schema.json.tmpl`; `shouldPreserve` preserva `sdd/` inteiro | nunca atualiza | **estender** | exceção por hash |
| Gravação do `.sddrc` | `UpdateSddrc` `scaffold.go:280` | sem `entregues` | **estender** | guarda hashes |

## Dependências novas

Nenhuma (`SEG-07`; `crypto/sha256` da biblioteca padrão).

## Rules e IDs da Constituição Aplicáveis

`SEG-02`, `DUP-06`, `DUP-03`, `DUP-02`, `ARQ-03`, `ARQ-11`, `TST-02` (goldens das rules), `TST-03`.

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**.

```bash
set -o pipefail
go test ./internal/scaffold -run '^(TestRuleNaoEditadaEAtualizada|TestRuleEditadaFicaIntocadaEGeraNovo|TestRuleDe50ReconhecidaPelaTabelaDeHashes|TestSchemaNaoEditadoEAtualizado|TestSchemaEditadoFicaIntocado|TestRuleDoUsuarioNuncaEMexida)$' -count=1
go test ./cmd/forge-sdd -run '^TestDoctorAvisaRuleComVersaoNova$' -count=1
[ "$(go test ./internal/scaffold ./cmd/forge-sdd -list '^(TestRuleNaoEditadaEAtualizada|TestRuleEditadaFicaIntocadaEGeraNovo|TestRuleDe50ReconhecidaPelaTabelaDeHashes|TestSchemaNaoEditadoEAtualizado|TestSchemaEditadoFicaIntocado|TestRuleDoUsuarioNuncaEMexida|TestDoctorAvisaRuleComVersaoNova)$' | grep -c '^Test')" -eq 7 ]
go test ./internal/scaffold ./cmd/forge-sdd -count=1
```

## Fora do escopo

- `queue.json` e esquemas de estado do fix-106-09 (versão de esquema, job rodando, versão mínima do binário).
- Rules da Constituição do usuário (118-03).

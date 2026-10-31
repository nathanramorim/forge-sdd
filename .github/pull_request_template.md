## Spec / Tarefa (`FLX-01`, `FLX-03`)

- **Feature / Task:** `sdd/features/feat-XX-...`
- **Discovery / Criteria (se aplicável):** `sdd/discovery/criteria-XX-...`
- **ADR relacionada (se alterar arquitetura ou constitution `FLX-05`, `GOV-01`):** `N/A`

## O que muda

- [Resumo objetivo das alterações entregues neste Pull Request]

## Impacto

- **Módulos / Pacotes afetados:** [Ex.: `cmd/forge-sdd`, `internal/scaffold`, `.agents/`]
- **Compatibilidade / Breaking changes:** [Descreva se há impacto para usuários existentes ou migração via `forge-sdd update`]
- **Dependências novas (`SEG-07`):** `Nenhuma` (ou liste pacote, versão e justificativa)

## Como testar (`TST-01`, `TST-05`)

```bash
go vet ./...
go test ./...
```

## Rollback

- [Como reverter de forma segura caso apresente regressão após o merge/release]

## Checklist da constitution (`FLX-03`, `FLX-06`)

- [ ] **Spec antes do código (`FLX-01`, `DUP-02`):** Spec aprovada com seção `## O que já existe` (Reuse Check) preenchida.
- [ ] **Arquitetura e Reúso (`ARQ-01..06`, `DUP-01`, `DUP-10`):** Fronteiras `cmd/` vs `internal/` respeitadas e fonte única preservada em `.agents/`.
- [ ] **Segurança (`SEG-01..07`):** Sem segredos no código/diff (`SEG-04`), sem dados sensíveis em logs (`SEG-05`) e dependências justificadas (`SEG-07`).
- [ ] **Qualidade e Testes (`COD-02`, `TST-01..05`):** `go vet ./...` e `go test ./...` passando (incluindo golden files atualizados quando aplicável).
- [ ] **Portão de Governança (`FLX-06`, `GOV-04`):** Camada A (`forge-sdd gate`) e Camada B (Analista de Governança) executadas sem violações abertas.

<!--
Caso este PR dependa de uma exceção temporária autorizada pelo humano (GOV-02 / check P-20),
descomente e preencha integralmente o bloco abaixo com uma data futura válida (AAAA-MM-DD):

```yaml
constitution-exception:
  regra: ARQ-04
  motivo: Justificativa objetiva para a exceção temporária
  prazo: 2026-12-31
  responsavel: "@usuario"
```
-->

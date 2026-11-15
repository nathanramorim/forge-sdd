# Feature 42 — Lançamento da Versão Estável v1.7.0

Esta feature consolida o ciclo de desenvolvimento das versões beta (`v1.6.1-beta.0` a `v1.6.1-beta.3`), preparando e publicando a versão estável oficial `v1.7.0` no NPM e atualizando toda a documentação da metodologia.

## Requisitos e Escopo

1. **Fusão e Consolidação:**
   * Todas as branches de feature beta devem ser mescladas e consolidadas na branch principal `main`.
2. **Bump de Versão:**
   * A versão `1.7.0` deve ser declarada como versão atual do CLI, dos scaffolds e dos metadados de dependência.
3. **Release Notes Oficiais:**
   * O histórico de releases deve conter notas consolidadas detalhadas com todas as novidades agregadas neste ciclo.
4. **Alinhamento dos Golden Tests:**
   * Os scaffolds gerados pelas suites de teste de integração devem ser regenerados com a nova versão estável `1.7.0` para passar nos testes de snapshot de scaffold.

## Critérios de Aceitação Executáveis

1. **Binário e CLI compilam e executam:**
   * Executar `go test ./...` sem falhas.
2. **Comando `version` reporta o novo valor:**
   * O comando `forge-sdd version` compilado localmente deve imprimir `1.7.0`.
3. **Publicação no NPM:**
   * O pipeline de release no GitHub Action deve rodar e publicar o pacote `@nathanramorim/forge-sdd` na versão `1.7.0` estável (canal default).

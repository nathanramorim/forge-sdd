# feat/cli-detect-existing

**Branch:** `feat/cli-detect-existing`
**Fase:** 38
**Depende de:** `feat-37-publish-1.6.0`
**Status:** `todo`

## Objetivo

Melhorar a experiência de uso do CLI (`forge-sdd`) para que, ao ser executado em um diretório que já possui a metodologia (presença do arquivo `sdd/.sddrc` ou pasta `sdd/`), ele detecte automaticamente a presença do projeto e pergunte de forma interativa se o usuário deseja realizar o upgrade para a versão oficial (latest) ou para a versão de teste (beta), mostrando o número exato e a tag correspondente obtidos via NPM Registry. A publicação desta feature deve ser feita como a versão beta `1.6.1-beta.0` (sem publicar como release oficial/estável).

## Requisitos Técnicos

1. **Detecção Automática:**
   - No `rootCmd` e `initCmd`, antes de iniciar o formulário interativo de criação, verificar se o arquivo `sdd/.sddrc` já existe no diretório de destino.
   - Se já existir, direcionar o fluxo para uma pergunta interativa ou redirecionar internamente para a rotina de atualização/upgrade.

2. **Consulta às Versões Remotas (NPM Registry):**
   - Implementar chamada HTTP GET com timeout curto (máx 2 segundos) para `https://registry.npmjs.org/@nathanramorim/forge-sdd`.
   - Extrair as tags `latest` e `beta` do objeto `"dist-tags"` retornado.
   - Tratar falhas de conexão de forma graciosa (caso ocorra timeout ou erro de rede, prosseguir usando as versões locais/estáticas como fallback).

3. **Interface de Upgrade Interativa:**
   - Apresentar opções interativas para o usuário:
     - Não atualizar (manter na versão atual).
     - Atualizar para a versão Oficial: `<versao_latest>` (latest).
     - Atualizar para a versão Beta: `<versao_beta>` (beta).

4. **Publicação Beta (Canal de Teste):**
   - Atualizar referências de versão do CLI para `1.6.1-beta.0`.
   - Empurrar a tag `v1.6.1-beta.0` para disparar os workflows do GitHub e do NPM no canal beta.

## Critério de conclusão

```bash
# 1. A versão atualizada deve conter a tag beta
grep -q "1.6.1-beta.0" GEMINI.md
grep -q "1.6.1-beta.0" cmd/forge-sdd/main.go
grep -q "1.6.1-beta.0" npm/package.json

# 2. Todos os testes unitários e de integração devem passar sem quebras
go test ./...
go vet ./...
```

## Tarefas

- [ ] **38-1** Criar especificação da feature em `sdd/features/feat-38-cli-detect-existing.md`
- [ ] **38-2** Atualizar `sdd/features/index.md` e `sdd/memory/progress.md`
- [ ] **38-3** Implementar função em Go (`internal/config` ou similar) para buscar versões do pacote `@nathanramorim/forge-sdd` no NPM Registry com timeout e tratamento de erros.
- [ ] **38-4** Alterar `cmd/forge-sdd/main.go` para que, ao executar `init` ou o comando base sem parâmetros, se o arquivo `sdd/.sddrc` for detectado no diretório atual, ele pergunte interativamente se deseja atualizar e permita escolher a versão `latest` ou `beta` obtida.
- [ ] **38-5** Atualizar a versão do CLI de `1.6.0` para `1.6.1-beta.0` em todos os arquivos (`GEMINI.md`, `cmd/forge-sdd/main.go`, `internal/config/config.go`, `npm/package.json`, `npm/README.md`, `docs/metodologia-sdd.md`, arquivos de teste e golden files).
- [ ] **38-6** Rodar a suite de testes e validar conformidade localmente.

---
description: Como aplicar validação de entrada, proteção de arquivos de domínio, higiene de segredos/logs e governança de dependências.
globs: ["**/*"]
regras: [SEG-01, SEG-02, SEG-04, SEG-05, SEG-07]
---

# Guia Prático — Segurança, Segredos e Dependências

Procedimentos práticos para atender aos IDs `SEG-01`, `SEG-02`, `SEG-04`, `SEG-05` e `SEG-07` de `sdd/memory/constitution.md` (`GOV-04`).

## 1. Validação de Borda e Escrita Segura (`SEG-01`, `SEG-02`)

- **Saneamento de entradas (`SEG-01`):**
  - Valide flags obrigatórias, enums permitidos (ex.: modo `observacao` vs `bloqueio`, agentes suportados) e schemas JSON antes de iniciar mutações em arquivos.
  - Rejeite identificadores ou slugs que contenham `..`, separadores de diretório inesperados ou caracteres de controle.
- **Preservação de domínio (`SEG-02`):**
  - Toda escrita gerada por `init`, `update` ou importação de governança deve construir caminhos com `filepath.Join(targetDir, ...)` e respeitar `shouldPreserve` para artefatos de memória (`sdd/memory/`), specs (`sdd/spec/`), features (`sdd/features/`) e regras customizadas do usuário em `.agents/rules/`.

## 2. Prevenção de Vazamento de Segredos e Dados Sensíveis (`SEG-04`, `SEG-05`)

- **Higiene do repositório e templates (`SEG-04`):**
  - Em testes e exemplos de documentação/templates, utilize placeholders claramente fictícios (`<SEU_TOKEN_AQUI>`, `example-token-placeholder`) — nunca chaves reais ou strings com formato de token de provedor (`ghp_`, `sk-`, `AKIA`, `BEGIN PRIVATE KEY`).
- **Logs, erros e telemetria (`SEG-05`):**
  - Mensagens de erro e arquivos em `sdd/.metrics/` devem registrar apenas metadados operacionais (comando, duração, status, caminho relativo da feature), sem despejar variáveis de ambiente ou cabeçalhos de autenticação.

## 3. Avaliação de Dependências Externas (`SEG-07`)

- Antes de alterar `go.mod` ou adicionar pacote externo:
  1. Confirme se o requisito não pode ser atendido pela `stdlib` do Go (`ARQ-04`).
  2. Caso indispensável, verifique o status do MCP `context7` em `sdd/memory/mcps.md` (ou consulte a documentação oficial na versão exata).
  3. Preencha a seção `## Dependências novas` na Spec da feature indicando nome do pacote, versão exata, licença e motivo pelo qual a `stdlib` não atende. Se não houver novas bibliotecas, registre explicitamente `Nenhuma.` nessa seção.

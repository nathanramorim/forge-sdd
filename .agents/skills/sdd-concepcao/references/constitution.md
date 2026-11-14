# Comando: constitution

**Uso:** `/constitution` ou "alinhar o projeto".

Garanta que a Constituição e a documentação técnica reflitam a realidade do projeto.
1. **Idioma:** pergunte o idioma do chat e o de commits e PRs; registre em `Regras` de `sdd/memory/constitution.md`.
2. **Nível de Linguagem:** `padrão` ou `iniciante`; registre como `Nível de Linguagem: <valor>`.
3. **Ferramentas:** VCS (`github`, `azure-devops` ou `nenhum`) e se os MCPs `context7` e `git` respondem; registre em `Ferramentas e Integrações` e em `sdd/memory/mcps.md`.
4. **Scan:** leia a raiz e as pastas principais; identifique stack, gerenciadores de dependência e padrões arquiteturais.
5. **Povoamento:** preencha a base de conhecimento arquitetural em `sdd/spec/`:
   - `sdd/spec/stack.md` (stack, versões, ferramentas de teste e build);
   - `sdd/spec/overview.md` (visão geral e diagramas C4 em Mermaid);
   - `sdd/spec/modules.md` (módulos e limites de contexto);
   - `sdd/spec/flows.md` (fluxos principais do sistema);
   - `sdd/spec/decisions.md` (decisões do projeto e índice de ADRs).
6. **ADRs:** crie uma ADR em `sdd/spec/adr/` para cada decisão estrutural (`GOV-01`).
7. **Segurança:** defina as regras e diretrizes de segurança (segredos, validação de entradas e dependências). Projeto vazio: proponha uma arquitetura base.
8. **Fechamento e validação:** execute `forge-sdd brief --check`. Só dê o alinhamento por concluído quando o comando reportar o estado `pronta` (código de saída 0).

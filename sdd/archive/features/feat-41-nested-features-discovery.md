# Feature: Agrupamento de Features e Discovery

## Contexto e Problema
À medida que projetos crescem, os arquivos de especificações de features (`feat-XX-*.md`) e tarefas (`task-YY-*.md`) podem poluir a raiz de `sdd/features/` com dezenas de arquivos soltos. Além disso, quando uma quebra de discovery (`split-features`) gera múltiplos arquivos, eles perdem a relação visual com o discovery original na árvore de diretórios.

## Proposta
1. **Estrutura de Subpastas de Feature:**
   - Ao quebrar uma feature grande em subfeatures/tasks, criar uma pasta correspondente com o nome da feature (ex: `sdd/features/feat-XX-nome-feature/`).
   - Os arquivos de tasks de implementação devem residir dentro dessa pasta (ex: `sdd/features/feat-XX-nome-feature/task-YY-*.md`).
2. **Quebra de Discovery:**
   - Ao converter um plano de discovery `sdd/discovery/disc-XX-nome.md` em features, a pasta gerada sob `sdd/features/` deve refletir o nome do discovery original: `sdd/features/feat-XX-nome/`.
   - As subfeatures geradas a partir daquele discovery devem ser armazenadas dentro desta pasta de feature correspondente.
3. **Varredura Recursiva no `doctor`:**
   - O comando `doctor` do CLI deve realizar uma varredura recursiva de arquivos dentro de `sdd/features/`, listando e validando todas as features (`feat-*`) e tasks (`task-*`) aninhadas em qualquer profundidade.

## Critério de Aceitação
- O CLI `doctor` deve encontrar e diagnosticar features em qualquer nível de subpasta em `sdd/features/`.
- Os testes unitários do comando `doctor` devem cobrir arquivos de feature aninhados.
- Os prompts e instruções de sistema dos 4 agentes de IA devem orientar explicitamente essa organização.

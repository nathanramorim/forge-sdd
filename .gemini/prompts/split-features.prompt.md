# Prompt: split-features

**Uso:** Peça "/split-features [arquivo-de-plano]" ou "quebrar plano em features"

**Ação:**
Você deve atuar como o **Specifier**. Sua tarefa é ler um plano de discovery (\`sdd/discovery/plan-XX.md\`) e quebrá-lo em múltiplas features independentes no SDD.

1. **Análise:** Leia o plano e identifique as etapas lógicas (ex: Infra, API, Frontend, Auth).
2. **Criação:** Crie uma subpasta com o nome do discovery (ex: `sdd/features/feat-XX-<nome-do-discovery>/`). Para cada etapa, crie o arquivo de feature correspondente dentro desta subpasta (ex: `sdd/features/feat-XX-<nome-do-discovery>/feat-XX-YY-<nome-da-feature>.md`).
3. **Indexação:** Atualize o arquivo `sdd/features/index.md` com os caminhos corretos e dependências das novas features aninhadas.

**Regras:**
- Garanta que as features sejam "fatias verticais" de valor sempre que possível.
- Mantenha o comando \`/nova-feature\` apenas para incrementos manuais e pontuais.
- Reporte a lista de features criadas ao final.

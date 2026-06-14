# Skill: Orquestrador

Você é o Orquestrador do forge-sdd. Sua responsabilidade é gerenciar o fluxo da sessão, ler o estado, decidir o próximo passo e delegar tarefas.

## Protocolo de sessão
1. **READ-MIN:** Leia `sdd/memory/progress.md` para entender o estado atual.
2. **Identificar:** Encontre a próxima feature com status `todo`.
3. **Detalhar:** Leia o arquivo `sdd/features/feat-XX.md` correspondente.
4. **Contextualizar:** Se necessário, leia `sdd/memory/constitution.md`.
5. **PLAN:** Reporte o status atual e as próximas tarefas. Aguarde confirmação humana.
6. **Branch:** Antes de qualquer ação de código, verifique ou crie a branch da feature: `git checkout -b <branch>`.
7. **Delegar:** Invoque a lógica de **Builder** (pode ser você mesmo mudando de "mindset" ou usando uma ferramenta) para implementar.
8. **Revisar:** Após a implementação, invoque a lógica de **Revisor**.
## Finalizar (Protocolo de Handoff)
1. **Documentar:** Marque as tasks em `feat-XX.md` como concluídas e atualize `progress.md` e `index.md`.
2. **Commit Automático:** Gere uma mensagem de commit semântica baseada nas mudanças e execute:
   ```bash
   git add . && git commit -m "feat(scope): resumo das mudanças"
   ```
3. **Push:** Suba as alterações para a branch atual:
   ```bash
   git push origin <branch-atual>
   ```
4. **PR Prompt:** Pergunte ao usuário: "Deseja abrir o Pull Request agora?". Se sim, forneça o link (se possível via CLI ou apenas o link padrão do GitHub).
5. **Merge (Opcional/Local):** Se instruído, realize o merge na main local: `git checkout main && git merge --no-ff <branch>`.
6. **Métricas:** Grave as métricas em `sdd/.metrics/session-<ISO8601>.json`.
7. **Archive:** Valide o budget de `progress.md` (≤ 1 KB). Se exceder, acione o **Archivist**.


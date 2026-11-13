# Comando: upgrade-sdd e Migrator

**Uso:** `/upgrade-sdd <versão>` atualiza a estrutura do projeto preservando o domínio.

1. Leia `sdd/.sdd-version` e confirme a versão alvo com o usuário.
2. `which forge-sdd`: se estiver em cache do npm, oriente `rm -rf ~/.cache/forge-sdd` e `npx @nathanramorim/forge-sdd@latest`.
3. Rode `forge-sdd update --dry-run`, apresente o que seria criado, movido ou removido e só então `forge-sdd update`.
4. **Preservação de domínio:** nunca sobrescreva `sdd/memory/constitution.md` fora de blocos demarcados, iniciativas em `sdd/discovery/`, `sdd/spec/` ou `.agents/rules/` customizadas.
5. Liste arquivos criados, atualizados e removidos e sugira `/doctor`.

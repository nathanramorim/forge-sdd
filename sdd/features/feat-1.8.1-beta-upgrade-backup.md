# Feature 1.8.1-beta — Backup Automático de Agentes no Upgrade

Esta especificação define os requisitos para a proteção de customizações locais de agentes (como prompts, skills e arquivos de configuração) durante o ciclo de upgrade/update do Forge-SDD, através de um mecanismo de backup automático não-destrutivo.

## Requisitos e Critérios de Aceitação

1. **Backup Automático Não-Destrutivo:**
   * Quando o CLI for reescrever/sobrescrever um arquivo que já existe no destino (onde a regra `shouldPreserve` do scaffold retorna `false`), o CLI deve criar uma cópia de backup do arquivo original em `sdd/backups/upgrade-<timestamp>/<caminho_relativo>`.
   * Essa regra de backup aplica-se a todos os arquivos fora da pasta `sdd/` gerados para os agentes Copilot, Gemini, Claude e OpenAI.
2. **Log Informativo de Backups:**
   * Durante a execução do CLI, deve ser impresso de forma visível a mensagem informando o backup realizado: `[BACKUP] <destino> -> <backup_destino>`.
3. **Persistência Local do Repositório:**
   * As pastas de backup localizadas em `sdd/backups/` não serão removidas automaticamente pelo CLI, permitindo ao usuário mesclar ou restaurar customizações.

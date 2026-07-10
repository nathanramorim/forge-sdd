# Feature 1.8.0-beta — Integração do Modo Piloto Automático e Nova Convenção de Versionamento

Esta especificação formaliza a integração oficial do modo Piloto Automático (Auto-Pilot) na metodologia Forge-SDD, permitindo o desenvolvimento contínuo em lote sem intervenção humana turno a turno, e introduz a nova convenção simplificada de tags de releases beta.

## Critérios de Aceitação Executáveis

1. **Leitura da Flag de Bypass pelo Orquestrador:**
   * O prompt e chatmode do Orquestrador devem verificar a presença da flag `.sdd-auto-pilot` na raiz do projeto.
   * Se presente, o Orquestrador deve prosseguir automaticamente pulando a etapa de confirmação humana do `PLAN`.
2. **Inclusão do Script Daemon e Guia Técnico no Scaffold:**
   * O CLI, ao executar `init` ou `update`, deve scaffoldar o arquivo de script `sdd/scripts/sdd-daemon.sh` e o guia técnico `sdd/docs/autonomia-autopilot.md`.
3. **Novas Regras de Versionamento de Releases:**
   * Atualizar a Constituição e o manual de metodologia para formalizar as tags beta simplificadas `1.x.y-beta` (sem sufixos numéricos sequenciais extras).

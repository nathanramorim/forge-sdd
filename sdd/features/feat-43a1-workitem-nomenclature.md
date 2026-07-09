# Feature 43a1 — Nomenclatura com Workitem e Hash Hexadecimal

Esta feature altera o padrão de nomenclatura rígida de features e discoveries para dar suporte a identificadores de Workitem de referência (ex: IDs de tarefas de ferramentas de gestão como Jira, ClickUp, etc.) ou hashes hexadecimais aleatórios de 4 caracteres, prevenindo colisões sequenciais no desenvolvimento paralelo.

## Motivação e Contexto

A numeração sequencial rígida (`01`, `02`, etc.) gera bloqueios e conflitos em times de desenvolvimento colaborativos e ambientes com múltiplos agentes de IA concorrentes. Quando duas ou mais frentes de trabalho tentam criar especificações simultâneas, ocorrem colisões que forçam um desenvolvedor a aguardar o merge alheio para obter a próxima numeração válida. Flexibilizar para ID de Workitem ou gerar um hash aleatório assíncrono resolve esse gargalo de concorrência.

## Critérios de Aceitação Executáveis

1. **Solicitação de Workitem:**
   * Toda criação de discovery (`/discovery`) ou nova feature (`/nova-feature`) orientada pelos agentes deve solicitar o Workitem de referência.
2. **Geração de Hash Hexadecimal:**
   * Caso o usuário informe que não possui ou omita o Workitem, o agente deve gerar de forma automatizada um hash hexadecimal de 4 caracteres (ex: `3ec4`) para compor o prefixo (ex: `feat-3ec4-nome-feature.md`).
3. **Padrão Flexível no CLI:**
   * Os comandos CLI (`doctor`, `destroy`, etc.) devem ser capazes de analisar, diagnosticar e gerenciar arquivos nomeados sob essa estrutura de forma recursiva e sem falhas, buscando apenas o prefixo do tipo (`feat-` ou `task-`).

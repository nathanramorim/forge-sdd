# Comando: split-features

**Uso:** Peça "/split-features [arquivo-de-plano]" ou "quebrar plano em features"

**Ação:**
O **Líder** (`.agents/roles/lider.md`) junto ao **Planejador** (`.agents/roles/planner.md`) automatiza a quebra do plano revisado em tasks prontas para a fila do Builder:

1. **Scan e Gate de Spec (`FLX-01`):** Leia o arquivo `sdd/discovery/plan-XX.md` (e o respectivo `criteria-XX.md`) após confirmação humana da revisão da Spec.
2. **Branch Única (`Regra 15` / `FLX-02`):** Verifique (`git branch --list feat/XX*`) se já existe branch para a feature e pergunte qual branch usar como ponto de partida (default `main`). Crie/ative a branch única `feat/XX-<nome-do-discovery>` antes de criar os arquivos de especificação.
3. **Criação das Tasks:** Crie a subpasta `sdd/features/feat-XX-<nome-do-discovery>/`. Para cada etapa do plano, crie `feat-XX-YY-<nome-da-task>.md` contendo obrigatoriamente:
   - `## Modelo do módulo` (`completo | enxuto` + justificativa — `ARQ-11`);
   - `## O que já existe` (tabela de Reuse Check — `DUP-02`);
   - `## Dependências novas` (`SEG-07`);
   - `## Rules e IDs da Constituição Aplicáveis` (arquivos `.agents/rules/` e IDs que o Builder e o Analista de Governança devem carregar);
   - `## Critérios de Aceitação Executáveis`.
4. **Mapeamento e Fila do Builder:** Atualize `sdd/features/index.md` e `sdd/memory/progress.md`. Quando confirmado pelo humano no Líder, enfileire as tasks com `forge-sdd queue enqueue` para execução pelo Builder.
5. **Gravação de Métricas (determinística):** Se a telemetria estiver habilitada em `sdd/.sddrc`, execute `forge-sdd session record --feature "<pasta sdd/features/feat-XX-nome/ criada>" --outcome approved --criterio-atendido=true`.

**Regra:** Utilize o comando `/nova-feature` apenas para adições manuais e isoladas.

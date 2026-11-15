# Comando: novo-fix

**Uso:** Peça "/novo-fix <descrição do bug>"

**Ação:**
0. **Clarify:** Antes do PASSO 1, avalie a descrição recebida contra a lógica única descrita em `sdd/memory/clarify.md`. Se algum sinal de lacuna for detectado, faça a rodada de perguntas antes de prosseguir; caso contrário, siga direto.
1. **PASSO 1 MANDATÓRIO — Branch:**
   - **Branch de partida:** Pergunte ao usuário qual branch usar como ponto de partida (default `main`). Só prossiga sem perguntar se o usuário já indicou explicitamente nesta mesma solicitação.
   - **Retomada:** Rode `git branch --list fix/*` para esse fix. Se já existir uma branch correspondente de uma sessão anterior, pergunte se deve continuar a partir dela em vez de criar uma nova.
   - Crie a branch do fix localmente no terminal (`git checkout -b fix/<nome>`, a partir da branch de partida escolhida) **antes** de criar qualquer especificação ou logs.
2. **Nomenclatura:** Siga a lógica única descrita em `sdd/memory/naming-convention.md`.
3. **Criação e Co-relação:**
   - **Identificação da Feature de Origem:** Todo fix deve co-relacionar a feature ou iniciativa que está corrigindo. Identifique ou pergunte ao usuário qual iniciativa/feature gerou o bug.
   - **Localização em Pasta:**
     - Se a feature pertencer a uma iniciativa de Discovery: crie o fix dentro da subpasta `fixes/` da iniciativa correspondente: `sdd/discovery/<id>-<resume>/fixes/fix-<id>-<nome>.md`.
     - Se a feature for independente: crie dentro de `sdd/features/<id>-<nome>/fixes/fix-<id>-<nome>.md` (suporta compatibilidade com `sdd/features/fix-ID-<nome>.md` legados).
   - **Seção Obrigatória de Co-relação:** O arquivo do fix deve conter a seção `## Feature Correlacionada` documentando o arquivo de origem corrigido, a causa-raiz do bug e a evidência do defeito, além das seções `## Modelo do módulo`, `## O que já existe`, `## Dependências novas`, `## Rules e IDs da Constituição Aplicáveis` e `## Critérios de Aceitação Executáveis`. Atualize o progresso e índice.
4. **Gravação de Métricas:** Se a telemetria estiver habilitada em `sdd/.sddrc`, execute `forge-sdd session record --feature "<arquivo fix criado>" --outcome blocked --criterio-atendido=false` (o fix ainda não foi implementado) — garante telemetria mesmo se a sessão terminar aqui.

**Handoff:** Gere um resumo para o comando `proxima-feature`.

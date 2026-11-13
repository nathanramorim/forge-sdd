# Comando: discovery

**Uso:** Peça "/discovery <descrição da demanda>"

**Ação:**
O **Líder** (`.agents/roles/lider.md`) coordena a concepção junto ao **Product Manager + Arquiteto** (`.agents/roles/pm-architect.md`) e ao **Planejador** (`.agents/roles/planner.md`):

0. **Clarify:** Antes de produzir os três artefatos, avalie a demanda recebida contra a lógica única descrita em `sdd/memory/clarify.md`. Se algum sinal de lacuna for detectado, faça a rodada de perguntas antes de prosseguir; caso contrário, siga direto.
1. **Concepção de Produto (PM):** Identifique o "porquê", o "para quem" e o "como" macro em `discovery-ID-<nome>.md`.
2. **Concepção de Arquitetura (Arquiteto):** Identifique restrições (citando os IDs de `sdd/memory/constitution.md`), diagramas **C4 Model (Mermaid)** e critérios de aceitação executáveis em `criteria-ID-<nome>.md`, incluindo obrigatoriamente as três seções de governança:
   - `## Modelo do módulo` (`completo | enxuto` + justificativa — `ARQ-11`);
   - `## O que já existe` (tabela `Procurei | Onde | Achei | Decisão (reusar, estender, criar) | Justificativa` — `DUP-02`);
   - `## Dependências novas` (`SEG-07`).
3. **Planejamento (`Plan`):** O Planejador consulta os IDs da Constituição (`sdd/memory/constitution.md`) e os frontmatters de `.agents/rules/*.md` para gerar `plan-ID-<nome>.md`, anotando quais rules/IDs e skills cada task da fila utilizará.

**Nomenclatura e ID do Discovery:** Siga a lógica única descrita em `sdd/memory/naming-convention.md` (aplicada ao prefixo `discovery-`).

**Entrega:**
Gere uma pasta autocontida para a iniciativa em `sdd/discovery/<id>-<nome>/`:
- `discovery.md`: Visão Produto/Negócio.
- `criteria.md`: Visão Técnica/Engenharia (com `Modelo do módulo`, `O que já existe`, `Dependências novas` e C4 Mermaid).
- `plan.md`: Roadmap preliminar, estimativa de quebra de tarefas/features e mapeamento de rules/IDs por task.
- `features/`: Subpasta reservada para as features atômicas fatiadas no comando `/split-features`.
- `fixes/`: Subpasta reservada para correções correlacionadas geradas pelo comando `/novo-fix`.

Mantenha o padrão Forge-SDD de escrita limpa e objetiva.

Verifique `Nível de Linguagem` em `sdd/memory/constitution.md`: se `iniciante`, explique conceitos como "C4 Model" ou "critério de aceitação" com exemplos concretos e linguagem simples, sem alterar os critérios de aceitação em si.

**Gravação de Métricas (determinística):** Se a telemetria estiver habilitada em `sdd/.sddrc`, execute `forge-sdd session record --feature "sdd/discovery/<id>-<nome>/discovery.md" --outcome approved --criterio-atendido=true` — garante telemetria mesmo quando a sessão não avança para `/split-features`.

**Gate 1 (Revisão Obrigatória de Spec — `FLX-01`) e Handoff:** O **Líder** submete a Spec gerada à revisão humana. Após aprovação, gere o resumo para o próximo passo (`/split-features`), listando os arquivos criados e instruindo a quebrar as features diretamente na subpasta `sdd/discovery/<id>-<nome>/features/` e enfileirá-las para o Builder.

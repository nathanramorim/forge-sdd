# Plano Preliminar 03 — Confirmação de Subagente, Comando Claude Quebrado, `/status` sem Sincronização e Clarify

Quatro frentes majoritariamente independentes entre si — ordem sugerida por impacto imediato (comando quebrado hoje) e depois por risco/superfície.

## Estimativa de Quebra de Features

1. **Corrigir nome do adaptador Claude (`.prompt.md` → `.md`)**
   Renomear os 12 arquivos em `.claude/commands/` e o template-fonte (`internal/scaffold/templates/agents/claude/.claude/commands/*.md.tmpl`), atualizar golden fixtures. Maior prioridade: hoje o comando sugerido por qualquer handoff (`/nova-feature`, `/status`, etc.) não executa por padrão em Claude. Baixo risco técnico (rename simples), mas superfície ampla (12 arquivos + fixtures + testes).

2. **Migração em `forge-sdd update` para o novo nome do adaptador Claude**
   Estender `update` (mesmo ponto de extensão já usado para `.agents/`) com um passo aditivo/idempotente: detecta `.claude/commands/*.prompt.md` remanescente, compara com template conhecido, renomeia se não customizado ou preserva conteúdo + avisa se customizado. Depende da feature 1 (precisa do novo template existir primeiro).

3. **Sincronização remota em `/status`**
   Adicionar etapa de `git fetch` + comparação ahead/behind + `gh pr list` (condicionada a VCS = `github`) em `.agents/commands/status.md`, com nova seção "Divergência Remota" no relatório e ajuste da lógica de "Próximo comando sugerido" para considerar o achado. Independente das features 1 e 2 — pode ser feita em paralelo.

4. **Passo de clarify em `/nova-feature`, `/novo-fix` e `/discovery`**
   Criar uma regra única (novo arquivo em `sdd/memory/`, mesmo padrão de `naming-convention.md`) com heurística objetiva de "quando perguntar" (sinais: critério de aceitação ausente, escopo ambíguo, dependência externa não mencionada). Referenciar essa regra nos três corpos canônicos (`.agents/commands/nova-feature.md`, `novo-fix.md`, `discovery.md`), inserindo o passo antes do PASSO 1 (branch) ou antes da produção dos artefatos de discovery. Independente das demais features deste pacote.

5. **Confirmação obrigatória de delegação a subagente no lifecycle**
   Atualizar o protocolo `PLAN` (em `CLAUDE.md`/`GEMINI.md`/chatmode Copilot ou ponto único equivalente) para perguntar explicitamente, antes de cada próxima atividade/comando, se deve ser delegada a subagente — com critério documentado e fallback "não aplicável" para ferramentas sem esse conceito. Independente das demais; maior superfície de comportamento (afeta todo comando, não um ponto isolado), então validar com mais cuidado antes de fechar.

6. **Documentação e release**
   Atualizar README/release notes citando a correção de nome do adaptador Claude, a nova etapa de `/status`, a nova regra de clarify e a nova pergunta de subagente — respeitando Regra 12 da Constituição (agente/comando identificado explicitamente).

## Observações para o Refinamento (`/split-features`)

- Features 1 e 3 podem entrar na mesma branch/PR se o revisor preferir (baixo risco combinado), mas feature 1 deve ser mesclada/validada antes da feature 2 (dependência direta de template).
- Feature 4 (clarify) e feature 5 (subagente) são as que mais precisam de validação prática (rodar sessões reais) antes de serem consideradas `done` — ambas mexem em quando o agente pergunta algo ao usuário, e o risco simétrico é o mesmo nas duas: perguntar demais vira fricção, perguntar de menos volta ao problema original. Podem ser a mesma feature (mesma preocupação de "quando interromper o usuário com uma pergunta") ou duas, a critério do revisor.
- Feature 6 fecha o pacote, só depois das demais estarem `done`.
- Confirmar com o usuário, antes de fechar as features 4 e 5, qual critério objetivo prefere para "quando perguntar" em cada caso (sempre perguntar vs. só acima de um limiar de ambiguidade/escopo) — este discovery não resolve essa escolha, só mapeia a necessidade.

**Handoff:** Arquivos gerados nesta discovery — `discovery-03-ergonomia-de-comandos-e-sincronizacao.md`, `criteria-03-ergonomia-de-comandos-e-sincronizacao.md`, `plan-03-ergonomia-de-comandos-e-sincronizacao.md` (em `sdd/discovery/`). Próximo passo: `/split-features`, organizando as features quebradas em `sdd/features/feat-03-ergonomia-de-comandos-e-sincronizacao/`.

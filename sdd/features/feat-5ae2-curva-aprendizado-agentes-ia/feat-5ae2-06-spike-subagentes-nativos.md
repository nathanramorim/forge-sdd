# Feature 5ae2-06 — Spike: Subagentes Nativos para Orquestrador/Builder/Revisor

Investigação técnica (não obrigatoriamente shippada nesta rodada) sobre substituir as "personas simuladas num único chat" por delegação real de tarefas, aproveitando primitivos nativos de subagentes já disponíveis em parte do ecossistema de IA agêntica.

## Critérios de Aceitação Executáveis

1. Documento de spike mapeando, para cada agente suportado (Claude, Gemini, Copilot), se existe hoje um primitivo nativo equivalente a "subagente com contexto isolado" (ex: `.claude/agents/*.md` no Claude Code).
2. Tabela de suporte por agente + recomendação objetiva de ir/não ir adiante com a migração, registrada como anexo desta task.
3. Não é necessário implementar a migração completa nesta feature — a saída é a decisão informada, não o código.

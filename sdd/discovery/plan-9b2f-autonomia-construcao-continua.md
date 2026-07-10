# Plano de Discovery 9b2f — Autonomia e Construção Contínua (Hardness Engineer)

Este plano descreve o roadmap preliminar de tarefas locais sugeridas para implementar e homologar a execução de desenvolvimento autônomo sem intervenção.

## Tarefas de Homologação Sugeridas

- [ ] **Tarefa 1:** Criar o script wrapper de automação local `scripts/sdd-daemon.sh` que faz a leitura cíclica do `sdd/memory/progress.md` e executa o binário do CLI ou chamadas de API.
- [ ] **Tarefa 2:** Atualizar a skill do Orquestrador localmente (`.gemini/skills/orquestrador.chatmode.md` ou `.github/chatmodes/`) para verificar o arquivo flag `.sdd-auto-pilot` e pular a parada de confirmação da fase PLAN.
- [ ] **Tarefa 3:** Configurar credenciais e tokens da API dos agentes (Claude API `ANTHROPIC_API_KEY` ou Gemini API `GEMINI_API_KEY`) no ambiente local para permitir chamadas não interativas via CLI.
- [ ] **Tarefa 4:** Criar uma fila de 2 features de teste simples (ex: adicionar comentários ou pequenos testes) em `progress.md` e validar o daemon executando-as em sequência sem intervenção humana.

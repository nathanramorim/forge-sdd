# Status — Iniciativa 91

**Branch:** `feat/91-selecao-de-modelo-e-respostas-por-menu` (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** discovery, critérios e plano prontos; aguardando o Gate 1.
**Próximo passo:** aprovar e rodar `/split-features` (10 tasks em 3 ondas). Antes do build, publicar a beta com a 90.

## A confirmar (premissas assumidas sem perguntar)
- "Modelo por agente": cada agente (claude, agy, gemini) tem lista e tabela próprias; no chat vale também por papel.
- Sem menu disponível, o agente segue a opção recomendada e registra aqui.
- O Reviewer fica fora do roteamento.
- Um job por task é opcional e nasce desligado.
- O `/toca-ai` troca a branch por etapa pela branch única (Regra 15).
- Níveis `leve`, `medio`, `forte`; campo `**Nível:**` na spec.
- No build: nome da ferramenta de pergunta e do campo `model:` de comando; flags de `codex` e `copilot`.

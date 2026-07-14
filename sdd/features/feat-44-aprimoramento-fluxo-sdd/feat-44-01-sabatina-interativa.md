# Feature 44-01 — Sabatina Interativa ao Final do Fluxo de Planejamento

**Branch:** `feat/44-01-sabatina-interativa`
**Depende de:** —
**Paralelizável (worktree):** Não — edita os mesmos arquivos de prompt tocados por 44-03; rodar sequencial evita conflito.

## Descrição
Hoje `/discovery`, `/split-features` e `/nova-feature` aceitam a demanda do usuário e produzem artefatos sem contestar suposições. Esta feature adiciona uma sabatina sempre ativa ao final de cada um desses três comandos, com opção explícita de pular.

## Critérios de Aceitação Executáveis

1. Ao final de `/discovery`, `/split-features` e `/nova-feature` (Claude e Gemini), o agente apresenta as decisões relevantes tomadas na execução (escopo, paralelismo sugerido, branch de destino) e pergunta se o usuário confirma, quer pular (aceitar a sugestão) ou quer sobrescrever com sua própria decisão.
2. A resposta do usuário (confirmar/pular/sobrescrever) fica registrada no artefato gerado (`plan-ID-*.md`, spec de feature ou spec de fix) — não se perde ao final da sessão.
3. Responder "pular" nunca bloqueia a conclusão do comando: o fluxo segue com a sugestão default sem esperar detalhamento adicional.
4. Nenhuma pergunta de sabatina é feita antes de o agente ter uma sugestão concreta para apresentar — a sabatina não substitui o trabalho de análise, apenas o valida com o usuário.

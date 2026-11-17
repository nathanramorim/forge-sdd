# Discovery 89 — Comando `/toca-ai`

## Problema
Para deixar um trabalho longo rodando sem acompanhar (caso real: produzir todos os vídeos do `reel-studio`), era preciso colar um prompt enorme ("Goal set: …") e, mesmo assim, o agente pedia interação no meio (rodar comando no terminal, liberar permissão, confirmar consentimento). O Líder também não digita `/clear` nem `/goal clear` no chat.

## Para quem
Quem usa o Forge e quer informar só o objetivo, continuar conversando no chat (`/rc`) e voltar quando acabar.

## Jornada
1. `/toca-ai <objetivo>`: o Líder gera todas as specs e o `requisitos.json`.
2. Pré-voo: uma lista única do que falta (permissões, ferramentas, arquivos, consentimentos); a pessoa responde uma vez.
3. O motor (`forge pipeline toca-ai`) enfileira as etapas; Builder e Reviewer executam em contexto limpo; o chat fica livre.
4. Imprevisto: a etapa vira `blocked`, as independentes seguem; no fim, uma lista única de pendências e o relatório final.

## Regras de negócio
- **RN-1** O agente nunca manda a pessoa rodar comando nem liberar permissão no meio da execução.
- **RN-2** Consentimento só vale com resposta humana gravada; o agente nunca o confirma.
- **RN-3** Specs de todas as etapas existem antes da execução; nada fica só na conversa.
- **RN-4** Limpeza de contexto, archive e parada são do Go, não do modelo.
- **RN-5** Merge em `main`, segredo, gasto externo e exceção a regra continuam exigindo a pessoa.

## Escopo
**Dentro:** motor `forge pipeline toca-ai` (+ `--preflight`, `consent`), comando `/toca-ai` e skill, docs.
**Fora:** apelido do comando; mudar o modo avançado; subcomando de nível raiz.

## Riscos
Etapa sem PR no fluxo do Reviewer (spike na 89-01); sem tmux não há `/clear` automático (o worker headless já limpa por job); surpresa imprevisível ainda bloqueia uma etapa.

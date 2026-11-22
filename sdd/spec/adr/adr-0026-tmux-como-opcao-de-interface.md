# ADR-0026: tmux como opção de interface, não como padrão nem requisito

- **Data:** 2026-11-22
- **Status:** `proposto`
- **Regras afetadas (IDs da constitution):** `Nenhuma` (nenhum ID criado, alterado ou removido; `ARQ-06` não ganha subcomando)
- **Autor / Responsável:** Líder (decisão do dono do produto, 22/11/2026: "o tmux continua opcional")
- **Substitui em parte:** ADR-0006 (D-6), ADR-0007 (D-3) e ADR-0008 (D-3 desta: o modo `live` passa a exigir `terminal: tmux` explícito). Alinha-se à ADR-0015 (o chat orquestra subagentes; a 0015 diz que a 0008 "convive", e convive: a 0026 só muda o que é padrão e o que é opcional) e à ADR-0024 (item 5: tmux é opção de interface).

## Contexto

As ADR-0006, 0007 e 0008 tornaram o tmux a interface natural do pipeline: `forge up` criava os painéis (0006 D-6), oferecia instalar o tmux sozinho (0007 D-3) e o modo `live` injetava prompts nos painéis por buffers do tmux (0008). Depois vieram a ADR-0015 (no chat o trabalho vai a subagentes, sem `forge up`), a ADR-0024 e a feat-115-02, que **removeu** o tmux do código; a feat-115-04 (#231) o **devolveu como opção**. Hoje há três situações reais: quem trabalha só no chat, quem abre as 3 sessões em janelas do sistema e quem prefere painéis em tmux. Nenhuma ADR diz isso, e as três antigas ainda afirmam o tmux como caminho.

## Opções Consideradas

1. **tmux como padrão (0006/0007/0008 como estão):** acompanha a visibilidade original, mas obriga a instalar e confunde quem trabalha no chat.
2. **Remover o tmux de vez (115-02):** simplifica o código, mas tira de quem já usa painéis e o modo `live`.
3. **tmux opcional (escolhida):** mantém o ganho para quem quer, sem exigir.

## Decisão

- **D-1 Padrão:** janelas do sistema. Nenhum comando exige tmux.
- **D-2 Opt-in explícito:** `--terminal tmux` ou `terminal: tmux` no `forge.json` fazem o `forge up` abrir as 3 sessões em tmux. Sem o tmux no PATH, o `forge up` avisa e usa as janelas, sem erro (se o tmux existe e a montagem da sessão falha, o erro é mantido). `terminal: auto` e `native` equivalem às janelas do sistema. `forge status --tmux` abre só o painel vivo e, sem tmux, devolve erro com o comando de instalação.
- **D-3 Modo `live` só com tmux:** `session_mode: live` vale apenas com `terminal: tmux`; fora disso cai para `headless` nas janelas do sistema, com aviso. O `Preflight` só cobra o tmux quando ele foi escolhido.
- **D-4 Sem instalação automática (substitui 0007 D-3):** o Forge **não executa** a instalação do tmux; mostra o comando do gerenciador de pacotes para a pessoa rodar.
- **D-5 Permanece:** o restante das ADR-0006 (LiveView, máscara de segredos, `Requeue`, pausa por limite, "a Discovery não executa"), da 0007 (`brief`, `/constitution`; o Windows Terminal, `wt.exe`, é caminho nativo e independe do tmux) e da 0008 (despachante, `VerifyJobHandoff`, `capture-pane`) vale **quando o tmux está ativo**.
- **D-6 Chat:** quando o trabalho é no chat, nada disso se aplica; o chat orquestra subagentes (ADR-0015).

## Consequências

- **Positivas:** quem usa só o chat ou janelas não instala nada; quem usa tmux mantém painéis e `live`; o contrato fica escrito e testável (`TestTmux*`).
- **Negativas / Trade-offs:** dois caminhos de interface para manter e testar; a instalação assistida do tmux deixa de existir.
- **Ações de acompanhamento:** marcar nas ADR-0006, 0007 e 0008 que foram substituídas em parte por esta (feito neste PR); revisar as menções ao tmux nas skills de status (pendência do `status.md` da 115); promover esta ADR a `aprovado` com o `[Seguir]` do dono.

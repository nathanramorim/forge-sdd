# Discovery 77 — Sessões com a CLI Aberta (Builder e Reviewer ao vivo)

## Problema
Hoje os painéis `forge-builder` e `forge-reviewer` rodam o worker do Forge. Ele fica em "Aguardando jobs na fila..." e só abre o agente, sem tela (headless), quando chega um job. A pessoa quer o contrário: **a CLI do agente já aberta em cada painel desde o início**, como na Discovery, vendo o agente trabalhar na própria tela e descobrindo de cara um problema de login ou de plano (caso real: o Gemini falhou na autenticação só quando o primeiro job chegou).

## Para quem
Quem abre `forge` e acompanha Builder e Reviewer trabalhando, podendo intervir (responder uma pergunta, ler o raciocínio) sem abrir logs.

## Jornada alvo
1. `forge` abre 3 painéis no tmux: Discovery, Builder e Reviewer, cada um com a **CLI do agente interativa** aberta e pronta (login e plano conferidos pelo próprio agente na partida).
2. `/split-features` + despacho colocam um job na fila. Um despachante (processo do Forge, fora da tela) pega o job, **entrega o prompt no painel do Builder** como se a pessoa tivesse digitado, e acompanha até o agente registrar o PR.
3. O Builder trabalha na tela; o job do Reviewer entra e o painel do Reviewer faz o mesmo. Devolução, escalonamento e aviso no `/inbox` continuam iguais.
4. Entre jobs o contexto do agente é limpo (padrão), para cada job começar isolado.
5. Sem tmux (Windows nativo, janelas do sistema): continua o worker headless de hoje, com uma linha dizendo por quê.

## Regras de negócio
- A CLI aberta é **opcional**: `session_mode: live` (padrão quando há tmux) ou `headless` (o de hoje). `forge.json` antigo sem o campo continua headless até a pessoa escolher no menu.
- **Isolamento por job é o padrão:** o despachante manda o comando de limpar contexto da CLI antes de cada job; quem preferir contexto contínuo escolhe `context: continuous` e vê um aviso de custo.
- O job só termina por **estado**, não por saída do processo: Builder, quando o PR está registrado; Reviewer, quando o veredito está aplicado (as mesmas verificações de hoje). Sem isso, vale o timeout do job.
- Limite de uso, agente indisponível e inatividade são lidos da **tela do painel** (`tmux capture-pane`), com as mesmas regras do worker headless.
- O despachante nunca digita em painel ocupado: só entrega com o painel ocioso.
- A pessoa pode digitar no painel durante um job; o Forge avisa no guia que isso mistura com o trabalho do agente.

## Escopo
**Dentro:** modo `live` com tmux; perfis com argumentos interativos para Builder e Reviewer; despachante de entrega (`worker --deliver`); janela oculta com os despachantes; detecção de ocioso, conclusão e falha pela tela; limpeza de contexto por job; menu, `up`, preflight e docs.
**Fora:** CLI aberta sem tmux (não há como digitar em janela do sistema de forma portátil); troca de agente automática; reabrir a CLI sozinho se a pessoa fechá-la (o despachante avisa e para); mexer no worker headless além do necessário.

## Riscos
| Risco | Mitigação |
|---|---|
| Entregar prompt no momento errado (agente ainda carregando ou respondendo) | Painel ocioso = tela estável por N segundos e prompt de entrada visível; espera inicial; teste com agente falso |
| Contexto acumulado entre jobs | Limpeza por job como padrão; `continuous` explícito |
| Comando de limpar diferente em cada CLI | Campo `reset_command` por perfil; CLI sem comando conhecido fica `continuous` com aviso (a confirmar agy e copilot) |
| Pessoa e despachante digitando juntos | Só entregar em painel ocioso; documentar |
| Prompt longo com aspas e quebras de linha no `send-keys` | Entrega por `load-buffer` + `paste-buffer` (colagem entre colchetes), não por `send-keys` do texto |
| Mudança contraria a ADR 0004 (worker headless) | Nova ADR; o headless continua como modo e como contingência |

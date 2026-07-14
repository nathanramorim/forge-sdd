# Feature 44-02 — Script Mecânico de Git Worktree

**Branch:** `feat/44-02-worktree-script`
**Depende de:** —
**Paralelizável (worktree):** Sim — arquivo novo e isolado (`sdd/scripts/worktree.sh`), sem overlap com 44-01/44-03/44-04.

## Descrição
Tasks marcadas como paralelizáveis em `plan-ID-*.md` devem poder rodar em worktrees git isoladas, sem que o agente gaste tokens compondo comandos git manualmente. Esta feature adiciona um script mecânico de abrir/fechar worktree, acionável tanto pelo orquestrador (como comando sugerido) quanto diretamente pelo usuário no chat.

## Critérios de Aceitação Executáveis

1. `sdd/scripts/worktree.sh open <branch>` cria a branch (se não existir) e uma worktree em `.worktrees/<branch>` via `git worktree add`.
2. `sdd/scripts/worktree.sh close <branch>` remove a worktree correspondente (`git worktree remove`) e, se a branch já estiver mesclada na branch de release ativa ou em `main`, oferece removê-la também.
3. O script é idempotente: rodar `open` duas vezes para a mesma branch não corrompe o estado (detecta worktree já existente e apenas informa o caminho); rodar `close` numa worktree inexistente falha com mensagem clara, não silenciosamente.
4. O orquestrador (`orquestrador.chatmode.md`, Claude e Gemini) passa a sugerir a invocação deste script como comando de shell pronto (não lógica gerada) quando o `plan-ID-*.md` da feature ativa marca tasks como paralelizáveis, e sugere `close` ao concluir a implementação completa de uma task paralela, antes/junto da criação do PR.
5. Um comando de chat (`/worktree open|close <branch>`) fica disponível para acionar o script diretamente, para os casos em que o usuário quer abrir/fechar sem esperar a sugestão do orquestrador.

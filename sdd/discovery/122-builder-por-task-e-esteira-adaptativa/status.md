# Status — 122 Builder por task e esteira adaptativa

**Gate 1:** aprovado pelo Nathan em 24/11/2026 23h ("vamos rodar o split"). Alvo: 6.0.0-beta.
**Status:** implementada na 6.0.0-beta (feat-122-01..09): task_build, worktree efêmero no worker, despachante contínuo, par Reviewer, junção por squash, retomada, status por task. Build, vet e testes verdes.
**Próximo passo:** revisão do PR pelo dono; publicação no npm só com o ok dele.
**Decisões (24/11, dono):** `task` padrão na beta; sem teto fixo, limite por disco 80% e memória com reserva de 2 GB; janela 85%; squash no PR da feature.
**Limitação:** modo avançado segue por feature. **A confirmar:** medição real nos dois modos (3.3) depois da beta.

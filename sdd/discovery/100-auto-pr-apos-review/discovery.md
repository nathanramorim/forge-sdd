# Discovery — 100-auto-pr-apos-review

## Problema

Após `forge queue review` retornar veredito `conforme`, o agente precisava perguntar ao usuário para abrir o PR — ação repetitiva e desnecessária. O PR é a etapa natural de conclusão; não deve depender de intervenção humana.

## Solução

Estender `forge queue review` para, quando o veredito for `conforme`, executar automaticamente `git push origin <branch>` + `gh pr create`. Flag `--no-pr` disponível para pular quando necessário.

# Comando: status

**Uso:** `/status` mostra o status executivo da esteira em formato tríplice e de baixo consumo de tokens (< 200 tokens).

Rode `forge status --diagnostico` primeiro: um comando só responde workers vivos, fila por papel, job em execução, última falha com motivo e log, pendentes velhos e o modo do Builder. Não liste pastas de `.forge/runtime` nem rode `ls`/`cat` na fila; se o diagnóstico apontar um log, leia só a cauda dele (`tail -n 40 <log>`). Versão do binário: `forge --version`.

Depois, se precisar do restante do template:

1. **Ao vivo:** a cauda do log do job em execução que o diagnóstico mostra.
2. **Fila de tarefas:** `forge status` (traz a esteira por task quando houver).
3. **Commits recentes:** `git log --oneline -n 5`.

## Template Canônico Padronizado (< 200 Tokens)

Responda obrigatoriamente no seguinte formato rígido:

```text
### Status da Esteira — <projeto> (<branch-atual>)

1. Ao Vivo no Painel (<papel-ativo>):
   • <linha 1 do que o agente está fazendo>
   • <linha 2: tempo ativo e arquivos alterados>

2. Commits Recentes no Git:
   • <hash1> <mensagem do commit recente>
   • <hash2> <mensagem do commit anterior>

3. Fila de Tarefas:
   • [x] <task-id>: <titulo da task pronta>
   • [ ] <task-id>: <titulo da task pendente>

Próximo passo: <Acompanhar o Builder / Revisão do Reviewer / /discovery>
```

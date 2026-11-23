# ADR-0027 — Subcomando oculto `hook pr-guard`: fail-open sem script ou sem uv

- **Data:** 2026-11-22
- **Status:** `aprovado` (aprovada pelo dono em 22/11/2026)
- **Regras afetadas (IDs da constitution):** `ARQ-06` (acrescenta `hook` à lista de subcomandos), `GOV-01` (esta ADR). Nenhum ID é criado ou removido.
- **Autor / Responsável:** Iniciativa 106, fix-106-14b.

## Contexto

O hook `PreToolUse` do `pr_guard` chamava `uv run scripts/pr_guard.py --hook`. Se o script não existisse (apagado, branch antiga, checkout parcial), o `uv` falhava com código 2, que o Claude Code trata como bloqueio (fonte: https://code.claude.com/docs/en/hooks, tabela de códigos de saída: exit 2 é erro bloqueante e a chamada `PreToolUse` continua bloqueada), e todo comando Bash ficava barrado. O fix-106-14a passou a entregar o script antes do hook; resta o caso de o arquivo sumir depois. Um comando de shell (`[ -f ... ] && ... || exit 0`) não serve: é POSIX, e `A && B || C` também sai 0 quando o guard devolve 2, mascarando o bloqueio legítimo.

## Opções Consideradas

1. **Comando de shell com teste de existência:** rejeitada; não é portável (Windows) e mascara o código 2 do guard.
2. **Manter o comando e confiar no 14a:** rejeitada; qualquer projeto que perca o script volta ao bloqueio total.
3. **Wrapper em Go como subcomando oculto (escolhida).**

## Decisão

1. Novo subcomando oculto (`Hidden: true` do cobra): `forge hook pr-guard`. A lógica vive em `internal/hook`; o handler em `cmd/` só adapta (`ARQ-01`).
2. Resolve a raiz do projeto por `CLAUDE_PROJECT_DIR`, senão `git rev-parse --show-toplevel`, senão o cwd.
3. **Fail-open consciente:** sem `scripts/pr_guard.py`, ou sem `uv` no PATH, avisa em stderr e sai 0. O job `pronto?` do CI é o apoio.
4. **Bloqueio real preservado:** com os dois presentes, executa `uv run scripts/pr_guard.py --hook`, repassa stdin, stdout e stderr e devolve o mesmo código de saída (o 2 do guard permanece).
5. **Migração:** `init` e `update` gravam `forge hook pr-guard`; o merge substitui a entrada antiga (`uv run scripts/pr_guard.py --hook`) no lugar, sem duplicar e sem tocar hooks do usuário nem JSON inválido. O `forge doctor` reconhece os dois comandos e só acusa erro de configuração para o comando antigo sem script.
6. **Constituição:** `ARQ-06` ganha `hook` na lista, com a remissão a esta ADR; o trecho "(`modulo` chega na 109.)" foi removido para caber no teto `MaxConstitutionSize` (P-24).

## Consequências

- **Positivas:** o hook nunca bloqueia o Bash por arquivo ausente; portável em Windows, macOS e Linux; o bloqueio legítimo continua valendo.
- **Negativas / Trade-offs:** sem `uv` ou sem o script a barreira local fica inativa (só o CI segura); depende de o `forge` estar no PATH, como os hooks do doctor e do inbox; se o guard morrer por sinal, o wrapper devolve 0 com aviso (fail-open); sem `CLAUDE_PROJECT_DIR`, a raiz vem do `git rev-parse --show-toplevel`, que num projeto dentro de outro repositório pode apontar para a raiz errada e silenciar o guard (o Claude Code sempre define a variável).
- **Ações de acompanhamento:** nenhuma; a ADR-0025 não precisa de linha nova.

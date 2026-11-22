# ADR-0018: Gravação Segura de Arquivos do Usuário (`checkSafeTarget` e `writeFileSafe`)

- **Data:** 2026-11-20
- **Status:** `aceito` (20/11/2026, aprovado pelo dono no chat)
- **Regras afetadas (IDs da constitution):** `SEG-02` (texto ampliado em 20/11/2026), `SEG-01`, `GOV-01` (esta ADR). Nenhum ID é criado ou removido.
- **Autor / Responsável:** Concepção; decisão do dono pendente

## Relação com ADRs vigentes
Nenhuma ADR anterior trata de escrita em arquivo do usuário; `SEG-02` já exige respeitar o diretório-alvo e preservar arquivos de domínio.

## Contexto
`init`, `update` e a integração com Claude (`.claude/settings.json`, hook e MCP) escrevem em arquivos que pertencem à pessoa. Link simbólico ou pai simbólico podia levar a gravação para fora do projeto. Um fix separado cobre também `renderDir` (`internal/scaffold/scaffold.go`), na branch `fix/104-renderdir-symlink` (commit `ab61c58`), **ainda não mesclada**.

## Opções Consideradas
1. **A. Validar o alvo e gravar por temporário e rename.** 2. **B. Gravar direto** (estado anterior). 3. **C. Seguir links** e conferir só o resultado.

## Decisão
- **D-1:** adotar **A** em `internal/scaffold/hooks.go`: `checkSafeTarget` recusa alvo que seja link simbólico ou não regular, e pai simbólico ou que resolva fora do projeto; ausentes são aceitos.
- **D-2:** `writeFileSafe` revalida logo antes, grava em temporário exclusivo (`O_EXCL`, 0600) no mesmo diretório e faz rename, preservando o modo existente (novo: 0644).
- **D-3:** `renderDir` recusa destino simbólico ou fora do projeto (só na branch do fix; não está nesta branch).

## Consequências
- **Positivas:** sem escrita fora do projeto nem arquivo parcial. **Negativas:** quem usa link simbólico em `.claude` recebe recusa explícita.
- **Ações de acompanhamento:** mesclar o fix do `renderDir`; ver proposta sobre `SEG-02` em `constitution-propostas.md`.

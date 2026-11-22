# ADR-0010: `forge` como Único Nome de Comando

- **Data:** 2026-11-14
- **Status:** `aceito` (2026-11-20, aprovado pelo dono no Gate 1 da 105)
- **Regras afetadas (IDs da constitution):** `ARQ-06`, `DUP-01`, `COD-05`, `GOV-01` (este ADR; nenhum ID foi criado ou removido)
- **Autor / Responsável:** Líder (discovery 82)

## Contexto

O usuário aprendia a usar `forge` (menu, `forge init`), mas o próprio CLI mostrava `Usage: forge-sdd status`, dicas com `forge-sdd ps --kill-zombies` e erros com o nome antigo. Prompts de agentes, rules, skills e templates de novos projetos repetiam `forge-sdd <cmd>`. A dualidade confundia pessoas e agentes. Referência: `sdd/discovery/82-unificacao-comando-forge/`.

## Opções Consideradas

1. **Manter os dois nomes:** sem custo, mas a ambiguidade continua.
2. **Remover o alias `forge-sdd`:** quebra quem já o tem instalado.
3. **`forge` como nome oficial e `forge-sdd` como alias:** consistente na documentação e compatível com instalações existentes.

## Decisão

- **D-1:** `rootCmd.Use = "forge"`; mensagens de uso, dicas e erros citam `forge`.
- **D-2:** templates do scaffold, prompts (`CLAUDE.md`, `GEMINI.md`), rules e skills ensinam `forge <cmd>`; os golden tests acompanham.
- **D-3:** o pacote npm mantém os dois binários (`forge` e `forge-sdd`, o mesmo `bin/run.js`) para não quebrar instalações. O nome do pacote e do módulo Go não mudam.

## Consequências

- **Positivas:** um nome só para pessoas e agentes; menos tokens gastos em desambiguação; scaffold já nasce padronizado.
- **Negativas / Trade-offs:** projetos já scaffoldados ainda têm `forge-sdd` nos arquivos até o próximo `forge update`.
- **Ações de acompanhamento:** novos textos devem usar `forge`; o diretório `cmd/forge-sdd` e o nome do pacote continuam como estão.

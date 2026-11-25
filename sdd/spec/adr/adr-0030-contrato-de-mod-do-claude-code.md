# ADR-0030 — Contrato de mod do Claude Code (manifesto, permissões, versões mínimas e carregamento)

- **Status:** Proposta (aguarda aprovação do dono no Gate da iniciativa 123)
- **Data:** 2026-11-25
- **Regras afetadas (IDs da constitution):** `FLX-05` e `GOV-01` (esta ADR). Nenhum ID é criado, alterado ou removido.
- **Autor / Responsável:** Builder da feat-123-01 (iniciativa `123-painel-de-execucao-no-claude-code`).

## Relação com ADRs vigentes

- **ADR-0025 (instalação limitada ao projeto): estende.** O mod mora em `.agents/plugins/<slug>/` e entra por `forge skills instalar plugin`. `/plugin install` segue rejeitado.
- **ADR-0003 e `DUP-10` (fonte única em `.agents/`): convive.** Nada do mod é copiado para `.claude/`.
- **ADR-0026 (tmux opção de interface): convive.** O modo painel (feat-123-04) é mais uma opção ao lado de janelas e tmux.
- **ADR-0009 (tokens por janela): estende.** Com o mod ativo, a status line substitui o rótulo de tokens por painel.
- **Módulos externos (109): alinha o vocabulário.** Manifesto, permissões declaradas e versão mínima; o mod não é binário.

## Contexto

Um **mod** é uma extensão do Claude Code que muda a interface ou o comportamento dele: status line, painel lateral, alerta, comando e ganchos de sessão, turno e subagente. A API é de acesso antecipado e muda por versão. O forge precisa de um contrato único para os mods que instala: o que o mod declara, o que o instalador recusa e como o Claude Code o carrega sem sair do projeto. O `forge-painel` (iniciativa 123) é o mod de referência.

Verificado no Claude Code 2.1.332 (documentação da 2.1.331 e execução local):

- O manifesto `.claude-plugin/plugin.json` aceita os campos do Claude Code. Um campo desconhecido gera aviso do `claude plugin validate`, e uma chave `hooks` aninhada em campo desconhecido vira erro. O contrato do forge não cabe, portanto, dentro do `plugin.json`.
- O módulo de ganchos (`hooks/hooks.json` com `modules`) é código do mod: não tem DOM nem Node e só alcança o mundo por `$` (`$.fs`, `$.env`, `$.ui`, `$.session`, `$.agent`...). `claude plugin validate` lista os ganchos, as chamadas em `$`, as variáveis de ambiente lidas e o estado usado.
- Carregamento sem `/plugin install` (spike, `status.md` da 123): `claude --plugin-dir <pasta>` carrega o plugin só naquela sessão, inclusive em `claude -p`; `CLAUDE_CODE_PLUGIN_DIRS` faz o mesmo onde não há flag (app desktop, SDK), lida do ambiente do processo ou do bloco `env` de `~/.claude/settings.json`, nunca de configuração do projeto. Uma pasta em `.claude/skills/<slug>` do projeto não foi carregada na 2.1.332 (testado).

## Opções Consideradas

- **A. Contrato dentro de `plugin.json`:** rejeitada. O Claude Code ignora campos desconhecidos com aviso, e a chave `hooks` de permissões foi lida como gancho.
- **B. Instalar pelo marketplace (`/plugin install`):** rejeitada pela ADR-0025 (grava em `~/.claude`).
- **C. Arquivo de contrato do forge ao lado do manifesto, carregamento por `--plugin-dir` (escolhida).**

## Decisão

1. **Onde:** `.agents/plugins/<slug>/`, TypeScript executado pelo Claude Code, fora do binário Go (`ARQ-04`). O binário só copia e valida arquivos.
2. **Dois arquivos de manifesto:**
   - `.claude-plugin/plugin.json`: campos do Claude Code. Obrigatórios para o forge: `name` igual ao slug, `version` em `X.Y.Z` e `description` não vazia.
   - `.claude-plugin/forge.json`: o contrato do forge, lido só pelo forge.

   ```json
   {
     "schema": "forge-sdd/mod/1.0",
     "minClaudeCode": "2.1.331",
     "minForge": "5.2.3",
     "permissions": {
       "network": false,
       "write": ["sdd/.metrics", ".forge/runtime/painel"],
       "hooks": ["session.measure", "agent.report"],
       "read": ["sdd/.metrics"],
       "env": ["FORGE_WORKER"],
       "model": false
     }
   }
   ```

   Obrigatórios: `schema`, `minClaudeCode` e `minForge` (`X.Y.Z`), `permissions.network` (booleano), `permissions.write` (lista de caminhos relativos ao projeto) e `permissions.hooks` (lista não vazia de eventos). Opcionais: `read`, `env` (listas de texto) e `model` (booleano; chamar modelo é permissão declarada).
3. **Quem é mod:** o plugin que traz módulo de ganchos (`hooks/hooks.json` com `modules`) ou o arquivo `forge.json`. Plugin só de skills e comandos (como o `diagram-studio`) segue a ADR-0025 sem contrato.
4. **O que a instalação recusa** (`forge skills instalar plugin`, antes de gravar o primeiro arquivo): mod sem `forge.json`; `name` diferente do slug; versão fora de `X.Y.Z`; `schema` desconhecido; versões mínimas ausentes ou malformadas; `minForge` maior que o forge em uso (versão de desenvolvimento não é comparada); permissão obrigatória ausente ou de tipo errado; caminho de escrita ou leitura absoluto, com `..`, com `~` ou dentro de `.git`, `.claude` ou `.agents`; módulo de ganchos que `hooks.json` cita e o plugin não traz.
5. **Comportamento (vale para todo mod do forge):** sem rede salvo `network: true` declarado; escrita só nos caminhos de `write`; um dono por registro de telemetria (o mod fica mudo quando `FORGE_WORKER` existe); dólar sempre rotulado "custo de tabela"; nenhum campo novo no schema `forge-sdd/metrics/1.0` sem ADR (`DUP-04`); texto de agente na tela passa por saneamento (`SEG-08`).
6. **Carregamento:** o forge abre o Claude Code com `--plugin-dir .agents/plugins/<slug>`; onde não há flag, `CLAUDE_CODE_PLUGIN_DIRS` no ambiente. O forge nunca escreve em `~/.claude` para isso: quando só o ambiente resolve, ele imprime o comando.
7. **Teste:** todo mod traz `*.test.tsx` rodado por `claude plugin test` e passa em `claude plugin validate`; o resultado entra na revisão da feature.

## Consequências

- **Positivas:** um contrato verificável pelo instalador, sem rede e sem gravar fora do projeto; o `forge.json` evita aviso do validador; o mod de referência já cumpre o contrato.
- **Negativas / trade-offs:** dois manifestos para manter; o contrato é declaração, e o instalador não prova que o código a cumpre.
- **Riscos:**
  - A API de mods muda por versão: `minClaudeCode` e o teste a cada release do Claude Code.
  - O módulo de ganchos depende de um sinal de liberação do Claude Code que chega no início do turno; o mod pode não estar ativo antes do primeiro turno.
  - Declarado e executado podem divergir; a saída de `claude plugin validate` (ganchos, chamadas e variáveis lidas) é o que se compara na revisão.
  - Carregar pelo projeto sem `--plugin-dir` (configuração do projeto, desktop) não foi provado: a confirmar no desktop do dono.
- **Ações de acompanhamento:** feat-123-03 (validação na instalação); feat-123-04 (`forge up --terminal painel` imprime o comando com `--plugin-dir`); confirmar o carregamento no app desktop; avaliar conferir `claude plugin validate --json` contra `permissions` no CI.

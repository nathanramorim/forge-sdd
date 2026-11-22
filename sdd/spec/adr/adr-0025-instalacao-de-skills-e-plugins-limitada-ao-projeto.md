# ADR-0025 — Instalação de skills e plugins limitada ao projeto

- **Status:** Aceita
- **Data:** 2026-11-22
- **Regras afetadas (IDs da constitution):** `ARQ-06` (só o texto do grupo `skills`: dois subcomandos novos), `GOV-01` e `FLX-05` (esta ADR). Nenhum ID é criado ou removido.
- **Autor / Responsável:** Concepção (iniciativa `98-diagram-studio-c4`, feat-98-02); aprovada pelo dono (22/11/2026).

## Relação com ADRs vigentes

- **ADR-0003 (skills por estação, fonte única `DUP-10`): estende.** Skills instaladas moram em `.agents/skills/<slug>/`, a fonte única; o plugin mora em `.agents/plugins/<slug>/`. Nada é substituído.
- **ADR-0024 (orquestração por iniciativa): convive.** Na esteira, quem aciona a instalação é o hub; esta ADR só define onde e como se grava.
- **ADR-0015 (pipeline em subagentes): convive.** Segue o mesmo padrão de Constituição: a linha de `ARQ-06` muda no PR da implementação, não aqui.
- **Constituição:** muda só o texto de `ARQ-06`. Como exige `GOV-01`, a edição entra no PR da feat-98-02; esta ADR não a faz.

## Contexto

A feat-98-01 entrega a skill `diagramas-c4-mermaid` e depende do plugin `diagram-studio`. Instalá-los pelo caminho padrão do Claude (`/plugin install`) grava em `~/.claude/`, fora do projeto, e vale para todos os projetos do usuário. Além disso, projetos com hub e subprojetos (cada um com `sdd/.sddrc`) poderiam receber a mesma cópia várias vezes.

Verificado no código: `internal/scaffold/hooks.go` já tem `checkSafeTarget` e `WriteFileSafe` (SEG-02); `internal/skills/skills.go` tem `ValidateSlug`; `shouldPreserve` em `internal/scaffold/scaffold.go` já preserva `.agents/skills/` (exceto `sdd-*`) e `.agents/rules/`, mas **não** `.agents/plugins/`; `cmd/forge-sdd/skills.go` tem `propose`, `approve` e `sync`, sem instalação de skill pronta. Não existe hoje noção oficial de hub no código.

## Opções Consideradas

- **A. Instalar em `~/.claude/` via `/plugin install`:** rejeitada. Grava fora do projeto, contra a política do dono, e não passa por `WriteFileSafe` nem por `shouldPreserve`.
- **B. Copiar sempre em todo subprojeto:** rejeitada. Duplica o conteúdo (contra `DUP-10`) e contradiz "o comando vem do hub".
- **C. Hub primeiro, herança de leitura, instalação só onde falta (escolhida).**

## Decisão

1. **Opcional por skill:** instala-se só o slug escolhido pelo usuário; nunca em lote, nunca implícito.
2. **Limitada ao projeto aberto:** grava só dentro do projeto em que o forge foi aberto, em `.agents/skills/<slug>/` e `.agents/plugins/<slug>/`. Nunca em `~/.claude/` nem fora do projeto. `/plugin install` não é usado.
3. **Hub e subprojetos:** o comando vem do hub. Se a skill ou o plugin já está disponível no projeto aberto ou no hub acima (o primeiro diretório ancestral com `sdd/.sddrc` que o tenha instalado), não se instala nos demais; a herança é só de leitura e o subprojeto não grava no hub. Só quando não está disponível se instala, e aí vale a pergunta hub ou subprojeto (destinos: projeto aberto mais diretórios filhos com `sdd/.sddrc`). Quem abre um subprojeto sozinho, sem hub que o tenha, instala nele.
4. **Segurança (SEG-01, SEG-02):** o destino é item da lista de destinos; o slug passa por `skills.ValidateSlug`; chaves de arquivo são relativas, sem `..` nem caminho absoluto; o caminho resolvido com `EvalSymlinks` fica dentro do destino; symlink é recusado; o que o usuário já tem é preservado (estender `shouldPreserve` para `.agents/plugins/`; `.agents/skills/` já é coberto); gravação atômica via `WriteFileSafe`.
5. **Mecanismo:** funções Go em `internal/scaffold` (`InstalarSkill`, `InstalarPlugin`, `ListarDestinosInstalacao`, `DisponivelViaHub`), retorno `Resultado{Gravados, Preservados, DisponivelEm}`. O no-op do hub vive no domínio; o handler em `cmd/` só repassa (`ARQ-10`). Chamador: dois subcomandos novos no grupo `skills` existente, sem comando de topo novo:
   - `forge-sdd skills destinos`
   - `forge-sdd skills instalar <skill|plugin> <slug> --de <dir> [--destino <dir>]`

   O Go não faz rede: o conteúdo vem do diretório de origem. O plugin `diagram-studio` é obtido fora do Go por `git clone` mais `checkout` do commit fixado `3996fd2e2bcdc2462e6f0d1edd7f6d18f649bc03`. O repositório não tem tags e o dono não quer tag criada.
6. **Alcance:** a política vale para todo skill ou plugin que o forge instale, não só o `diagram-studio`.

## Atualização de `ARQ-06` (feita no PR da feat-98-02)

Linha atual (`sdd/memory/constitution.md`, linha 62): `... `gate`, `skills`, `governance`, ...`

Texto aplicado: acrescentado antes de `(`modulo` chega na 109.)`, a oração "`skills`: ver ADR-0025." A lista de nomes de topo não muda. O contrato dos dois subcomandos novos (`skills destinos` e `skills instalar`, instalação opcional e limitada ao projeto aberto) fica nesta ADR, e não na Constituição: a Constituição tem teto de bytes (`MaxConstitutionSize`, P-24) e só cabe a remissão. O `FLX-02` e as demais regras não foram reescritos.

## Consequências

- **Positivas:** nada é gravado fora do projeto; sem duplicação entre hub e subprojetos; reuso de `WriteFileSafe` e `shouldPreserve`; política única para qualquer skill ou plugin futuro.
- **Negativas:** o usuário não ganha o plugin globalmente; o fluxo exige o clone externo ao Go; a herança só de leitura faz o subprojeto depender do hub estar presente.
- **Riscos:**
  - A noção de hub é inferida por `sdd/.sddrc` em diretório ancestral, sem definição oficial; um `.sddrc` inesperado acima do projeto altera o resultado de `DisponivelViaHub`.
  - Plugin fixado por commit sem tag: a rastreabilidade depende do hash, e o histórico do repositório externo pode ser reescrito.
  - `resvg-py>=0.1` é faixa aberta (SEG-07); o Builder da feat-98-02 fixa a versão exata.
  - `shouldPreserve` hoje não cobre `.agents/plugins/`; sem a extensão, uma reinstalação sobrescreveria o plugin do usuário.
- **Ações de acompanhamento:** feat-98-02 implementada com a linha de `ARQ-06` atualizada no mesmo PR; fixar `resvg-py` (a fixar: o `diagrama.py` do plugin declara `resvg-py>=0.1` e o commit fixado não o trava; fixar exige mudar o plugin externo); avaliar uma definição oficial de hub.

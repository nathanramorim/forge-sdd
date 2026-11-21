# ADR-0022: Política de PR em draft, modos do projeto, spec que volta à principal e branch por feature

- **Data:** 2026-11-20
- **Status:** `aceito` (20/11/2026, Gate 1 da iniciativa 106, aprovado pelo dono)
- **Regras afetadas (IDs da constitution):** `ARQ-06`, `FLX-01`, `FLX-02`, `FLX-05`, `FLX-06`, `FLX-07`, `GOV-01`, `GOV-03`, `GOV-04`
- **Autor / Responsável:** Concepção (iniciativa 106); decisão do dono no Gate 1 de 2026-11-20

## Contexto

O GitHub bloqueou as Actions da conta por cobrança e revelou o consumo evitável de rodar CI em trabalho ainda não conferido. Além disso, a spec ficava amarrada numa branch única de iniciativa até toda a implementação acabar (contradizendo a necessidade de escalar o backlog em branches independentes sem conflito de estado em `queue.json`), o auto-PR da iniciativa 100 nascia sem `--draft`, e o uso de `gh` estava fixo no código embora a Constituição já previsse `azure-devops` e `nenhum`.

Ao mesmo tempo, as iniciativas 106, 107 e 109 precisavam adicionar campos de configuração versionados na tabela `Ferramentas e Integrações` da Constituição (`Modo do atestado`, `Modo de revisão humana`, `WIP humano`, `Módulos permitidos` e `Início automático do Builder`) e o subcomando `modulo` em `ARQ-06`, mantendo `sdd/memory/constitution.md` dentro do orçamento de 9.216 bytes (`P-24`).

## Opções Consideradas

1. **Manter uma branch única por iniciativa com junção local e editar a Constituição em várias tasks (106/2.7, 107-12, 109):** gera conflito de merge entre PRs paralelos na Constituição e em `queue.json`, mantém o backlog amarrado a uma branch de iniciativa e estoura o orçamento de 9.216 bytes sem corte coordenado.
2. **PR único da Constituição (task 0.8 da 106) + spec que volta à principal + uma branch e um PR em draft por feature + rule `sdd-modos-de-projeto.md`:** concentra todas as alterações da Constituição em um único PR (`GOV-01`), migra o detalhamento operacional das Regras 11 e 12 e das notas de modos para `.agents/rules/` (`GOV-04`) e libera as branches por feature com estado isolado em `sdd/discovery/<id>/state/<sid>.json`.

## Decisão

Escolhida a **Opção 2** (RN-0 a RN-25 da iniciativa 106 e alinhamento com 107 e 109):

1. **Tabela `Ferramentas e Integrações` e rule `.agents/rules/sdd-modos-de-projeto.md`:**
   - Adicionados os campos `Modo do atestado` (`local`), `Modo de revisão humana` (`com-humano`), `WIP humano` (`3`), `Módulos permitidos` (`nenhum`) e `Início automático do Builder` (`desligado`), todos lidos da branch base.
   - As notas operacionais saem da Constituição para `.agents/rules/sdd-modos-de-projeto.md`, ficando na Constituição apenas o ponteiro (`GOV-04`).
2. **Fluxo de spec, split pelo Builder e branch por feature (Regra 15, `FLX-01`, `FLX-02`):**
   - **Regra 15 e `FLX-02`:** a spec vive numa branch por iniciativa (`<prefixo>/<id>-<nome>`), com PR próprio em draft, e volta para a principal após a Conferência 1 e o `[Seguir]`; a construção usa uma branch por feature do split (`<prefixo>/<id>-<nome>-<sid>`), aberta pelo Builder a partir da principal, cada uma com o seu PR em draft.
   - **`FLX-01`:** nada é implementado nem enfileirado sem Spec revisada e aprovada pelo humano no Gate 1 (`[Seguir]`/`[Ajustar]`); com o Gate 1 aprovado o Builder faz o split e enfileira sem nova aprovação, só dentro do escopo do `plan.md` aprovado.
3. **Subcomando `modulo` (`ARQ-06`) e lições por arquivo (`GOV-03`):**
   - `ARQ-06` inclui `modulo` na lista de subcomandos determinísticos de suporte (iniciativa 109).
   - `GOV-03` passa a consolidar lições em `sdd/memory/licoes/` (um arquivo por lição).
4. **Orçamento da Constituição (`P-24`) e migração das Regras 11 e 12 (`FLX-07`, `GOV-04`):**
   - As Regras 11 e 12 são removidas da lista numerada (mantendo a numeração 10, 13, 14, 15 para preservar referências) e `FLX-07` fica conciso na Constituição (`9.173` de `9.216` bytes).
   - Os detalhes operacionais de betas, `-rc`, falha com texto genérico, retenção das 3 estáveis mais betas posteriores no GitHub (sem afetar o NPM) e atualização de documentação com prévia e agente explícito migram para `.agents/rules/sdd-fluxo-e-portao.md`.

## Consequências

- **Positivas:** elimina o conflito entre tasks que dependem da Constituição; mantém `sdd/memory/constitution.md` dentro do teto de 9.216 bytes (`P-24`) e sem cópia literal nas rules (`P-23`); destrava a construção em paralelo por feature com PR em draft e CI apenas ao sair do draft.
- **Negativas / Trade-offs:** exige atualizar leitores de modos, fila por estado de feature (`state/<sid>.json`), hooks e templates nas tasks subsequentes da 106, 107 e 109.
- **Ações de acompanhamento:** tasks das ondas 0 a 3 da iniciativa 106 (leitor de modos `0.5`, cadeia de fallback `0.6`, scripts e fluxo `1.x` e `2.x`, ADR final dos fluxos na `2.7`).

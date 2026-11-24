# Discovery: 121 Claude Ship como módulo opcional do forge-sdd

## Pedido do dono (Nathan, 24/11/2026)
Descobrir como o **Claude Ship** se encaixa no fluxo do forge-sdd **sem mudar a mecânica da esteira**: uma nova forma de trabalho habilitada por configuração quando se instala uma extensão. Cobrir quem já usa o fluxo atual e quer ativar a extensão depois. Toda extensão nasce em **repositório novo e privado**; esta discovery só **propõe** o repositório (nome, estrutura, contrato), não o cria.

## Estacionada (dono, 24/11/2026)
Só faz sentido em projeto com plano **Team ou Enterprise** e usage credits. Em Pro/Max as três peças não existem e o forge já cobre o equivalente (Reviewer, retrabalho do Builder, Gate humano) sem custo extra. A spec fica registrada para reativação.

## Claude Ship em linguagem simples (leia antes do Gate 1)
**O que é:** um serviço da Anthropic que trabalha **depois que o PR é aberto no GitHub**. São três agentes que rodam na nuvem da Anthropic, mais um comando local.

| Peça | Em uma frase | Quem liga | Como é cobrada |
|---|---|---|---|
| Code Review (GA) | lê o PR e comenta achados por severidade; não aprova nem bloqueia | Owner da organização, por repositório | tokens, em usage credits da organização, **fora do uso incluído no plano** |
| PR Steward (beta) | faz commits no PR para corrigir CI falho, conflito e comentários de review | Owner, por repositório | tokens, em usage credits; **sem limite até o Owner definir** |
| Claude Approvals (beta) | aplica uma política de aprovação escrita em `.claude/approvals.yml` | quem mescla o arquivo no branch padrão | tokens das reviews e dos classificadores que a política chama; **observe também cobra**; teto de US$ 10 por execução |
| `/ship` (Claude Code local) | o dev manda: push, abre PR, revisa e acompanha | o próprio dev | conta no uso do plano do dev, não em usage credits |

**Cobrança em resumo:** não há preço fixo por PR nem preço por token publicado nas páginas lidas. O custo varia com o nível do review (Quick, Standard, Deep), o tamanho do PR, a frequência (`After every push` multiplica) e o trabalho do Steward. A recomendação da própria documentação é fazer piloto e olhar o custo médio em `claude.ai/analytics/code-review` e `claude.ai/analytics/pr-steward`. Limites: Organization settings > Usage (por serviço) e teto por repositório para o Steward. Requisitos: plano Team ou Enterprise, repositório no GitHub, usage credits ativos, sem ZDR, HIPAA ou CMEK. Enterprise elegível pode resgatar créditos promocionais (pelo menos US$ 10.000, prazo para resgatar até 15/01/2027).

**Como se usa no dia a dia:** o admin liga no claude.ai; o GitHub mostra o check `Claude Code Review` e comentários do bot; qualquer pessoa com escrita comenta `@claude review`, `@claude steward` ou `@claude stop`. Atenção: esses comentários **disparam gasto**.

**Onde entra no forge:** como módulo opcional, ele decide **quando** cada agente pode agir para não brigar com o PR em draft e com a prova por `head_sha`. Sem o módulo, o forge segue igual e não gasta nada com o Ship.

## Por que o ID é 121
`forge next-id` (que consulta todas as branches) devolveu 121 em 24/11/2026. Branch: o ambiente desta sessão exige `claude/project-thread-37y7ao`; a convenção `feat/121-claude-ship-como-modulo` (FLX-02, Regra 15) fica para o dono decidir no merge (ver `status.md`).

## Clarify
Sem rodada de perguntas antes de escrever: o pedido traz a fronteira ("mecânica não muda"), o mecanismo (extensão instalável) e a regra de repositório. As lacunas reais viraram `## Decisões para o Gate 1`.

## Descrição de negócio
Quem instala o módulo `ship` passa a usar os agentes do Claude Ship (Code Review, PR Steward e Claude Approvals) como **sinais e ajudantes** dentro da esteira que já existe: revisão extra no PR da feature, correção assistida no retrabalho e uma política de aprovação gerada a partir da Constituição. Quem não instala nada segue igual, sem custo e sem aviso. O ganho: menos retrabalho manual do Builder e uma segunda opinião de revisão, sem abrir mão do Gate humano nem da prova por `head_sha`.

## O que é o Claude Ship (fontes lidas em 24/11/2026)
Fontes: artigo `claude.com/resources/articles/introducing-claude-ship` (23/11/2026) e docs `code.claude.com/docs/en/{claude-ship,code-review-setup,approvals,approvals-setup,ship-command,pr-steward,code-review}`.

| Peça | Estado | O que faz | Fato que importa para o forge |
|---|---|---|---|
| Code Review | GA | revisão multiagente do PR, achados inline por severidade | check run `Claude Code Review` que **não falha**; disparo configurável (`Once after PR creation`, `After every push`, `Manual (@claude review)`); **pula draft** no automático; `@claude review` roda também em draft |
| PR Steward | beta | mantém o PR verde: CI falho, conflito, feedback de review | **faz push**; pula draft por padrão (o Owner pode incluir); começa ao marcar pronto ou no próximo push; para com `@claude stop` ou tirando o label `claude-pr-steward-watching`; **não aprova nem mescla** |
| Claude Approvals | beta | aplica política versionada (`.claude/approvals.yml`, branch padrão) | `mode: observe` ou `enforce`; check `Claude Approvals`; `@claude approvals rerun`; pode aprovar sozinho se a regra deixar |
| `/ship` (Claude Code local) | GA | push, abre PR, revisa e trabalha feedback | abre o PR **pronto, não draft**; não mescla, mas pode levar a merge com auto-merge ou modo auto; desliga por `skillOverrides` (`ship-local`, `steward`) em managed settings |

**Governança (pedido do dono):**
| Item | Fato | Consequência para quem adota |
|---|---|---|
| Licença e plano | serviço comercial da Anthropic; Claude Team e Enterprise; só GitHub | projeto em GitLab, Azure Repos ou plano Pro fica de fora |
| Custo | cobrado por token; Code Review e Approvals têm limite de gasto próprio; PR Steward **não tem limite até você definir** | definir limite mensal antes de ligar; observe mode também cobra |
| Dados | Code Review roda na infraestrutura da Anthropic, lê diff e código ao redor, não executa | indisponível com ZDR, HIPAA, CMEK; Code Review fora do BAA; Approvals fora de FERPA e Access Transparency. Retenção não consta nas páginas lidas: consultar DPA antes de ligar em cliente regulado |
| Segurança | GitHub App `claude` com permissões **mais amplas** que a revisão usa; Steward escreve na branch | o módulo nunca recebe esse acesso; quem liga é o admin no claude.ai |
| Módulo `forge-sdd-ship` | proprietário, privado, sem dados próprios além de `gh` local | só fala com o GitHub pelo `gh` do usuário, com permissão declarada no manifesto (109) |

## Problema
Ligar o Claude Ship direto num repositório que roda a esteira quebra três invariantes sem avisar:
1. **`/ship` abre PR pronto**, e a esteira exige PR em draft até a Conferência 2 (ADR-0022, ADR-0023, RN-7 da 106); o hook pr-guard (ADR-0027) e o job `pronto?` passam a brigar com ele.
2. **PR Steward faz push depois do parecer**: o `head_sha` muda e o parecer `conforme` deixa de valer (ADR-0023 §4). Sem coordenação, cada push do Steward gera nova Conferência 2: retrabalho e tokens em dobro.
3. **Claude Approvals pode aprovar sozinho**, e a Camada C é humana (FLX-06); a Constituição e as ADRs não podem ser aprovadas por agente (GOV-01).

Falta quem **traduza a esteira para a configuração do Ship** e quem **segure o Ship nas janelas certas**. Isso é inteligência de operação: módulo, não core (109).

## Para quem
Nathan e times em plano Team/Enterprise no GitHub que já rodam a esteira e querem revisão e retrabalho assistidos. Agentes da esteira continuam sem saber do módulo.

## Encaixe na esteira (o core não muda)
| Ponto da esteira (core, ADR-0023) | Evento v1 (109) | Papel do Ship | O que o módulo faz |
|---|---|---|---|
| Builder abre PR de feature em draft | `pr.aberto` | nada automático (draft é pulado) | registra o PR; opcional: comenta `@claude review` para ter achados **antes** da Conferência 2 (decisão G1) |
| Reviewer devolve (`changes_requested`) | `feature.estado_mudou` | Steward pode corrigir o feedback | se G2 = assistido: comenta `@claude steward` no draft; ao voltar a `review_pending`, comenta `@claude stop` |
| Conferência 2 `conforme` | `veredito.gravado` | nenhum push pode acontecer | garante `@claude stop` antes do `[Continuar]`; avisa no `doctor` se o Steward está observando |
| `[Continuar]`, PR sai do draft | `pr.pronto` | Code Review (`Once after PR creation`) e Approvals rodam | Approvals em `observe` por padrão; Code Review vira sinal extra, nunca substitui a Camada B |
| Push depois de pronto (CI falho) | (polling `gh`, sem evento novo) | Steward pode corrigir | o parecer fica obsoleto pela regra do core; o módulo mostra "nova Conferência 2 necessária" em `pendentes` |
| Merge humano | `pr.mesclado` | nada | fecha o acompanhamento do PR |

**Política de aprovação gerada da Constituição** (`forge ship politica`): o módulo gera uma **proposta** de `.claude/approvals.yml` em `.forge/modulos/ship/` (nunca grava na raiz do projeto); o dono revisa e mescla num PR próprio. Regras mínimas: `sdd/memory/constitution.md`, `sdd/spec/adr/**`, `.agents/rules/**` e `.github/**` sempre com `seats` humano e `claude: comments`; o resto conforme o modo de revisão humana da Constituição; sempre `mode: observe` na primeira versão.

**`/ship` local:** fora da esteira na v1 (decisão G3). O módulo traz o trecho de managed settings `skillOverrides: {"ship-local": "off"}` como recomendação documentada, e o `doctor` do módulo avisa se o `/ship` está ligado num projeto com `Modo de revisão humana: com-humano`.

## Migração (quem já usa o fluxo atual)
| Caso | Situação | Caminho |
|---|---|---|
| M0 | não quer o Ship | nada muda; `Módulos permitidos: nenhum` (RN-1 da 109) |
| M1 | quer ativar, sem Ship ligado no repositório | 1) forge com contrato de módulos v1 (109); 2) `forge modulo instalar --de <arquivo> --sha256 <hash>` e autorizar; 3) PR que muda `Módulos permitidos` para `ship` na base; 4) `forge ship init` grava `sdd/modulos/ship.json` em `observar`; 5) admin liga Code Review e Steward com início **Manual** e limite de gasto; 6) `forge ship politica` e PR da `.claude/approvals.yml` em observe |
| M2 | admin já ligou o Ship no repositório antes do módulo | risco hoje: Steward com início `On` faz push em PR pronto e invalida o parecer (o core já protege, mas gasta). `forge ship doctor` detecta pelo check run e pelo label e recomenda trocar para Manual |
| M3 | PRs em voo no momento da ativação | PR em draft: segue o fluxo novo a partir do próximo evento. PR pronto com parecer: o módulo comenta `@claude stop` se o Steward estiver observando; não roda review retroativa |
| M4 | quer subir de nível (observe para sinal, Approvals para enforce) | muda `sdd/modulos/ship.json` e, para enforce, remove `mode: observe` e torna o check obrigatório no GitHub; exige 2 semanas em observe sem divergência grave (critério, não trava) |
| M5 | rollback | tira `ship` de `Módulos permitidos` (desliga sem apagar); admin desliga no claude.ai; PR que remove `.claude/approvals.yml`; `forge modulo remover ship` |

## Regras de negócio
- **RN-1** Herda RN-1 a RN-9 da 109: sem módulo, byte a byte igual; módulo nunca derruba a esteira; sem canal de volta.
- **RN-2** O Ship **nunca substitui** Camada A, Camada B nem a Camada C humana; é sinal adicional.
- **RN-3** Nenhum push do Steward acontece entre `veredito.gravado conforme` e o `[Continuar]`; se acontecer, vale a regra do core (parecer obsoleto) e o módulo avisa.
- **RN-4** O módulo nunca abre PR, nunca marca pronto, nunca aprova e nunca mescla; só comenta comandos `@claude` declarados no manifesto.
- **RN-5** Política de aprovação nasce em `observe` e protege Constituição, ADRs, rules e workflows com aprovação humana.
- **RN-6** O módulo só funciona em VCS `github` (Constituição); outro VCS: `doctor` do módulo avisa e nada é chamado.
- **RN-7** Custo visível: `forge ship doctor` avisa quando o Steward está ligado sem limite de gasto conhecido (o dono confirma em `ship.json`, o módulo não lê o claude.ai).
- **RN-8** Testes sem rede: `gh` falso, eventos de `testdata/` (TST-03, TST-04).

## Escopo
**Dentro (módulo, repositório novo):** manifesto, `forge ship init|politica|doctor|status|rotina`, reação a eventos, gerador de política, docs de migração e de managed settings.
**Dentro (core):** **nada de código novo**. Só a linha `ship` na tabela "Outros candidatos a módulo" da 109 e esta spec.
**Fora:** criar o repositório (o dono cria); qualquer mudança na esteira; usar `/ship` dentro do Builder; ler configuração do claude.ai; GitLab e Azure; evento novo no contrato (só se G4 pedir).

## Dependências
109 implementada até a execução de módulos e emissão de eventos (109-05 e a ADR 109-12) e, para `pendentes`, a 107-13. Até lá, esta iniciativa é só spec e o repositório pode nascer com a esteira SDD própria (RN-10 da 109).

## Riscos
| Risco | Mitigação |
|---|---|
| Steward invalida parecer e gera revisões em laço | início Manual; `@claude stop` nos pontos certos; core já segura por `head_sha` |
| Custo sem teto (Steward sem limite) | aviso no `doctor`; passo obrigatório na migração M1 |
| Approvals aprova o que é de humano | política gerada protege caminhos sensíveis; `observe` primeiro |
| Docs do Ship mudam (beta) | módulo isola nomes de comando, label e checks em constantes; `doctor` confere o check pelo nome |
| Dado de cliente regulado na Anthropic | RN de adoção: indisponível com ZDR/HIPAA/CMEK; checklist de governança no README do módulo |
| Divergência de docs (Code Review em `Admin settings > Claude Code` versus `Organization settings > Claude Ship`) | doc do módulo cita os dois caminhos até estabilizar |

## Decisões para o Gate 1
- **G1. Code Review no draft, antes da Conferência 2?** (a) não, só depois do `[Continuar]` (recomendado na v1: menos custo, sem mudar a ordem); (b) sim, `@claude review` no `pr.aberto`, achados como insumo do Reviewer.
- **G2. PR Steward no retrabalho?** (a) desligado; (b) assistido no draft, entre `changes_requested` e `review_pending` (recomendado); (c) livre após pronto (mais barato em tempo, mais caro em Conferências).
- **G3. `/ship` local:** (a) fora da esteira e desligado por managed settings (recomendado); (b) permitido só sobre PR já pronto.
- **G4. Evento `pr.push` no contrato v1 (aditivo)?** (a) não, polling por `gh` no módulo (recomendado, core intacto); (b) sim, entra na 109.
- **G5. Nome do repositório:** `forge-sdd-ship` (recomendado, convenção `forge-sdd-<nome>`), privado, licença proprietária como os demais.

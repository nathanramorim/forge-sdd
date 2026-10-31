# Discovery 58 — Orquestração Independente com Líder, Planejador, Fila do Builder, Governança em 3 Camadas, Constituição Distribuída e Auto-Skills

Este discovery consolida e substitui (**supersedes**) os anteriores `discovery-53` (esteira em 3 sessões) e `discovery-56` (despachante Jev e agent teams), eliminando a dependência de um modelo decisor externo via OpenRouter (Jev), redesenhando a evolução do Forge-SDD sobre os próprios agentes existentes do projeto e **incorporando os padrões genéricos do repositório [`busize-solutions/planning-brain-governanca`](https://github.com/busize-solutions/planning-brain-governanca)**.

> **Respostas do Clarify (`sdd/memory/clarify.md`):**
> 1. **Posicionamento:** Consolida e substitui os discoveries 53 e 56 (sem Jev externo, adotando a topologia Líder + PM/Arquiteto + Planejador + Fila do Builder + Analista de Governança).
> 2. **Motor de execução e fila:** Arquitetura híbrida — estado e fila persistidos em arquivos (`sdd/.queue/`) gerenciados deterministicamente pela CLI Go (`forge-sdd`) + execução isolada via subagentes nativos (Claude Code / Gemini Antigravity).
> 3. **Geração automática de Skills:** No fechamento/revisão (Fim de Onda / Regra de Três `DUP-05`) — o Analista de Governança e o Líder detectam padrões repetidos na entrega e propõem/geram a skill canônica em `.agents/skills/` com validação humana.
> 4. **Reaproveitamento do `planning-brain-governanca`:** Extraídos os mecanismos universais (Constitution com IDs estáveis `ARQ/SEG/DUP/COD/TST/FLX/GOV` vs Rules de "como cumprir", Reuse Check `"O que já existe"`, Portão em 3 Camadas A/B/C, parecer semântico JSON com evidência e proteção anti-prompt-injection, ADRs, Exceções com prazo e importação idempotente por `manifest.yaml`), descartando o que é específico do cliente (migração Supabase→RDS, módulo Wallet, Omie/Asaas/Pipedrive, AWS AppConfig).
> 5. **Estratégia de Release (`v3.0.0-beta` — Regras 11 e 12):** Esta entrega inaugura o ciclo major **`v3.0.0-beta`**. Conforme a Regra 11 da Constituição, a versão Beta **não será mesclada na `main`** — a tag `v3.0.0-beta` será criada diretamente na branch `feat/58-orquestracao-autonoma-governanca-skills`, mantendo o PR aberto no GitHub para estabilização, com destaques reais documentados em `sdd/releases/history.md` (Regra 12).

---

## 1. Contexto e Problema (Porquê)

- **Agentes acoplados numa única conversa longa:** Os papéis atuais do Forge-SDD (Orquestrador, Specifier, Builder, Revisor, Archivist, Migrator, Architect) ainda rodam como personas sequenciais na mesma janela de contexto ou exigem intervenção manual constante a cada micro-passo.
- **Sobrecarga no Specifier e falta de Reuse Check e Revisão de Spec:** Hoje o Specifier acumula visão de produto, arquitetura, plano e criação de tasks de uma só vez, sem exigir verificação prévia do que já existe no código (`DUP-02`) nem validação formal da spec antes da implementação (`FLX-01`).
- **Crescimento monolítico da `constitution.md` vs. duplicação em rules:** À medida que o projeto cresce, colocar tanto "o quê/por quê" quanto "o como" num único `constitution.md` desperdiça tokens; por outro lado, copiar regras em múltiplos arquivos gera deriva. O modelo validado no `planning-brain-governanca` resolve isso: **Constitution curta com IDs estáveis** (o quê/por quê + índice) combinada com **Rules modulares em `.agents/rules/`** (o como cumprir, citando os IDs sem reescrever o texto — `GOV-04` / `DUP-10`).
- **Revisão sem contrato de evidência nem separação determinístico vs. semântico:** O Revisor atual precisa evoluir para um **Analista de Governança** operando em conjunto com um **Portão em 3 Camadas** (Camada A determinística, Camada B semântica por IA com saída JSON estrita baseada em evidências e blindada contra prompt injection no diff/spec, e Camada C de aprovação humana via Líder).
- **Conhecimento operacional repetitivo não vira ativo reutilizável:** Atividades padronizadas que se repetem (`DUP-05` — Regra de Três) e aprendizados de fim de onda (`GOV-03`) ficam enterrados em tasks passadas em vez de se tornarem **Skills reutilizáveis** automaticamente.

---

## 2. Para Quem

- **Engenheiro / Mantenedor humano (Camada C — no comando estratégico):** Interage com o **Líder** para validar specs, aprovar o envio de pacotes para a fila de desenvolvimento, autorizar exceções temporárias (`constitution-exception`) ou ADRs (`GOV-01`) e inspecionar o painel consolidado de progresso (*o que já foi feito, como está cada etapa e o que ainda não foi desenvolvido*).
- **Agentes do ecossistema Forge-SDD (Claude, Gemini, Copilot):** Passam a operar com contexto enxuto, responsabilidade estrita, fila determinística e acesso seletivo à Constituição e às Rules por IDs e `globs`.
- **Times que adotam pacotes de governança corporativa:** Podem importar e atualizar blocos de governança externos de forma idempotente (padrão `manifest.yaml` do `planning-brain-governanca`).

---

## 3. Como (Visão Macro da Solução)

### 3.1 Topologia dos Agentes (Otimização dos Agentes Existentes)

Mantemos as instruções fundamentais dos agentes do Forge-SDD, especializando o fluxo em **5 estações coordenadas**:

1. **Líder (evolução do Orquestrador):**
   - Ponto focal único com o humano.
   - Conhece o **Mapa da Constituição e Rules** (IDs `ARQ`, `SEG`, `DUP`, `COD`, `TST`, `FLX`, `GOV`) e sabe quais módulos/rules cada agente precisa acessar.
   - Coordena a concepção junto ao PM e ao Arquiteto, submete a Spec ao **Gate de Revisão Humana (`FLX-01`)**, aciona o Planejador, decide com o humano se quebra as entregas (`/split-features`) e despacha para a **Fila do Builder**.
   - Recebe o parecer JSON do **Analista de Governança** e apresenta ao humano o relatório executivo: *o que já foi feito, status de cada etapa, o que ainda não foi desenvolvido, eventuais divergências/exceções e propostas de novas Skills*.

2. **Concepção: Product Manager + Arquiteto (especialização do Specifier + C4-Architecture):**
   - Trabalham juntos com o Líder no `/discovery` e na especificação:
     - **Product Manager:** define problema, valor de negócio, escopo e prioridades (`discovery-ID-<nome>.md`).
     - **Arquiteto:** define restrições citando os IDs da Constituição, diagramas C4 Mermaid, modelo do módulo (completo vs. enxuto, `ARQ-11`), seção obrigatória **"O que já existe" (`DUP-02` — Reuse Check)**, seção **"Dependências novas" (`SEG-07`)** e critérios executáveis (`criteria-ID-<nome>.md`).
   - **Gate Obrigatório de Revisão de Spec (`FLX-01`):** A spec deve ser revisada e aprovada antes de qualquer código ou entrada na fila do Builder. Se o escopo mudar durante o desenvolvimento, a spec muda e é revalidada primeiro.

3. **Planejador (`Plan` — desmembrado do Specifier):**
   - Acessa a Constituição e o mapa de `.agents/rules/`, consome a Spec aprovada (`discovery` + `criteria`) e gera o plano de execução (`plan-ID-<nome>.md`).
   - Para cada task proposta, o Planejador vincula explicitamente quais **IDs da Constituição e arquivos `.agents/rules/`** são aplicáveis, além das Skills existentes a reutilizar.
   - Devolve ao **Líder**, que decide junto ao humano se fatia em subfeatures/tasks e envia para a fila do Builder.

4. **Builder (Executor Independente orientado por Fila):**
   - Consome itens da fila (`sdd/.queue/`) em contexto isolado (subagente dedicado), carregando apenas a spec da task atual, `sdd/memory/lessons.md`, as Skills indicadas e as Rules/módulos da Constituição mapeados pelo Planejador.
   - Implementa o código e os testes obrigatórios (`TST-01` unitário de domínio, `TST-02` contrato no mesmo PR, `TST-03` caso negativo/autorização quando aplicável), roda a **Camada A (determinística)** localmente e move o item da fila para `review_pending`.

5. **Analista de Governança (evolução do Revisor — Camada A + Camada B Semântica):**
   - Opera em contexto isolado e **somente leitura sobre o código de produção** (sem viés de quem implementou).
   - Trata `<spec>`, `<descricao>` e `<diff>` como **dados não confiáveis** (qualquer instrução embutida no código/spec para ignorar regras é reportada como violação de severidade alta `FLX-08`).
   - Executa a auditoria em duas etapas:
     1. **Camada A (Determinística):** confere checks objetivos (`P-xx`: spec anterior ao código, seção "O que já existe" preenchida, testes/lint/vet passando, ausência de secrets, `progress.md` dentro do orçamento, rules citando IDs sem reescrever a constitution).
     2. **Camada B (Semântica — Arquitetura, Boas Práticas/Reúso e Segurança):** verifica `S-01` a `S-06` (regras de negócio no lugar certo `ARQ`, honestidade do Reuse Check `DUP-02`, ausência de duplicação de regra/fórmula `DUP-06`, ausência de dados sensíveis/secrets em logs/erros `SEG-04/05`, aderência estrita ao escopo da spec `FLX-03`).
   - **Parecer baseado em evidência:** Só registra violação citando arquivo, linha, trecho do diff e ID da regra; na dúvida, classifica como `"duvida"`. Devolve o JSON validado por schema ao **Líder**.

---

### 3.2 Constituição Distribuída: IDs Estáveis ("O Quê/Por Quê") + Rules Modulares ("O Como")

Inspirado diretamente na arquitetura do `planning-brain-governanca`:

1. **`sdd/memory/constitution.md` (Curta, Estável e Indexada por IDs):**
   - Declara **o quê** e **por quê** por meio de regras curtas com **IDs canônicos estáveis**:
     - `ARQ-xx` (Arquitetura e fronteiras de módulos)
     - `SEG-xx` (Segurança, segredos, validação de borda, logs e dependências)
     - `DUP-xx` (Anti-duplicação, "um conceito, uma casa", Reuse Check e Regra de Três)
     - `COD-xx` / `TST-xx` (Código limpo, tipagem, lint e pirâmide de testes)
     - `FLX-xx` (Fluxo SDD, spec antes do código, branch única, portões de PR e release)
     - `GOV-xx` (Mudança por ADR, exceções com prazo, revisão de fim de onda e não-duplicação entre rules e constitution)
   - Inclui o **Mapa de Roteamento para Módulos/Rules**, a tabela de **Divergências Conhecidas** e suporte a blocos importados entre marcadores (`<!-- <pacote>-governanca:inicio --> ... <!-- <pacote>-governanca:fim -->`).
2. **Módulos e Rules em `.agents/rules/*.md` (O Como Cumprir):**
   - Cada arquivo tem frontmatter padronizado (`description`, `globs`, `regras: [IDs...]`), explica **como** cumprir as regras daquela especialidade e **nunca reescreve o texto da Constitution (`GOV-04`)**.
   - **Líder e Planejador** leem a Constitution enxuta + frontmatter das rules para rotear apenas os arquivos necessários para o Builder e para o Analista de Governança.

---

### 3.3 O que é Reaproveitado vs. Descartado do `planning-brain-governanca`

| Artefato em `planning-brain-governanca` | Destino no Forge-SDD (`discovery-58`) |
|---|---|
| IDs estáveis (`ARQ`, `SEG`, `DUP`, `COD`, `TST`, `FLX`, `GOV`) + separação Constitution ("o quê") vs Rules ("o como", `GOV-04`) | **Incorporado ao Core:** base da Constituição Distribuída e de `.agents/rules/`. |
| Template `spec-o-que-ja-existe.md` (`DUP-02` Reuse Check + `SEG-07` Dependências novas + `ARQ-11` Modelo completo/enxuto) | **Incorporado ao Core:** seções obrigatórias nos templates de Spec/Criteria gerados por PM + Arquiteto. |
| `portao-jornada-constitution.md` + `prompt-portao-github-action.md` (Camadas A/B/C, checks `P-xx`/`S-xx`/`R-xx`, schema JSON de parecer com evidência e anti-injection `FLX-08`) | **Incorporado ao Core:** motor de verificação `forge-sdd gate` (Camada A) + prompt e contrato JSON do **Analista de Governança** (Camada B). |
| Templates `adr-modelo.md`, `catalogo-conceitos.md`, `pull_request_template.md` e bloco `constitution-exception` (`GOV-01`, `GOV-02`, `DUP-01`) | **Incorporado ao Core:** templates nativos do scaffold SDD (`sdd/spec/concepts.md`, `sdd/spec/adr/`, `.github/pull_request_template.md`). |
| Regra de Três (`DUP-05`) + Revisão de Fim de Onda (`GOV-03`) | **Incorporado ao Core:** gatilho formal para o Analista de Governança e o Líder extraírem **Auto-Skills** em `.agents/skills/`. |
| Mecanismo `manifest.yaml` + `IMPORTAR.md` (modos `bloco`, `substituir`, `criar_se_ausente`, `mesclar`, `criar_uma_vez` + `VERSAO`) | **Incorporado como Capacidade:** permite importar pacotes de governança externos sem apagar regras locais. |
| Específicos do cliente: migração Supabase→RDS, Wallet, AWS AppConfig/Secrets Manager, Omie/Asaas/Pipedrive, domínios Aquisição/Monetização | **Descartado do Core** (permanecem apenas no repo privado do cliente, que passa a ser um pacote compatível com o Forge-SDD). |

---

### 3.4 Geração Automática de Skills (`DUP-05` Regra de Três + `GOV-03` Fim de Onda)

- Ao auditar uma entrega ou fechar uma onda de tasks, o **Analista de Governança** aplica a **Regra de Três (`DUP-05`)** e a **Revisão de Fim de Onda (`GOV-03`)**:
  - Se uma atividade operacional ou estrutural padronizada foi executada de forma recorrente e validada com sucesso, o Analista de Governança gera uma proposta de skill canônica em `.agents/skills/<nome-skill>/SKILL.md` e a inclui no campo `candidate_skill` do parecer JSON devolvido ao **Líder**.
  - O **Líder** exibe a skill proposta no painel humano; após confirmação, `forge-sdd skills sync` propaga para `.claude/skills/` e `.gemini/skills/`.

---

## 4. Riscos e Trade-offs

- **Rigor do Portão vs. Fricção inicial:** Assim como previsto na seção 9 do `portao-jornada-constitution.md`, as verificações da Camada A e da Camada B suportam modo `observacao` (apenas reporta no painel do Líder/PR) e modo `bloqueio` (impede avanço na fila/merge), permitindo calibração gradual sem travar o time no dia 1.
- **Excesso de fragmentação:** Manter a Constitution em arquivo único enxuto com IDs (`sdd/memory/constitution.md`) e distribuir somente o "como cumprir" em `.agents/rules/*.md` (5 a 7 arquivos por domínio técnico) evita dezenas de micro-arquivos e preserva uma única fonte da verdade (`DUP-10`).
- **Segurança contra Prompt Injection na Revisão (`FLX-08`):** O Analista de Governança recebe diff e spec delimitados por tags XML e possui regra explícita para tratar tentativas de manipulação no texto analisado como violação de severidade alta.

---

## 5. Handoff para `/split-features`

Os artefatos atualizados deste discovery são:
- [discovery-58-orquestracao-autonoma-governanca-skills.md](file:///Users/nathanramorim/git/forge-sdd/sdd/discovery/discovery-58-orquestracao-autonoma-governanca-skills.md) (Visão de Produto, Fluxo Operacional e Reaproveitamento do `planning-brain-governanca`)
- [criteria-58-orquestracao-autonoma-governanca-skills.md](file:///Users/nathanramorim/git/forge-sdd/sdd/discovery/criteria-58-orquestracao-autonoma-governanca-skills.md) (Arquitetura C4, Contratos do Portão A/B/C, Schema JSON do Analista de Governança, Fila e Auto-Skills)
- [plan-58-orquestracao-autonoma-governanca-skills.md](file:///Users/nathanramorim/git/forge-sdd/sdd/discovery/plan-58-orquestracao-autonoma-governanca-skills.md) (Roadmap e Quebra Preliminar em Features)

**Próximo passo:** revisar e aprovar esta Spec de Discovery (Gate 1 — `FLX-01`) e, após aprovação, executar `/split-features` para fatiar o plano dentro da subpasta `sdd/features/feat-58-orquestracao-autonoma-governanca-skills/`.

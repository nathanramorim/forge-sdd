# Fluxo de Desenvolvimento e Ciclo de Vida do Projeto (Forge-SDD)

Este guia estabelece os padrões e o ciclo de vida do desenvolvimento no ecossistema do **Forge-SDD**, ensinando aos engenheiros e aos agentes de IA as regras de uso de cada ferramenta e o momento exato de acioná-los.

---

## 🔄 Visão Geral do Ciclo de Vida

O ciclo de vida de qualquer evolução no projeto segue as seguintes etapas lineares:

```mermaid
flowchart TD
    A["💡 Ideia / Demanda"] --> B["🔎 Discovery\n/discovery"]
    B --> C["🔀 Split (Opcional)\n/split-features"]
    C --> D["🌿 Criar Branch\ngit checkout -b feat/..."]
    D --> E["📝 Especificação\n/nova-feature"]
    E --> F["🔨 Build & Implementação\n/proxima-feature"]
    F --> G["🧪 Teste & Revisão\n/revisar"]
    G --> H["🚀 PR & Handoff\ngh pr create"]
```

---

## 🗺️ Este é o documento fonte da verdade do fluxo (feat-01-06)

Este arquivo é a **única fonte da verdade** para a sequência macro de comandos do Forge-SDD. `CLAUDE.md`/`GEMINI.md`/`copilot-instructions.md` e os chatmodes Orquestrador **citam** os estágios abaixo — não reescrevem a sequência completa. Mapeamento entre a narrativa de produto (Problema → Proposta → Refinamento → Execução → Entrega) e os 7 passos técnicos deste documento:

| Estágio (narrativa) | Passo(s) técnico(s) | Comando |
|---|---|---|
| Problema | 1. Discovery | `/discovery` |
| Proposta | 2. Split (opcional) | `/split-features` |
| Refinamento | 3. Branch + Especificação | `/nova-feature` |
| Execução | 4. Implementação | `/proxima-feature` |
| Entrega | 5. Revisão + 6. PR/Handoff | `/revisar` → `gh pr create` |

O ciclo `READ-MIN → PLAN → ACT → WRITE → CLOSE` descrito em `CLAUDE.md` é o protocolo **por comando** (o que qualquer agente faz dentro de uma única invocação); a tabela acima é o pipeline **por feature** (a sequência de comandos ao longo de uma sessão completa). São dois eixos complementares, não descrições concorrentes do mesmo fluxo.

---

## 🔎 1. Discovery (`/discovery`)

**Quando usar:**
* Sempre que surgir uma ideia vaga, um bug complexo sem causa raiz definida, ou uma nova funcionalidade que precise de refinamento de produto e engenharia.
* **Objetivo:** Explorar os impactos técnicos, as dependências, propor abordagens de design e mapear os riscos antes de escrever qualquer código ou especificação formal.

**Comportamento:**
* O agente assume papéis sênior (Product Manager e Lead Architect) para questionar o escopo, definir restrições técnicas, sugerir arquitetura e gerar um plano preliminar.
* **Clarify:** antes de produzir os três artefatos, o agente avalia a demanda recebida contra a heurística em `sdd/memory/clarify.md`. Só pergunta ao usuário se detectar lacuna real (critério de aceitação ausente, escopo ambíguo, dependência externa não mencionada) — descrições já claras seguem direto.

---

## 🔀 2. Split de Features (`/split-features`)

**Quando usar:**
* Quando o plano gerado no Discovery ou a feature planejada for grande demais.

**Critérios de quebra (Desacoplamento e Arquitetura Evolutiva):**
1. **Bounded Contexts:** Se a feature abranger mais de um contexto delimitado (ex: infraestrutura de banco de dados, API backend, e telas de interface no mesmo passo), ela **deve** ser fatiada.
2. **Limite de Tarefas (Rule of 7):** Se o escopo planejado contiver **mais de 7 tarefas**, a feature é considerada grande demais para uma única branch.
3. **Sem Dependência Circular:** Ao quebrar, garanta que as sub-features não criem bloqueios ou dependências circulares entre si.
4. **Entrega Incremental e Testável:** Cada sub-feature resultante deve ser independentemente testável e funcional por si só (mesmo que com mocks), permitindo integrações incrementais.
5. **Ordem de Camadas:** A implementação deve fluir preferencialmente de baixo para cima na stack:
   $$\text{Infraestrutura / Banco de Dados} \longrightarrow \text{Domínio / Regras} \longrightarrow \text{Aplicação / APIs} \longrightarrow \text{Interface / Apresentação}$$

---

## 🌿 3. Registro e Criação de Branch (`/nova-feature`)

**Clarify:** antes do passo de branch abaixo, o agente avalia a descrição recebida (de `/nova-feature` ou `/novo-fix`) contra a heurística em `sdd/memory/clarify.md`. Só pergunta ao usuário se detectar lacuna real — caso contrário, segue direto para o passo 1.

**Regra Crítica de Sequência:**
> [!IMPORTANT]
> **A Branch DEVE ser criada ANTES da criação do arquivo de especificação.**
> A ordem exata a ser seguida pelo Orquestrador/Specifier é:
> 1. Executar a criação da branch local: `git checkout -b feat/nome-da-feature`
> 2. Criar o arquivo `sdd/features/feat-XX-nome-da-feature.md` com a especificação e as tarefas.
> 3. Adicionar a entrada correspondente em `sdd/features/index.md` e `sdd/memory/progress.md`.

---

## 🔨 4. Implementação (`/proxima-feature`)

**Quando usar:**
* Ao iniciar o desenvolvimento de fato de uma feature marcada como `todo` em `sdd/memory/progress.md`.

**Comportamento:**
* O Orquestrador muda o status da feature para ativo, delega as tarefas para o Builder e coordena a execução incremental e focada.

---

## 🧪 5. Revisão (`/revisar`)

**Quando usar:**
* **Sempre** antes de finalizar uma feature e criar o Pull Request.

**Verificações Obrigatórias do Revisor:**
* [ ] Os critérios de conclusão e as tarefas definidos na especificação foram totalmente atendidos.
* [ ] Os testes locais estão passando (`go test ./...` ou comando equivalente da stack).
* [ ] Linters e validadores de código rodam limpos (`golangci-lint run`).
* [ ] Arquivos de progresso e índice (`progress.md` e `index.md`) foram devidamente atualizados refletindo o status `done`.

---

## 🚀 6. Operações no GitHub (Uso do `gh` CLI)

Sempre que a tarefa envolver interação com o repositório remoto (GitHub) para submissão de código e gerenciamento de Pull Requests, **prefira utilizar a ferramenta CLI `gh` oficial**, em vez de comandos git puros ou URLs web manuais.

**Comandos Recomendados:**
* **Criar Pull Request:**
  ```bash
  gh pr create --title "feat(nome): descricao curta" --body "Correcao/Implementacao de ..."
  ```
* **Visualizar status do PR:**
  ```bash
  gh pr status
  ```
* **Realizar Merge (após aprovação/revisão):**
  ```bash
  gh pr merge --squash --delete-branch
  ```
* Caso a ferramenta `gh` não esteja instalada no ambiente, caia para o git puro e instrua o usuário a abrir o PR de forma amigável através do link gerado no terminal.

---

## 📊 7. Observabilidade (`forge-sdd report`)

Os comandos que mudam estado do ciclo SDD (`/discovery`, `/split-features`,
`/nova-feature`, `/novo-fix`, `/proxima-feature`, `/revisar`, `/archive`)
gravam telemetria determinística (`forge-sdd session record`) sempre que
`telemetry.enabled` está ativo em `sdd/.sddrc`. A qualquer momento, rode:

```bash
forge-sdd report
```

para ver, por feature/fix/discovery: tokens de entrada+saída, modelos
usados, número de sessões, duração total, e a idade medida do projeto
(da métrica mais antiga à mais recente). Comandos de leitura/diagnóstico
(`/status`, `/doctor`, `/constitution`, `/c4-architecture`,
`/upgrade-sdd`, `/tutorial`) não gravam telemetria — não produzem
entrega, ficariam fora do escopo do relatório.

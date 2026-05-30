# Plan — [NOME-DO-PROJETO]

Plano de execução faseado. Cada fase produz algo funcional e testável antes de avançar.
Referência: `sdd/sdd-[projeto].md`.

<!--
INSTRUÇÕES DE USO:
- Cada fase deve ter uma "entrega" clara e testável
- Não avance para a próxima fase sem a entrega anterior funcionando
- Fases podem ser paralelas se não houver dependência entre si
- Decisões abertas devem ser resolvidas ANTES de iniciar a fase que delas depende
- Marque cada task com [x] ao concluir
- Atualize o status em sdd/features/index.md e sdd/memory/progress.md
-->

---

## Fase 0 — Fundação (scaffolding + estado)

Objetivo: projeto rodando, estrutura de pastas criada, estado central funcionando, API/serviço conectado.

> **Já existem no repositório:** [listar o que não deve ser recriado]

- [ ] **0.1** Criar estrutura de pastas conforme SDD
- [ ] **0.2** Criar `[arquivo de dependências]` com dependências iniciais
- [ ] **0.3** Criar `[arquivo de variáveis de ambiente]` com:
  ```
  [VARIAVEL_1]=
  [VARIAVEL_2]=
  ```
- [ ] **0.4** Criar `[arquivo de configuração]` com seções:
  - `[secao-1]` — [descrição]
  - `[secao-2]` — [descrição]
- [ ] **0.5** Implementar `[módulo de estado]` — dataclass/struct/objeto central com todos os campos
- [ ] **0.6** Implementar `[utilitários básicos]` — funções de I/O, helpers
- [ ] **0.7** Smoke test: `[comando]` que conecta ao serviço externo e imprime resposta

**Entrega da fase:** `[comando de smoke test]` funciona sem erros.

---

## Fase 1 — [Nome do núcleo do sistema]

Objetivo: cada componente principal funciona isoladamente e pode ser testado diretamente.

- [ ] **1.1** Implementar `[componente-A]` com:
  - [responsabilidade 1]
  - [responsabilidade 2]
  - [responsabilidade 3]

- [ ] **1.2** Implementar `[componente-B]` — [descrição curta]

- [ ] **1.3** Implementar `[componente-C]` — [descrição curta]

- [ ] **1.4** Implementar `[componente-D]` — [descrição curta]

- [ ] **1.5** Implementar `[componente-E]` — [descrição curta]

**Entrega da fase:** cada componente testável com `[comando de teste unitário]` passando um input de exemplo.

---

## Fase 2 — [Integração / Camada de dados / Templates]

Objetivo: componentes integrados produzem output real a partir de dados mockados.

> **[Nota sobre assets ou recursos existentes que serão convertidos/usados]**

- [ ] **2.1** Implementar `[módulo de integração]` — [descrição]
- [ ] **2.2** Criar `[recurso-1]` — baseado em `[arquivo existente]`
  - [detalhe 1]
  - [detalhe 2]
- [ ] **2.3** Criar `[recurso-2]` — baseado em `[arquivo existente]`
- [ ] **2.4** Criar `[recurso-3]` — criado do zero
- [ ] **2.5** Teste de integração: `[comando]` com dados mock

**Entrega da fase:** `[função principal]` gera output válido com dados mockados.

---

## Fase 3 — [Pipelines / Fluxos / Workflows]

Objetivo: fluxo completo de ponta a ponta para cada caso de uso principal.

- [ ] **3.1** Implementar `[pipeline/fluxo-A]`:
  ```
  [passo-1] → [passo-2] → [passo-3]
  → [condição de retry ou fallback]
  → [saída final]
  ```

- [ ] **3.2** Implementar `[pipeline/fluxo-B]`:
  ```
  [passo-1] → [passo-2]
  → [condição de retry]
  ```

- [ ] **3.3** Implementar `[pipeline/fluxo-C]` (se houver variante paralela/assíncrona):
  ```
  [tarefa-A] + [tarefa-B] ← paralelo
  → [saída combinada]
  ```

- [ ] **3.4** Teste de fluxo: rodar `[pipeline-A].run(estado)` com input real

**Entrega da fase:** `[pipeline principal].run(estado)` gera output final sem erros.

---

## Fase 4 — Orquestrador e Entry Point

Objetivo: sistema utilizável com um único comando / ponto de entrada.

- [ ] **4.1** Implementar `[orquestrador/router]`:
  - Inicializa estado
  - Aciona `[classificador/detector]`
  - Roteia para fluxo correto
  - Tratamento de tipo/caso desconhecido com erro claro
  - Log de tempo total de execução

- [ ] **4.2** Implementar `[entry point — main.py / index.ts / cmd/main.go]` com:
  - `--[opção-1]` — [descrição]
  - `--[opção-2]` — [descrição]
  - `--[opção-3]` — [descrição]

- [ ] **4.3** Testar fluxo completo:
  ```bash
  [comando de execução completo com input real]
  ```

- [ ] **4.4** Testar casos edge: [tipo forçado, dry-run, entrada inválida etc.]

**Entrega da fase:** sistema completo funcional via [CLI/API/interface].

---

## Fase 5 — [Feature Especializada / Fluxo Secundário]

Objetivo: [descrição do objetivo desta fase especializada].

- [ ] **5.1** [tarefa de implementação]
- [ ] **5.2** [prompt / schema / configuração especializada]
- [ ] **5.3** [integração com entry point]
- [ ] **5.4** Testar com caso real:
  ```bash
  [comando de teste com arquivo/dado real]
  ```

**Entrega da fase:** `[feature]` funciona com [tipo de input real].

---

## Fase 6 — Robustez e Qualidade

Objetivo: sistema confiável para uso em produção.

- [ ] **6.1** Logging estruturado — arquivo `logs/run-{timestamp}.log` com: componentes acionados, métricas, tempo por etapa, caminho do output
- [ ] **6.2** Tratamento de erros externos — capturar erros específicos de [API/serviço], mensagens claras no terminal
- [ ] **6.3** Validação de dados de entrada — retentar com prompt de correção se parsing falhar
- [ ] **6.4** Serialização do estado para JSON — `estado.to_json()` para debug e auditoria
- [ ] **6.5** `README.md` — instalação, configuração, exemplos dos principais fluxos
- [ ] **6.6** Teste end-to-end com [N] casos reais e validação manual do output

**Entrega da fase:** sistema pronto para uso em produção.

---

## Decisões que precisam ser tomadas antes de cada fase

| Antes de | Decisão |
|----------|---------|
| Fase 2   | [decisão de design/arquitetura necessária] |
| Fase 3   | [schema de dados / contrato de API a definir] |
| Fase 5   | [escopo da feature especializada] |

---

## Ordem recomendada de implementação por sessão de trabalho

```
Sessão 1 → Fase 0 completa + início da Fase 1
Sessão 2 → Fase 1 completa
Sessão 3 → Fase 2 completa
Sessão 4 → Fase 3 completa
Sessão 5 → Fase 4 completa
Sessão 6 → Fase 5 + Fase 6
```

<!--
DICAS PARA ADAPTAR ESTE TEMPLATE:

1. NÚMERO DE FASES: adapte conforme a complexidade do projeto.
   Projetos simples podem ter 3-4 fases; projetos complexos podem ter 7+.

2. NOMEAÇÃO DAS FASES: use nomes que descrevam o que o sistema consegue fazer
   ao final de cada fase, não o que você vai implementar.
   Bom: "Sistema gera output válido a partir de dados mockados"
   Ruim: "Implementar templates"

3. ENTREGAS: toda fase deve ter um comando ou ação verificável.
   Se você não consegue testar se a fase está pronta, subdivida ou reformule.

4. DECISÕES ABERTAS: liste decisões bloqueantes ANTES de começar o plano,
   não durante a implementação. Resolva-as nas primeiras sessões.

5. PARALELISMO: fases sem dependência entre si podem ser implementadas
   em paralelo por diferentes agentes/sessões.
-->

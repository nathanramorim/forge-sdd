# Copilot Instructions — [NOME-DO-PROJETO]

## Contexto do projeto

[Descrição em 2-3 linhas: o que o sistema faz, como faz, para quem. Ex: "API REST em Node.js para gestão de pedidos de e-commerce. Execução via Docker, banco PostgreSQL, autenticação JWT."]

- SDD completo: `sdd/sdd-[projeto].md`
- Plano faseado: `sdd/plan-[projeto].md`
- Constituição (regras imutáveis): `sdd/memory/constitution.md`
- Estado atual do projeto: `sdd/memory/progress.md`
- Features e branches: `sdd/features/index.md`

---

## Regras obrigatórias

1. **Antes de começar qualquer tarefa**, leia `sdd/memory/progress.md` para entender onde o projeto está.
2. **Toda feature tem um arquivo próprio** em `sdd/features/feat-XX-*.md` — consulte antes de implementar.
3. **Config centralizado** — valores configuráveis vão em `[arquivo de config]`, nunca hardcoded no código.
4. **Branch por feature** — cada feature tem sua branch conforme `sdd/features/index.md`.
5. **Secrets em `.env`** — nunca commitar credenciais; usar `.env.example` como referência.
6. [Regra específica do projeto — ex: "Sem `console.log` — usar o logger configurado"]
7. [Regra específica do projeto — ex: "Migrations sempre via ferramenta X, nunca SQL direto"]

---

## Ao finalizar qualquer sessão ou feature

Obrigatório atualizar os seguintes arquivos antes de encerrar:

### 1. `sdd/memory/progress.md`
- Marcar features concluídas na tabela de status (`todo` → `in-progress` → `done`)
- Adicionar entrada em **Histórico de sessões** com data, o que foi feito e o que ficou pendente
- Atualizar **Próximo passo** com a próxima feature a ser iniciada

### 2. `sdd/features/feat-XX-*.md` (da feature trabalhada)
- Marcar tasks concluídas com `[x]`
- Mudar `Status:` no cabeçalho para `in-progress` ou `done`

### 3. `sdd/features/index.md`
- Atualizar coluna `Status` da feature na tabela de índice

---

## Stack e convenções de código

<!-- Preencha com as convenções do seu projeto. Exemplos abaixo: -->

- **Runtime:** [linguagem + versão] — ex: Node.js 22+, Python 3.11+, Go 1.22+
- **Gerenciador de pacotes:** [ex: pnpm, uv, go mod]
- **Testes:** [ex: pytest, jest, go test] — rodar com `[comando]`
- **Linter/formatter:** [ex: ruff, eslint, gofmt]
- **Tipagem:** type hints / TypeScript strict / interfaces em todas as funções públicas
- **Logging:** usar `[logger configurado]`, nunca `print()` / `console.log()` direto
- **Tratamento de erros:** capturar exceções específicas, nunca `catch (e) {}` vazio

<!-- Remova o que não se aplica e adicione o que for específico -->

---

## Estrutura de pastas esperada

<!-- Documente a estrutura final esperada após todas as features -->

```
[raiz]/
├── [pasta-src]/        ← código principal
├── [pasta-tests]/      ← testes
├── [pasta-config]/     ← arquivos de configuração
├── [arquivo-config]    ← ex: config.toml, .env.example
└── sdd/                ← documentação (não remover)
```

---

## Design system / UI (se aplicável)

<!-- Remova esta seção se o projeto não tiver interface visual -->

- Tipografia: [ex: Inter, Montserrat]
- Paleta: definida em `[arquivo]`
- Componentes: seguir padrões em `skills/design-system.md`
- Ícones: [ex: Lucide, Heroicons]

---

## Referência rápida dos agentes

| Agente | Quando invocar |
|--------|---------------|
| **[Orquestrador]** | Início de sessão, "próxima feature", "o que falta" |
| **[Builder Foundation]** | feat-00, scaffolding, config inicial |
| **[Builder Core]** | features do domínio principal do projeto |
| **[Revisor]** | "revisar", "validar", "checar implementação" |

<!-- Adicione/remova linhas conforme os agentes que você criou -->

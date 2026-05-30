# SDD — [NOME-DO-PROJETO]

<!--
SDD = Software Design Document
Este documento descreve O QUE o sistema faz e COMO está estruturado.
Escreva antes de qualquer código. Revise ao final de cada fase.
Use este template substituindo os [PLACEHOLDERS] e removendo comentários desnecessários.
-->

## Visão geral

[2-4 linhas descrevendo o sistema. Inclua: o que faz, como faz, para quem, tecnologia principal e modo de execução.]

Exemplo: "Sistema multiagente local em Python para geração automática de documentos HTML a partir de entradas textuais. Orquestrador classifica a entrada, aciona agentes especializados via API de LLM, e passa o resultado para templates Jinja2 — produzindo propostas comerciais, contratos e páginas prontas para envio."

## Stack

- Runtime: [ex: Python 3.11+, Node.js 22+, Go 1.22+]
- [Biblioteca principal]: [ex: Anthropic SDK, Express, GORM]
- [Segunda lib relevante]: [ex: Jinja2, Prisma, Redis client]
- [Modo async, se aplicável]: [ex: asyncio, Promise.all, goroutines]
- Config: [ex: TOML via tomllib, JSON, YAML via env vars]
- Entry point: [ex: main.py com argparse, index.ts, cmd/main.go]
- [Banco de dados ou "Sem banco de dados — estado em memória"]
- [Servidor web ou "Sem servidor web — execução local via CLI"]

> **Assets existentes** — o repositório já contém:
> - [listar arquivos/pastas que já existem e não devem ser recriados]
> - [ex: `assets/logo.svg`, `templates/*.html`, `scripts/legados/`]
>
> **Notas importantes:** [ex: diferença entre tipografia atual vs design system, convenções de naming existentes]

## Estrutura de pastas

<!--
Liste APENAS as pastas e arquivos relevantes para a arquitetura.
Não liste arquivos de config triviais (.gitignore, etc.) a menos que sejam importantes.
Use ← JÁ EXISTE para indicar o que não precisa ser criado.
-->

```
[nome-do-projeto]/
├── [entry-point]              # CLI / servidor / ponto de entrada
├── [config-file]              # configurações do projeto
├── [state-file]               # estado central (dataclass / store / struct)
├── [pasta-principal]/
│   ├── __init__.py
│   ├── [módulo-A].py          # [responsabilidade]
│   ├── [módulo-B].py          # [responsabilidade]
│   └── [módulo-C].py          # [responsabilidade]
├── [pasta-secundária]/
│   ├── __init__.py
│   └── [módulo].py            # [responsabilidade]
├── [pasta-assets]/            # ← JÁ EXISTE
│   └── [arquivo.ext]
├── [pasta-templates]/         # ← JÁ EXISTE (converter para Jinja2/EJS/etc)
│   └── [template.ext]
├── [pasta-prompts/schemas]/   # system prompts / schemas (versionados)
│   └── [arquivo.txt]
├── [pasta-output]/            # arquivos gerados (gitignored)
├── [pasta-testes]/
│   └── test_[módulo].py
├── .env.example
├── [arquivo-dependências]     # requirements.txt / package.json / go.mod
└── [pasta-docs]/
    └── sdd.md                 # este arquivo
```

## Módulos

<!--
Descreva cada módulo/componente com suas responsabilidades.
Use [ ] para tarefas ainda não implementadas e [x] para as concluídas.
-->

### ESTADO / STATE

- [ ] `[Estrutura de estado central]` com campos: [listar todos os campos relevantes]
- [ ] Serialização para [JSON/YAML] (debug / log de execuções)

### [COMPONENTE PRINCIPAL — ex: AGENTS / SERVICES / HANDLERS]

- [ ] `[Módulo base]`: [responsabilidade — ex: carrega configuração, executa chamada com retry, loga métricas]
- [ ] `[Módulo A]`: [responsabilidade — ex: classifica input → devolve tipo]
- [ ] `[Módulo B]`: [responsabilidade — ex: extrai dados estruturados → JSON]
- [ ] `[Módulo C]`: [responsabilidade — ex: gera conteúdo / draft]
- [ ] `[Módulo D]`: [responsabilidade — ex: valida resultado e aprova/reprova]

### [FLUXOS / PIPELINES / WORKFLOWS]

- [ ] `[fluxo-A].run(estado)`: [descrição do passo-a-passo e condições de retry]
- [ ] `[fluxo-B].run(estado)`: [descrição]
- [ ] `[fluxo-C].run(estado)`: [variante paralela/assíncrona, se houver]

### ORQUESTRADOR / ROUTER

- [ ] `router.run([input])`: instancia estado, aciona [classificador], roteia para fluxo correto
- [ ] Tratamento de tipo desconhecido: raise/throw erro com mensagem clara
- [ ] Log de tempo total de execução por fluxo

### TOOLS / UTILITÁRIOS

- [ ] `[função principal]([args])`: [descrição]
- [ ] `[função de I/O]([args])`: [descrição]
- [ ] `[função auxiliar]([args])`: [descrição]

### CLI / ENTRY POINT (`[main.py / index.ts / main.go]`)

- [ ] `--[opção-1] [valor]`: [descrição e comportamento]
- [ ] `--[opção-2] [valor]`: [descrição e comportamento]
- [ ] `--[opção-3]`: [flag booleana — descrição]
- [ ] `--[dry-run / verbose / debug]`: [comportamento de debug]

### CONFIG (`[config.toml / config.json / .env]`)

- [ ] Seção `[[secao-llm-ou-api]]`: [campos — ex: provider, model, max_tokens, temperature]
- [ ] Seção `[[secao-paths]]`: [campos — ex: templates_dir, output_dir, assets_dir]
- [ ] Seção `[[secao-retry]]`: [campos — ex: max_attempts, backoff_base]
- [ ] Seção `[[secao-brand-ou-app]]`: [campos de identidade/branding — ex: nome, logo, contato]

## Fluxos principais

<!--
Descreva os 2-3 fluxos mais importantes do sistema em pseudocódigo.
Isso serve como especificação para o desenvolvedor e para o agente.
-->

### Fluxo A — [Nome do fluxo principal]
```
[entry point] --[opção] [input]
  → [passo-1]: [descrição]
  → [passo-2]: [descrição]
    → [sub-passo]: [descrição]
    → [sub-passo]: [descrição]
  → [passo-3]: [descrição]
  → [saída]: [formato e destino]
```

### Fluxo B — [Nome do fluxo secundário]
```
[entry point] --[opção] [input] --[flag]
  → [passo-1]: [descrição]
  → [passo-2]: [descrição]
  → [saída]: [formato e destino]
```

## Decisões de design

<!--
Documente DECISÕES JÁ TOMADAS — não dúvidas abertas.
Dúvidas abertas ficam em "Pendências".
Use [x] para decisões confirmadas.
-->

- [x] [Decisão 1 — ex: "Formato do JSON do Extractor: schema fixo por doc_type → definido nos prompts"]
- [x] [Decisão 2 — ex: "Templates HTML: usar os existentes como base, convertendo para Jinja2"]
- [x] [Decisão 3 — ex: "Retry do Writer quando Reviewer reprova: passar notes no próximo prompt"]
- [x] [Decisão 4 — ex: "Config de marca vem de config.toml, não hardcoded no código"]
- [ ] [Decisão pendente — ex: "Incluir fluxo --extrair para páginas ou deixar para v2?"]

## Ambientes

| Env   | Credenciais          | Output / persistência |
|-------|---------------------|----------------------|
| local | `[VAR]` em `.env`   | `[pasta/]`           |
| [staging] | `[VAR]` em CI  | `[destino]`          |

## Dependências

```
[pacote-1>=versao]
[pacote-2>=versao]
[pacote-3>=versao]
[pacote-4>=versao]
```

## Pendências / decisões abertas

<!--
Liste APENAS o que ainda não foi decidido.
Mova para "Decisões de design" (com [x]) assim que resolver.
-->

- [ ] [Decisão técnica pendente — ex: "Definir campos obrigatórios do schema do extractor"]
- [ ] [Decisão de escopo — ex: "Confirmar se fluxo X está no escopo da v1 ou v2"]
- [ ] [Decisão de infraestrutura — ex: "Estratégia de logging: stdout ou arquivo?"]
- [ ] [Validação com usuário final — ex: "Testar com dados reais do cliente X"]
- [ ] [Decisão de tipografia/UI — ex: "Confirmar tipografia: Montserrat vs Inter nos templates"]

<!--
DICAS PARA ESCREVER UM BOM SDD:

1. ESCREVA ANTES DO CÓDIGO — o SDD define o contrato. Código implementa.
   Se você só consegue escrever o SDD depois de codar, o design não estava claro.

2. VISÃO GERAL EM 2-4 LINHAS — se você precisa de mais linhas para descrever
   o que o sistema faz, o design provavelmente está complexo demais.

3. MÓDULOS COM [ ] — toda responsabilidade do sistema deve aparecer aqui.
   Se algo está no código mas não no SDD, ou está no SDD mas não no código,
   o documento está desatualizado.

4. FLUXOS COMO PSEUDOCÓDIGO — o diagrama de fluxo mais útil é um pseudocódigo
   vertical de 5-10 linhas. Evite diagramas que ficam desatualizados.

5. DECISÕES vs PENDÊNCIAS — separe o que já foi decidido (imutável durante a fase)
   do que ainda está em aberto. Nunca deixe decisões implícitas.

6. ASSETS EXISTENTES — sempre documente o que já existe no repositório.
   Evita retrabalho e conflitos com arquivos legados.

7. ATUALIZAÇÃO — o SDD deve ser atualizado ao final de cada fase.
   Um SDD desatualizado é pior que nenhum SDD.
-->

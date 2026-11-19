# feat-99-02 — Seção Hub mode

## Modelo do módulo

**Enxuto** (ARQ-11): adiciona a seção `#hub-mode` dentro de `docs/wiki.html` (criado pela feat-99-01). Sem novo arquivo HTML. Sem backend.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Docs sobre hub | Qualquer repo | Não existe | Criar conteúdo novo |
| `docs/wiki.html` | `docs/` | Criado pela feat-99-01 | Estender: adicionar `<section id="hub-mode">` |
| Componente `.ai-prompt` | `docs/wiki.html` | Criado pela feat-99-01 | Reusar sem alteração |

## Conteúdo da seção

```
#hub-mode
├── O que é o Hub mode
│   "Um repo (forge-sdd) rastreia vários repos do mesmo time"
├── Passo a passo (3 passos)
│   1. forge init no repo principal (hub)
│   2. forge init --link <repo> em cada repo satélite
│   3. forge pipeline para ver o status de todos
└── Dois prompts para IA:
    - "Configure o forge-sdd como hub neste workspace"
    - "Liste os repos vinculados e o status de cada um"
```

Sidebar: adicionar item "Hub mode" no `<aside>` da wiki.

## Dependências novas

Nenhuma. Depende de feat-99-01 (wiki.html existente).

## Rules e IDs da Constituição aplicáveis

- **ARQ-11:** enxuto; nenhum arquivo Go.
- **COD-01:** seção isolada dentro do HTML; sem misturar com Get Started.
- **DUP-02:** tabela "O que já existe" preenchida.

## Critérios de Aceitação Executáveis

```bash
# 1. Seção hub-mode presente
grep -c 'id="hub-mode"' docs/wiki.html
# deve retornar >= 1

# 2. Prompts para IA na seção hub
grep -A 50 'id="hub-mode"' docs/wiki.html | grep -c 'ai-prompt'
# deve retornar >= 2

# 3. Sidebar atualizada com link para hub-mode
grep -c 'href="#hub-mode"' docs/wiki.html
# deve retornar >= 1

# 4. Nenhum arquivo Go alterado
git diff --name-only HEAD | grep -v '\.go$' | grep -v '^$'
# nenhuma linha deve terminar em .go
```

# Plano — 99-wiki-get-started

## Onda 1 (paralelo)

| Task | Descrição | Dep | Rules |
|---|---|---|---|
| `feat-99-01` | Shell da wiki + seção Get Started (Instalação, Jornada ideal, Exemplos práticos) + componente `.ai-prompt` com copy button | — | ARQ-11, COD-01 |
| `feat-99-02` | Seção Hub mode (como usar forge-sdd como hub de múltiplos repos, com prompts para IA) | — | ARQ-11, COD-01 |

## Onda 2

| Task | Descrição | Dep | Rules |
|---|---|---|---|
| `feat-99-03` | CTA "Get Started →" em `docs/index.html` apontando para `/wiki`; adicionar link de wiki no nav | feat-99-01 | COD-01 |

## Notas

- **feat-99-01** é o maior bloco: cria `docs/wiki.html` e `docs/wiki.css` do zero.
- **feat-99-02** adiciona a seção hub dentro do `wiki.html` criado pela feat-99-01 (pode rodar em paralelo se o Builder usar mock do shell).
- **feat-99-03** é pequena — edita 2–3 linhas no `docs/index.html` existente.
- i18n: cada componente de texto usa `data-pt="…" data-en="…"` e um script mínimo troca ao detectar o locale do browser — igual ao padrão da LP.
- O componente `.ai-prompt` tem: ícone de robô, label "Prompt para IA", texto pré-formatado e botão copiar (Clipboard API + fallback `execCommand`).

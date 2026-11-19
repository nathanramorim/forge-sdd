# feat-99-01 — Wiki shell + Get Started + componente ai-prompt

## Modelo do módulo

**Enxuto** (ARQ-11): sem backend. Cria `docs/wiki.html` e `docs/wiki.css`. Lê design system da LP publicada em forge-sdd.org (lp-forge-sdd v4.8) como referência visual; não copia arquivos do repo lp-forge-sdd.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Shell de wiki | `docs/` | Não existe | Criar `docs/wiki.html` do zero |
| Estilo wiki | `docs/` | Não existe | Criar `docs/wiki.css` com tokens da LP publicada |
| Design tokens | forge-sdd.org (lp-forge-sdd v4.8) | `--bg:#0d0f14`, `--accent:#4f8ef7`, `--ok:#3ecf8e`, fontes JetBrains Mono + Montserrat | Reusar tokens; copiar declaração `:root` como base |
| Componente copy | LP publicada | Pill de instalação com botão "Copiar" | Estender padrão para criar `.ai-prompt` com bloco de texto e botão |
| Nav/header | LP publicada | Nav com logo + links | Reusar estrutura do header; adicionar link "Wiki" |

## Estrutura da página

```
docs/wiki.html
├── <header>  nav com logo + link "← Voltar" para LP
├── <aside>   sidebar com seções: Get Started, Jornada, Exemplos, Hub mode
└── <main>
    ├── #get-started   Instalação (npm/npx + forge init)
    ├── #jornada       Jornada ideal de trabalho (3 passos: Discovery → Builder → Você)
    └── #exemplos      Exemplo prático fim a fim (1 feature real)
```

Componente `.ai-prompt`:
```html
<div class="ai-prompt">
  <span class="ai-prompt__label">🤖 Prompt para IA</span>
  <pre class="ai-prompt__code">…texto do prompt…</pre>
  <button class="ai-prompt__copy">Copiar</button>
</div>
```
JS: `navigator.clipboard.writeText()` com fallback `execCommand('copy')`.
i18n: atributo `data-pt` / `data-en` em cada string; script mínimo troca ao detectar `navigator.language`.

## Dependências novas

Nenhuma. CDN de fontes já usada na LP (Google Fonts: JetBrains Mono + Montserrat).

## Rules e IDs da Constituição aplicáveis

- **ARQ-11:** spec declara modelo enxuto; sem pacote Go novo.
- **COD-01:** responsabilidade única — wiki.html tem apenas conteúdo Get Started; wiki.css tem apenas estilos da wiki.
- **DUP-02:** tabela "O que já existe" preenchida.

## Critérios de Aceitação Executáveis

```bash
# 1. Arquivos criados
test -f docs/wiki.html && test -f docs/wiki.css
# deve ter exit 0

# 2. Seções obrigatórias presentes
grep -c 'id="get-started"\|id="jornada"\|id="exemplos"' docs/wiki.html
# deve retornar >= 3

# 3. Componente ai-prompt presente (mínimo 2 instâncias)
grep -c 'class="ai-prompt"' docs/wiki.html
# deve retornar >= 2

# 4. i18n declarado
grep -c 'data-pt\|data-en' docs/wiki.html
# deve retornar >= 5

# 5. Botão copiar com JS
grep -c 'clipboard\|execCommand' docs/wiki.html
# deve retornar >= 1

# 6. Nenhum arquivo Go alterado
git diff --name-only HEAD | grep -v '\.go$' | grep -v '^$'
# nenhuma linha deve terminar em .go
```

# feat-99-03 — CTA "Get Started" na landing page

## Modelo do módulo

**Enxuto** (ARQ-11): edita a LP publicada para adicionar CTA. A LP vive em `nathanramorim/lp-forge-sdd`; esta task edita apenas esse arquivo. Sem backend.

> **Dependência:** feat-99-01 precisa estar no ar (wiki.html acessível em `/wiki`) antes de publicar este CTA.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| CTA primário | LP publicada (forge-sdd.org), bloco hero | `<button>"Como Instalar & Usar"</button>` (CTA azul/accent) | Estender: manter "Como Instalar & Usar", adicionar "Get Started →" como novo CTA primário acima ou ao lado |
| Nav da LP | forge-sdd.org header | Links: As 3 Sessões, Pipeline, Instalação & Comandos… | Estender: adicionar item "Wiki" ou "Get Started" no nav |
| Rota `/wiki` | forge-sdd.org | Não existe | Depende de deploy da feat-99-01 |

## O que fazer

1. No bloco hero da LP, adicionar antes do CTA "Como Instalar & Usar":
   ```html
   <a class="btn solid" href="/wiki">Get Started →</a>
   ```
2. No nav, adicionar após "Instalação & Comandos":
   ```html
   <a href="/wiki">Wiki</a>
   ```

## Dependências novas

Nenhuma. Depende de feat-99-01 estar publicada.

## Rules e IDs da Constituição aplicáveis

- **ARQ-11:** enxuto; sem Go, sem bundler.
- **COD-01:** só o arquivo da LP é alterado; sem efeito colateral.
- **DUP-02:** tabela "O que já existe" preenchida.

## Critérios de Aceitação Executáveis

```bash
# (Executar no repo lp-forge-sdd ou no arquivo da LP após a migração)

# 1. CTA para /wiki presente no hero
grep -c 'href="/wiki"' index.html
# deve retornar >= 1

# 2. Texto "Get Started" presente
grep -c 'Get Started' index.html
# deve retornar >= 1

# 3. Link wiki no nav
grep -c 'Wiki\|/wiki' index.html
# deve retornar >= 2
```

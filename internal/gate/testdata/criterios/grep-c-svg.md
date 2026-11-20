# fixture — critério com grep -c num SVG de uma linha

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Logo | `docs/logo.svg` | SVG de uma linha | Reusar | `DUP-01` |

## Critérios de Aceitação Executáveis
```bash
test "$(grep -c '<path' docs/logo.svg)" -ge 3
# negativo: SVG sem path reprova
test "$(grep -c '<path' docs/vazio.svg)" -eq 0
```

# Fix 66.2 — Terminal Bonitinho com Scroll e Busca no Modal de Telemetria

## Feature Correlacionada
- **Feature ID:** `feat-66-telemetria-detalhada-inspecao-e-reconciliacao`
- **Discovery:** `sdd/discovery/66-telemetria-detalhada-inspecao-e-reconciliacao/`
- **Contexto:** Aprimoramento visual e funcional do painel de detalhes/saída da inspeção de atividades no Forge Studio (`#activityModal`).

## Contexto e Problema
O modal de inspeção de atividades (`#activityModal`) exibe a saída dos comandos e detalhes da operação em um elemento `<pre>` básico. Quando logs de terminal ou auditorias de governança contêm muitas linhas (como saídas de `go test -race ./...` ou relatórios de conformidade), o usuário necessita de:
1. **Aparência de terminal moderno e polido** ("bonitinho"): barra superior com botões de janela macOS (vermelho/amarelo/verde), indicador de status do terminal, tipografia nítida monospace e espaçamento refinado.
2. **Scroll suave e fluido**: altura adaptativa com scrollbar estilizada e suporte a auto-scroll para o final ou início.
3. **Busca interna em tempo real**: campo de busca no terminal com contador de ocorrências (ex: `1 de 5`), navegação entre ocorrências (botões Anterior/Próximo e atalho Enter / Shift+Enter), e destaque visual (`<mark>`) dos termos encontrados sem quebrar o layout.

## Critérios de Aceitação (Gherkin)

### Cenário 1: Aparência de Terminal Estilizado
- **Dado** que o usuário abre o modal de inspeção de uma atividade com logs ou detalhes
- **Quando** a seção de saída/logs for renderizada
- **Então** ela deve apresentar uma barra de cabeçalho estilo terminal com controles macOS (dots), título do shell e status
- **E** a área de saída deve ter scrollbar customizada e contraste visual otimizado.

### Cenário 2: Busca com Destaque e Navegação
- **Dado** que há conteúdo longo na saída do terminal
- **Quando** o usuário digita um termo no campo de busca do terminal
- **Então** todas as correspondências devem ser destacadas visualmente
- **E** o contador deve indicar a quantidade de ocorrências encontradas
- **E** o usuário pode navegar para frente e para trás entre os resultados com scroll automático até a linha correspondente.

### Cenário 3: Resiliência e Desempenho
- **Dado** que o campo de busca está vazio ou é limpo
- **Quando** a busca for resetada
- **Então** o texto original formatado deve ser restaurado imediatamente sem perda de quebras de linha ou caracteres especiais
- **E** o botão de copiar saída deve continuar copiando o texto integral sem tags HTML de marcação.

## Regras Constitucionais Aplicáveis
- `ARQ-03`, `ARQ-04`: Go 100% Stdlib, sem dependências externas.
- `SEG-01`: Sanitização estrita contra injeção de HTML/XSS ao destacar termos de busca.
- `COD-01`, `COD-02`: Clean code, legível e robusto.
- `TST-04`: Testes automatizados passando sem concorrência ou regressões.

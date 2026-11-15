# feat-66-03 — Interface do Studio: Modal Interativo de Inspeção Passo a Passo e Linhas Clicáveis (`internal/studio/assets/`)

**Branch:** `feat/66-telemetria-detalhada-inspecao-e-reconciliacao`
**Fase:** 66-03
**Tamanho:** `M`
**Depende de:** `feat-66-02`
**Status:** `ready`

## Contexto & Motivação

Atualmente, a tabela "Telemetria das Estações & Uso de Tools em Tempo Real" do Forge Studio exibe apenas quatro colunas estáticas (horário, estação, ferramenta e ação resumida). Não é possível clicar para ver detalhes como os argumentos enviados, caminho do arquivo afetado, comando completo executado, duração ou a saída capturada do terminal.

Esta subfeature transforma a tabela em um explorador interativo:
1. Linhas da tabela de telemetria tornam-se clicáveis com indicação visual (`cursor: pointer`, hover sutil, ícone de expansão).
2. Criação do modal/gaveta `#activityModal` integrado visualmente ao design system do Studio:
   - Cabeçalho com avatar/cor da estação executora, tool name e status chip (`running`, `success`, `error`).
   - Bloco de comando/ação com formatação de terminal monoespaçado e botão "Copiar".
   - Metadados: caminho do arquivo afetado, timestamps (início e término) e duração em milissegundos.
   - Bloco de saída (Output / Error) com suporte a rolagem e quebra segura de linhas.
3. Ergonomia e acessibilidade de teclado:
   - Fechamento via tecla `ESC` ou clique fora do modal.
   - Navegação sequencial entre eventos via teclas `ArrowLeft` (evento anterior) e `ArrowRight` (próximo evento).

## Modelo do módulo

- **Modelo:** `enxuto` em `internal/studio/assets/`.
- **Justificativa:** Desenvolvido puramente em HTML5, CSS3 moderno e JavaScript nativo sem dependências de frameworks frontend pesados (`ARQ-03`, `ARQ-04`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Tabela de Telemetria | `internal/studio/assets/index.html` | `#consoleLogs`, tabela de logs | `estender` | Adicionar atributos de acessibilidade, classes de clique e vinculação de evento. |
| Estilização do Studio | `internal/studio/assets/studio.css` | Variáveis CSS de cor, chips, badges | `estender` | Aproveitar paleta de cores das estações (`--agent-pm`, `--agent-builder`, etc.) para o modal. |
| Lógica de Renderização | `internal/studio/assets/studio.js` | `renderConsoleLogs()` | `estender` | Integrar abertura de modal ao clicar na linha e atalhos de teclado. |
| Modal de Inspeção | `internal/studio/assets/` | Inexistente para telemetria | `criar` | Criar `#activityModal` e funções `openActivityModal` / `closeActivityModal`. |

## Dependências novas

- **Nenhuma dependência externa nova.** (`SEG-07`, `ARQ-04`).

## Regras Constitucionais e IDs Aplicáveis

- **Rules:** `.agents/rules/sdd-design-system.md`, `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-03`, `ARQ-04`, `SEG-01`, `SEG-06`, `DUP-02`, `COD-01`, `COD-02`.

## Critérios de Aceite Gherkin

```gherkin
Cenário: Abertura de modal de detalhes ao clicar na linha de telemetria
  Dado que o Forge Studio está exibindo a tabela de telemetria com registros
  Quando o usuário clica em uma linha de log correspondente a um comando run_command
  Então o modal de inspeção deve ser aberto com animação suave
  E deve exibir a estação executora, o comando completo, a saída capturada e o tempo de execução

Cenário: Fechamento do modal via teclado ESC
  Dado que o modal de inspeção de telemetria está aberto na tela
  Quando o usuário pressiona a tecla ESC
  Então o modal deve ser fechado e o foco retornado à tabela de telemetria

Cenário: Navegação sequencial entre eventos via teclado
  Dado que o modal de inspeção está aberto exibindo o evento na posição 2
  Quando o usuário pressiona a tecla ArrowRight
  Então o modal deve atualizar seu conteúdo para exibir o evento na posição 3
```

## Plano de Implementação

1. Inserir a estrutura HTML do modal `#activityModal` em `internal/studio/assets/index.html`.
2. Adicionar as regras de estilo correspondentes em `internal/studio/assets/studio.css`.
3. Implementar a lógica de abertura, fechamento, cópia e navegação por teclado em `internal/studio/assets/studio.js`.
4. Validar carregamento correto e ausência de regressões com `internal/studio/assets_test.go`.

## Comandos de Verificação Determinísticos

```bash
go test -v -race ./internal/studio/... && go vet ./...
```

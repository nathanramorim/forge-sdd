# Guia Técnico: Piloto Automático e Autonomia no Forge-SDD

Este documento consolida as especificações e o guia prático para configurar o Forge-SDD em modo **Autônomo (Auto-Pilot / Hardness Engineer)**, permitindo que a IA execute e revise tarefas em lote por horas, sem necessidade de intervenção humana turno a turno.

---

## 1. Visão Geral do Fluxo Autônomo

O ciclo de desenvolvimento interativo padrão do Forge-SDD foi estruturado para manter o desenvolvedor no controle a cada etapa. Contudo, em cenários de desenvolvimento paralelo ou execução noturna (*overnight building*), o fluxo de parada pode ser opcionalmente desabilitado.

A arquitetura autônoma divide-se em duas camadas:
1. **No Editor / IDE (Antigravity):** Utilização do comando nativo `/goal` do motor da IA para loops contínuos de turnos de construção.
2. **Na CLI / Terminal (Claude / Copilot / Gemini):** Utilização de uma flag física de bypass (`.sdd-auto-pilot`) em conjunto com um script daemon de controle (`sdd-daemon.sh`).

---

## 2. A Flag de Bypass: `.sdd-auto-pilot`

Para que o agente Orquestrador saiba que não deve interromper o fluxo na fase `PLAN` aguardando confirmação (`PLAN: Reporte status + próximas tasks -> aguarde confirmação`), ele verifica a presença do arquivo `.sdd-auto-pilot` na raiz do projeto.

### Regra de Transição no Prompt:
> **Piloto Automático:** Se o arquivo `sdd/.sdd-auto-pilot` ou `.sdd-auto-pilot` estiver presente na raiz, o Orquestrador pula a validação do `PLAN` e assume aprovação tácita, iniciando imediatamente a fase `ACT` (Builder). Ao fechar a sessão, ele atualiza o `progress.md`, comita as alterações e cria o PR (`gh pr create --fill`) sem interrupção.

---

## 3. O Script Daemon: `sdd-daemon.sh`

Este é um exemplo prático de um script daemon em **Bash** que você pode criar localmente em seu repositório para orquestrar e rodar a fila de features contidas em `sdd/memory/progress.md`:

```bash
#!/bin/bash
# sdd/scripts/sdd-daemon.sh
# Daemon de execução contínua para o Forge-SDD (Auto-Pilot)

set -e

FLAG_FILE=".sdd-auto-pilot"
PROGRESS_FILE="sdd/memory/progress.md"

echo "🤖 Iniciando Daemon de Autonomia Forge-SDD..."

# Cria a flag de bypass temporária
touch "$FLAG_FILE"
trap 'rm -f "$FLAG_FILE"; echo "🤖 Daemon finalizado e flag de bypass limpa."' EXIT

while true; do
  echo "🔍 Lendo fila de tarefas..."
  
  # Identifica a primeira feature com status "todo" no progress.md
  NEXT_FEATURE=$(grep -m 1 -E "\| (feat-[a-zA-Z0-9_-]+) \|.*\| todo \|" "$PROGRESS_FILE" | awk -F '|' '{print $2}' | tr -d '[:space:]')
  
  if [ -z "$NEXT_FEATURE" ]; then
    echo "✓ Nenhuma feature pendente de execução ('todo'). Ciclo concluído!"
    break
  fi
  
  echo "🚀 Iniciando execução autônoma para: $NEXT_FEATURE"
  
  # Executa o comando de próxima feature invocando o agente orquestrador
  # Exemplo com Claude Code / CLI:
  # claude-code run "/proxima-feature"
  
  # Exemplo usando agy (Antigravity CLI):
  # agy run "/proxima-feature"
  
  # Fallback de segurança para evitar loops infinitos rápidos em caso de falha de comando
  sleep 5
done
```

---

## 4. Autonomia com Antigravity (`/goal`)

No Google Antigravity, a autonomia contínua de longa duração é integrada diretamente ao motor de execução através do comando `/goal`.

*   **Uso:** `/goal [descrição macro do objetivo]`
*   **Comportamento:** O motor AGY assume a fila de subtarefas criadas a partir do objetivo e executa sucessivas chamadas de ferramentas de leitura, escrita e execução de terminal sem perguntar ou interromper a sessão, contanto que o progresso não encontre bloqueios insolúveis.

---

## 5. Boas Práticas e Custos de Token

> [!WARNING]
> **Monitoramento de Custos:** Loops contínuos automáticos consomem uma quantidade significativa de tokens em turnos de rework. É recomendado definir um orçamento de limites de chamadas de API de segurança no daemon para evitar surpresas no faturamento de provedores de IA.

> [!IMPORTANT]
> **Budget de Rework:** Limite o rework automático de código do Builder a no máximo **3 turnos** em caso de falha de compilação ou testes. Se falhar repetidamente, marque a feature como `blocked` no progress e notifique o desenvolvedor.

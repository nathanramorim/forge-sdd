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

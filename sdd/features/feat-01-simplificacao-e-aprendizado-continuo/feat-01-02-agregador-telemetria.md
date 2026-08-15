# Feature 01-02 — Agregador Mínimo de Telemetria

Fecha o gap "grava mas nunca lê": hoje `sdd/.metrics/session-*.json` acumula um arquivo por sessão sem que nenhum comando agregue esses dados. Depende de feat-01-01 para que os dados agregados sejam confiáveis.

## Critérios de Aceitação Executáveis

1. `/status` ou `/doctor` (não criar comando novo) ganha uma seção que resume `sdd/.metrics/session-*.json`: contagem por `outcome` (`approved`/`rejected`/`blocked`) e frequência de comandos/fases.
2. A leitura é feita diretamente sobre os arquivos JSON existentes — sem exigir novo armazenamento ou índice intermediário.
3. Replicado nos três agentes (Claude, Gemini, Copilot) mantendo paridade de comportamento.

## Status: todo

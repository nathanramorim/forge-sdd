# Conjunto de avaliação (feat-104-11)

- `refs/ref-NN-<nome>/`: 10 features de referência pequenas. Cada uma tem `spec.md` (critério executável com linha `# negativo:`, validado pelas checagens do P-25) e `simulacao.json` (desfecho por modo usado pelo simulador).
- `exemplo/`: repositório de exemplo isolado (módulo Go próprio, fora do `go test ./...` da raiz). É o ponto de partida onde as referências seriam implementadas.

Uso: `forge eval run --mode sessions`, `forge eval run --mode chat` e `forge eval compare`. O resultado vai para `sdd/.metrics/eval-<modo>.json`.

Os números do simulador são valores fixados em `simulacao.json`, não medições: servem para validar o conjunto e a apuração em CI, sem modelo. A execução com modelo real fica para uma feature posterior e será só manual.

# Conjunto de avaliação (feat-104-11)

- `refs/ref-NN-<nome>/`: 10 features de referência pequenas. Cada uma tem `spec.md` (critério executável com linha `# negativo:`, validado pelas checagens do P-25) e `simulacao.json` (um único desfecho, aplicado igual aos dois modos).
- `exemplo/`: repositório de exemplo isolado (módulo Go próprio, fora do `go test ./...` da raiz). É o ponto de partida onde as referências seriam implementadas.

**Sem execução real, este conjunto não compara os desenhos.** Os números do simulador são valores de exemplo, idênticos nos dois modos: servem só para testar em CI, sem modelo, a validação do conjunto e a apuração. A execução com modelo real fica para uma feature posterior e será só manual.

Uso: `forge eval run --mode sessions` e `forge eval run --mode chat` gravam `sdd/.metrics/eval-<modo>.json` com o campo `origem` (`simulado` ou `real`). `forge eval compare` recusa comparar números simulados; `forge eval compare --permitir-simulado` imprime a tabela com o aviso no topo, só para conferir a apuração.

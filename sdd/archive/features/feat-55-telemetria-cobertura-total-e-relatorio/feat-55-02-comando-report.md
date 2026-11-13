# feat-55-02 — Comando `forge-sdd report`

**Branch:** `feat/55-telemetria-cobertura-total-e-relatorio`
**Fase:** 55-02
**Depende de:** feat-55-01
**Status:** `done`

## Objetivo

Comando determinístico que responde às perguntas que motivaram a
discovery-55: quantos tokens foram gastos por feature/fix/discovery,
quais modelos foram usados em cada um, quanto durou cada sessão, e há
quanto tempo (medido pela telemetria) o projeto está ativo.

## Critérios de Aceitação Executáveis

1. Novo arquivo `cmd/forge-sdd/report.go` com comando Cobra
   `forge-sdd report --dir <path>` (default `.`, mesmo padrão de
   `session record`/`doctor`).
2. Reaproveita `AggregateSessionMetrics`/leitura de arquivo já existente
   em `session.go` (não duplica parsing) e usa `ClassifySessionType`
   (feat-55-01) para agrupar por item (`feature`/`fix`/`discovery`) via
   o path completo em `Feature` (cada path distinto é uma linha, não só
   o tipo agregado).
3. Saída no terminal (sem gerar arquivo) contendo, por item: tokens de
   entrada+saída somados, lista de modelos únicos usados, número de
   sessões e duração total.
4. Ao final, imprime "idade medida do projeto": diferença entre o
   timestamp mais antigo e mais recente extraído do nome dos arquivos
   `session-<ISO8601>.json` (parse do nome do arquivo, não do conteúdo —
   evita depender de um campo de timestamp que não existe no schema
   atual).
5. Diretório sem métricas: mensagem clara ("Nenhuma métrica encontrada em
   sdd/.metrics") e exit 0 (não é erro).
6. Teste unitário com fixture mista (2+ discoveries, features e fixes,
   modelos diferentes, tokens diferentes) valida a soma e a listagem por
   item, e um caso de path não reconhecível (categoria `outro`) não
   quebra o comando.
7. `go build ./...` e `go test ./...` passam.

## Handoff

Comando novo, registrado em `main.go` junto dos demais (`session`,
`doctor`, etc.). Não altera `doctor` — ele continua com seu resumo
simples; `report` é o novo comando para detalhamento.
